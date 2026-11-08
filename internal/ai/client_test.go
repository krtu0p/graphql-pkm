package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newChatCaptureClient builds a client pointed at a stub chat server that
// captures the last user prompt sent to /chat/completions.
func newChatCaptureClient(t *testing.T) (*Client, func() (string, int)) {
	t.Helper()

	var lastPrompt string
	var lastMaxTokens int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openRouterChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		if len(req.Messages) > 0 {
			lastPrompt = req.Messages[len(req.Messages)-1].Content
		}
		lastMaxTokens = req.MaxTokens

		// Reply whatever the caller is expected to parse.
		reply := "Localized summary text."
		if strings.Contains(lastPrompt, "Respond in this exact JSON format") {
			reply = "{}"
		} else if strings.Contains(lastPrompt, "JSON array") {
			reply = "[]"
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
	}))
	t.Cleanup(srv.Close)

	client := NewClientWithConfig(ClientConfig{
		Enabled: true,
		BaseURL: srv.URL + "/v1",
		Model:   "test-model",
	})
	return client, func() (string, int) { return lastPrompt, lastMaxTokens }
}

const wantLanguageRule = "Write your answer in the language of the note contents"

func TestGenerateSummaryPromptLanguageRule(t *testing.T) {
	client, lastPrompt := newChatCaptureClient(t)

	summary, err := client.GenerateSummary("Sourdough Schedule", "Feed the starter at 8am.")
	if err != nil {
		t.Fatalf("GenerateSummary: %v", err)
	}
	if summary != "Localized summary text." {
		t.Fatalf("unexpected summary from stub: %q", summary)
	}

	prompt, _ := lastPrompt()
	if !strings.Contains(prompt, wantLanguageRule) || !strings.Contains(prompt, "Do not translate or mix languages.") {
		t.Fatalf("summary prompt is missing the language rule:\n%s", prompt)
	}
	if !strings.Contains(prompt, "Summarize this note in 1-2 sentences") {
		t.Fatalf("summary prompt lost its original instruction:\n%s", prompt)
	}
}

func TestExtractKeyConceptsPromptLanguageRule(t *testing.T) {
	client, lastPrompt := newChatCaptureClient(t)

	concepts, err := client.ExtractKeyConcepts("Sourdough Schedule", "Feed the starter at 8am.")
	if err != nil {
		t.Fatalf("ExtractKeyConcepts: %v", err)
	}
	if len(concepts) != 0 {
		t.Fatalf("expected no concepts from stub, got %v", concepts)
	}

	prompt, _ := lastPrompt()
	if !strings.Contains(prompt, wantLanguageRule) || !strings.Contains(prompt, "Do not translate or mix languages.") {
		t.Fatalf("key concepts prompt is missing the language rule:\n%s", prompt)
	}
	if !strings.Contains(prompt, `["concept1", "concept2", "concept3"]`) {
		t.Fatalf("key concepts prompt lost its JSON array format:\n%s", prompt)
	}
}

func TestSmartSearchPromptLanguageRuleAndJSONKeys(t *testing.T) {
	client, lastPrompt := newChatCaptureClient(t)

	result, err := client.SmartSearch("baking", map[string]string{"note-1": "Sourdough: bake at 9pm"})
	if err != nil {
		t.Fatalf("SmartSearch: %v", err)
	}
	if result.Explanation != "" {
		t.Fatalf("expected empty explanation from stub, got %+v", result)
	}

	prompt, _ := lastPrompt()
	if !strings.Contains(prompt, "in the language of the note contents") || !strings.Contains(prompt, "Keep all JSON keys exactly as shown") {
		t.Fatalf("smart search prompt is missing the language rule:\n%s", prompt)
	}

	// JSON keys must stay English because the code parses them.
	for _, key := range []string{`"relevantNotes"`, `"noteId"`, `"reason"`, `"score"`, `"explanation"`, `"gaps"`, `"connections"`} {
		if !strings.Contains(prompt, key) {
			t.Fatalf("smart search prompt lost JSON key %s:\n%s", key, prompt)
		}
	}
}
