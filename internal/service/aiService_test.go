package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"graphql-pkm/internal/ai"
	"graphql-pkm/internal/models"
)

func TestCheckRelationshipPromptRules(t *testing.T) {
	noteA := &models.Note{
		ID:      "a",
		Title:   "Sourdough Baking Schedule",
		Content: "Feed the starter at 8am, shape at 6pm, bake at 9pm.",
	}
	noteB := &models.Note{
		ID:      "b",
		Title:   "Emergency Fund Plan",
		Content: "Save six months of expenses in a separate account.",
	}

	prompt := buildCheckRelationshipPrompt(noteA, noteB)

	if !strings.Contains(prompt, "Write your answer in the language of the note contents (the body text, not the titles and not these instructions). If the two notes are in different languages, use the language of Note A. Do not translate or mix languages.") {
		t.Fatalf("prompt is missing the language rule:\n%s", prompt)
	}
	if !strings.Contains(prompt, "same concrete topic") || !strings.Contains(prompt, "Abstract similarities do not count") {
		t.Fatalf("prompt is missing the strictness rule:\n%s", prompt)
	}
	if !strings.Contains(prompt, "the first line of your answer must be exactly NO_RELATIONSHIP") {
		t.Fatalf("prompt is missing the NO_RELATIONSHIP rule:\n%s", prompt)
	}
}

func TestStripNoRelationshipLine(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "marker with trailing newline and sentence",
			input: "NO_RELATIONSHIP\nNote A is a baking schedule; note B is a savings plan.",
			want:  "Note A is a baking schedule; note B is a savings plan.",
		},
		{
			name:  "marker with leading whitespace and lowercase variant",
			input: "   no_relationship\r\nThe first note plans bread baking; the second plans savings.",
			want:  "The first note plans bread baking; the second plans savings.",
		},
		{
			name:  "marker preceded by blank lines",
			input: "\n\n No_Relationship \n\nThe first note plans bread baking; the second plans savings.",
			want:  "The first note plans bread baking; the second plans savings.",
		},
		{
			name:  "marker only",
			input: "\n  NO_RELATIONSHIP  \n",
			want:  "",
		},
		{
			name:  "normal relationship answer passes through unchanged",
			input: "Both notes are directly about planning a bread baking routine.\nThe starter feeding schedule depends on the grocery list note.",
			want:  "Both notes are directly about planning a bread baking routine.\nThe starter feeding schedule depends on the grocery list note.",
		},
		{
			name:  "marker quoted by the model later in the text is kept",
			input: "No relationship between them, the token NO_RELATIONSHIP was not used correctly.",
			want:  "No relationship between them, the token NO_RELATIONSHIP was not used correctly.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripNoRelationshipLine(tc.input); got != tc.want {
				t.Fatalf("stripNoRelationshipLine(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestCheckRelationshipWithStubServer(t *testing.T) {
	reply := "NO_RELATIONSHIP\nNote A is a sourdough baking schedule; note B is an emergency fund plan."
	capturedMaxTokens := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}
		var req struct {
			MaxTokens int `json:"max_tokens"`
			Messages  []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		capturedMaxTokens = req.MaxTokens

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
	defer srv.Close()

	aiClient := ai.NewClientWithConfig(ai.ClientConfig{
		Enabled: true,
		BaseURL: srv.URL + "/v1",
		Model:   "test-model",
	})
	aiSvc := NewAIService(aiClient, nil, nil)

	noteA := &models.Note{
		ID:      "a",
		Title:   "Sourdough Baking Schedule",
		Content: "Feed starter 8am, shape 6pm, bake 9pm.",
	}
	noteB := &models.Note{
		ID:      "b",
		Title:   "Emergency Fund Plan",
		Content: "Save six months of expenses.",
	}

	got, err := aiSvc.CheckRelationship(noteA, noteB)
	if err != nil {
		t.Fatalf("CheckRelationship: %v", err)
	}
	if capturedMaxTokens != relationshipMaxTokens {
		t.Fatalf("expected request max_tokens %d for hidden reasoning, got %d", relationshipMaxTokens, capturedMaxTokens)
	}
	if strings.Contains(strings.ToUpper(got), "NO_RELATIONSHIP") {
		t.Fatalf("token leaked to the caller: %q", got)
	}
	want := "Note A is a sourdough baking schedule; note B is an emergency fund plan."
	if got != want {
		t.Fatalf("CheckRelationship = %q, want %q", got, want)
	}

	// Relationship answers pass through without the token being touched.
	reply = "Both notes are about planning a sourdough baking routine."
	if _, err := aiSvc.CheckRelationship(noteA, noteB); err != nil {
		t.Fatalf("CheckRelationship: %v", err)
	}

	// A bare marker without any description falls back to the English default.
	reply = "\n  NO_RELATIONSHIP  \n"
	got, err = aiSvc.CheckRelationship(noteA, noteB)
	if err != nil {
		t.Fatalf("CheckRelationship: %v", err)
	}
	if got != noRelationshipFallback {
		t.Fatalf("expected fallback %q, got %q", noRelationshipFallback, got)
	}
}
