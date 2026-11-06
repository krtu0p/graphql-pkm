package service

import (
	"encoding/json"
	"graphql-pkm/internal/ai"
	"graphql-pkm/internal/database"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAsyncEnrichmentAndOllamaCompatibleEndpoints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/chat/completions":
			var req struct {
				Messages []struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"messages"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			userPrompt := ""
			if len(req.Messages) > 0 {
				userPrompt = req.Messages[len(req.Messages)-1].Content
			}

			reply := "Concise generated summary."
			if strings.Contains(userPrompt, "JSON array") {
				reply = "```json\n[\"local-first\", \"sqlite\", \"ollama\"]\n```"
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"choices": []map[string]interface{}{
					{
						"message": map[string]string{
							"role":    "assistant",
							"content": reply,
						},
					},
				},
			})
		case "/v1/embeddings":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"embedding": []float32{0.6, 0.8, 0.0},
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	db, err := database.NewSQLiteDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("NewSQLiteDB: %v", err)
	}
	defer db.Close()

	aiClient := ai.NewClientWithConfig(ai.ClientConfig{
		Enabled:      true,
		BaseURL:      srv.URL + "/v1",
		Model:        "llama3.2",
		EmbedBaseURL: srv.URL + "/v1",
		EmbedModel:   "nomic-embed-text",
	})

	noteSvc := NewNoteService(db, aiClient)
	note, err := noteSvc.CreateNote("Local AI Note", "Testing async background enrichment", []string{"AI"})
	if err != nil {
		t.Fatalf("CreateNote: %v", err)
	}
	if note == nil || note.ID == "" {
		t.Fatalf("expected note to be created immediately")
	}

	// Wait up to 2s for async enrichment to populate summary and key concepts
	var enrichedSummary string
	var enrichedConcepts []string
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, err := noteSvc.GetNote(note.ID)
		if err != nil {
			t.Fatalf("GetNote: %v", err)
		}
		if got != nil && got.Summary != nil && len(got.KeyConcepts) > 0 {
			enrichedSummary = *got.Summary
			enrichedConcepts = got.KeyConcepts
			break
		}
		time.Sleep(25 * time.Millisecond)
	}

	if enrichedSummary != "Concise generated summary." {
		t.Fatalf("expected enriched summary, got %q", enrichedSummary)
	}
	if len(enrichedConcepts) != 3 || enrichedConcepts[0] != "local-first" {
		t.Fatalf("expected enriched concepts, got %v", enrichedConcepts)
	}

	cache := database.NewSQLiteEmbeddingsCache(db, "nomic-embed-text")
	aiSvc := NewAIService(aiClient, cache, noteSvc)
	aiSvc.SetSemanticThreshold(0.5)

	semResults, err := aiSvc.SemanticSearch("local")
	if err != nil {
		t.Fatalf("SemanticSearch: %v", err)
	}
	if len(semResults) != 1 || semResults[0].Note.ID != note.ID {
		t.Fatalf("expected semantic match, got %+v", semResults)
	}
}

func TestGracefulDegradationWhenAIUnreachable(t *testing.T) {
	db := database.NewMemoryDB()
	// Point AI client to an closed local port so requests fail
	aiClient := ai.NewClientWithConfig(ai.ClientConfig{
		Enabled:      true,
		BaseURL:      "http://127.0.0.1:1/v1",
		Model:        "llama3.2",
		EmbedBaseURL: "http://127.0.0.1:1/v1",
		EmbedModel:   "nomic-embed-text",
		Timeout:      200 * time.Millisecond,
	})

	noteSvc := NewNoteService(db, aiClient)
	note, err := noteSvc.CreateNote("Offline Note", "Works without Ollama running", []string{"offline"})
	if err != nil {
		t.Fatalf("CreateNote should succeed when AI is unreachable: %v", err)
	}

	sqliteDB, err := database.NewSQLiteDB(filepath.Join(t.TempDir(), "fallback.db"))
	if err != nil {
		t.Fatalf("NewSQLiteDB: %v", err)
	}
	defer sqliteDB.Close()

	cache := database.NewSQLiteEmbeddingsCache(sqliteDB, "nomic-embed-text")
	aiSvc := NewAIService(aiClient, cache, noteSvc)

	sem, err := aiSvc.SemanticSearch("Offline")
	if err != nil {
		t.Fatalf("SemanticSearch should fall back without error: %v", err)
	}
	if len(sem) != 1 || sem[0].Note.ID != note.ID || sem[0].MatchType != "keyword" {
		t.Fatalf("expected keyword fallback result, got %+v", sem)
	}

	smart, err := aiSvc.SmartSearch("Offline")
	if err != nil {
		t.Fatalf("SmartSearch should fall back without error: %v", err)
	}
	if len(smart.Results) != 1 || smart.Results[0].Note.ID != note.ID {
		t.Fatalf("expected keyword fallback in SmartSearch, got %+v", smart)
	}

	hybrid, err := aiSvc.HybridSearch("Offline")
	if err != nil {
		t.Fatalf("HybridSearch should succeed: %v", err)
	}
	if len(hybrid) != 1 || hybrid[0].Note.ID != note.ID {
		t.Fatalf("expected 1 deduplicated result in HybridSearch, got %+v", hybrid)
	}

	// Nil-check in CheckRelationship
	if _, err := aiSvc.CheckRelationship(nil, note); err == nil {
		t.Fatalf("expected error when noteA is nil")
	}
}
