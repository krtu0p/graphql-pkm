package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"graphql-pkm/internal/ai"
	"graphql-pkm/internal/database"
	"graphql-pkm/internal/models"
)

func testAISettingsDefaults() AISettingsDefaults {
	return AISettingsDefaults{
		Enabled:      true,
		Provider:     "ollama",
		BaseURL:      "http://env.example/v1",
		APIKey:       "env-key",
		Model:        "env-model",
		EmbedBaseURL: "http://env.example/v1",
		EmbedAPIKey:  "env-embed-key",
		EmbedModel:   "env-embed-model",
	}
}

func testSaveAISettingsInput(apiKey *string) models.SaveAISettingsInput {
	return models.SaveAISettingsInput{
		Enabled:      true,
		Provider:     "ollama",
		BaseURL:      "http://saved.example/v1",
		Model:        "saved-model",
		APIKey:       apiKey,
		EmbedBaseURL: "http://saved.example/v1",
		EmbedModel:   "saved-embed-model",
	}
}

func TestAISettingsAPIKeyKeepClearReplace(t *testing.T) {
	defaults := testAISettingsDefaults()
	store := database.NewMemoryDB()
	client := ai.NewClientWithConfig(ai.ClientConfig{
		Enabled:      defaults.Enabled,
		BaseURL:      defaults.BaseURL,
		APIKey:       defaults.APIKey,
		Model:        defaults.Model,
		EmbedBaseURL: defaults.EmbedBaseURL,
		EmbedAPIKey:  defaults.EmbedAPIKey,
		EmbedModel:   defaults.EmbedModel,
	})
	svc := NewAISettingsService(store, client, defaults)

	settings, err := svc.GetAISettings()
	if err != nil {
		t.Fatalf("GetAISettings: %v", err)
	}
	if !settings.HasAPIKey {
		t.Fatalf("expected environment API key to be reported")
	}
	resolved, err := svc.resolve()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if resolved.APIKey != "env-key" {
		t.Fatalf("expected env key before save, got %q", resolved.APIKey)
	}

	replace := "saved-key"
	if _, err := svc.SaveAISettings(testSaveAISettingsInput(&replace)); err != nil {
		t.Fatalf("SaveAISettings replace: %v", err)
	}
	resolved, err = svc.resolve()
	if err != nil {
		t.Fatalf("resolve after replace: %v", err)
	}
	if resolved.APIKey != "saved-key" {
		t.Fatalf("expected replaced key, got %q", resolved.APIKey)
	}
	if resolved.BaseURL != "http://saved.example/v1" || resolved.Model != "saved-model" {
		t.Fatalf("saved settings were not applied: %+v", resolved)
	}
	if !client.IsEnabled() {
		t.Fatalf("expected shared client to be enabled after reload")
	}

	if _, err := svc.SaveAISettings(testSaveAISettingsInput(nil)); err != nil {
		t.Fatalf("SaveAISettings keep: %v", err)
	}
	resolved, err = svc.resolve()
	if err != nil {
		t.Fatalf("resolve after keep: %v", err)
	}
	if resolved.APIKey != "saved-key" {
		t.Fatalf("omitted API key should be kept, got %q", resolved.APIKey)
	}
	if value, found, err := store.GetAppSetting(settingAIAPIKey); err != nil || !found || value != "saved-key" {
		t.Fatalf("expected stored key to remain, found=%v value=%q err=%v", found, value, err)
	}

	clear := ""
	if _, err := svc.SaveAISettings(testSaveAISettingsInput(&clear)); err != nil {
		t.Fatalf("SaveAISettings clear: %v", err)
	}
	resolved, err = svc.resolve()
	if err != nil {
		t.Fatalf("resolve after clear: %v", err)
	}
	if resolved.APIKey != "" {
		t.Fatalf("empty API key should clear the saved key, got %q", resolved.APIKey)
	}
	settings, err = svc.GetAISettings()
	if err != nil {
		t.Fatalf("GetAISettings after clear: %v", err)
	}
	if settings.HasAPIKey {
		t.Fatalf("expected HasAPIKey false after clear")
	}
	if value, found, err := store.GetAppSetting(settingAIAPIKey); err != nil || !found || value != "" {
		t.Fatalf("expected stored key to be empty, found=%v value=%q err=%v", found, value, err)
	}
}

func TestAISettingsRuntimeReloadSharedByServices(t *testing.T) {
	var chatCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/chat/completions":
			atomic.AddInt32(&chatCalls, 1)
			var req struct {
				Messages []struct {
					Content string `json:"content"`
				} `json:"messages"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			prompt := ""
			if len(req.Messages) > 0 {
				prompt = req.Messages[len(req.Messages)-1].Content
			}

			reply := "Runtime reply."
			if strings.Contains(prompt, "Summarize this note") {
				reply = "Runtime summary."
			}
			if strings.Contains(prompt, "JSON array") {
				reply = `["runtime", "reload"]`
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"choices": []map[string]interface{}{
					{"message": map[string]string{"role": "assistant", "content": reply}},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	db, err := database.NewSQLiteDB(filepath.Join(t.TempDir(), "runtime.db"))
	if err != nil {
		t.Fatalf("NewSQLiteDB: %v", err)
	}
	defer db.Close()

	client := ai.NewClientWithConfig(ai.ClientConfig{
		Enabled:    false,
		Model:      "initial-model",
		EmbedModel: "nomic-embed-text",
	})
	settings := NewAISettingsService(db, client, AISettingsDefaults{
		Enabled:      false,
		Provider:     "ollama",
		Model:        "initial-model",
		EmbedBaseURL: "",
		EmbedModel:   "nomic-embed-text",
	})
	noteSvc := NewNoteService(db, client)
	aiSvc := NewAIService(client, database.NewSQLiteEmbeddingsCache(db, "nomic-embed-text"), noteSvc)

	if _, err := aiSvc.AskAI("hello"); err == nil {
		t.Fatalf("expected AI to be disabled before reload")
	}

	_, err = settings.SaveAISettings(models.SaveAISettingsInput{
		Enabled:      true,
		Provider:     "ollama",
		BaseURL:      srv.URL + "/v1",
		Model:        "runtime-model",
		EmbedBaseURL: srv.URL + "/v1",
		EmbedModel:   "nomic-embed-text",
	})
	if err != nil {
		t.Fatalf("SaveAISettings: %v", err)
	}

	reply, err := aiSvc.AskAI("hello")
	if err != nil {
		t.Fatalf("AIService should use the reloaded shared client: %v", err)
	}
	if reply != "Runtime reply." {
		t.Fatalf("unexpected AI reply %q", reply)
	}

	note, err := noteSvc.CreateNote("Runtime Enrichment", "worker should use new settings", []string{"runtime"})
	if err != nil {
		t.Fatalf("CreateNote: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, err := noteSvc.GetNote(note.ID)
		if err != nil {
			t.Fatalf("GetNote: %v", err)
		}
		if got != nil && got.Summary != nil && len(got.KeyConcepts) > 0 {
			if *got.Summary != "Runtime summary." {
				t.Fatalf("unexpected summary %q", *got.Summary)
			}
			if atomic.LoadInt32(&chatCalls) == 0 {
				t.Fatalf("expected chat completions to reach the reloaded endpoint")
			}
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for enrichment from the reloaded shared client")
}

func TestListAIModelsAgainstServer(t *testing.T) {
	var lastAuth atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		lastAuth.Store(r.Header.Get("Authorization"))
		switch r.URL.Path {
		case "/v1/models":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]string{
					{"id": "qwen2.5"},
					{"id": "llama3.2"},
				},
			})
		case "/broken/v1/models":
			http.Error(w, "provider unavailable", http.StatusServiceUnavailable)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	defaults := testAISettingsDefaults()
	svc := NewAISettingsService(database.NewMemoryDB(), ai.NewClientWithConfig(ai.ClientConfig{}), defaults)

	key := "unsaved-key"
	result, err := svc.ListAIModels(models.ListAIModelsInput{
		BaseURL: srv.URL + "/v1",
		APIKey:  &key,
	})
	if err != nil {
		t.Fatalf("ListAIModels: %v", err)
	}
	if result.Error != nil {
		t.Fatalf("unexpected error result: %s", *result.Error)
	}
	if len(result.Models) != 2 || result.Models[0] != "llama3.2" || result.Models[1] != "qwen2.5" {
		t.Fatalf("unexpected models: %v", result.Models)
	}
	auth, _ := lastAuth.Load().(string)
	if auth != "Bearer unsaved-key" {
		t.Fatalf("expected unsaved key to be used, got %q", auth)
	}

	result, err = svc.ListAIModels(models.ListAIModelsInput{BaseURL: srv.URL + "/v1"})
	if err != nil {
		t.Fatalf("ListAIModels with effective key: %v", err)
	}
	if result.Error != nil {
		t.Fatalf("unexpected error result: %s", *result.Error)
	}
	auth, _ = lastAuth.Load().(string)
	if auth != "Bearer env-key" {
		t.Fatalf("expected effective key to be used, got %q", auth)
	}

	result, err = svc.ListAIModels(models.ListAIModelsInput{BaseURL: srv.URL + "/broken/v1"})
	if err != nil {
		t.Fatalf("ListAIModels broken: %v", err)
	}
	if result.Error == nil {
		t.Fatalf("expected a readable error for provider failure")
	}
	if !strings.Contains(*result.Error, "503") {
		t.Fatalf("expected status in error, got %q", *result.Error)
	}
}
