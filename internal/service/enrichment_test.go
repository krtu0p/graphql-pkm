package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"graphql-pkm/internal/ai"
	"graphql-pkm/internal/database"
)

// aiStub counts chat calls and can delay responses to simulate a slow model.
// maxInFlight records the highest number of concurrent AI calls, which is 1
// while the worker processes jobs serially.
type aiStub struct {
	calls       atomic.Int64
	delay       time.Duration
	inFlight    atomic.Int64
	maxInFlight atomic.Int64
}

func (s *aiStub) handler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current := s.inFlight.Add(1)
		for {
			max := s.maxInFlight.Load()
			if current <= max || s.maxInFlight.CompareAndSwap(max, current) {
				break
			}
		}
		defer s.inFlight.Add(-1)

		if s.delay > 0 {
			time.Sleep(s.delay)
		}

		var req struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		userPrompt := ""
		if len(req.Messages) > 0 {
			userPrompt = req.Messages[len(req.Messages)-1].Content
		}
		s.calls.Add(1)

		reply := "Concise generated summary."
		if strings.Contains(userPrompt, "JSON array") {
			reply = "```json\n[\"concept-a\", \"concept-b\"]\n```"
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
	}
}

// waitForCalls polls until the stub has seen at least `want` chat calls.
func waitForCalls(t *testing.T, stub *aiStub, want int64) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if stub.calls.Load() >= want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d AI calls, got %d", want, stub.calls.Load())
}

// waitUntilQuiet waits for the expected call total, then asserts nothing else
// arrives during a short settle window (i.e. no extra jobs got queued).
func waitUntilQuiet(t *testing.T, stub *aiStub, want int64) {
	t.Helper()
	waitForCalls(t, stub, want)
	time.Sleep(200 * time.Millisecond)
	if got := stub.calls.Load(); got != want {
		t.Fatalf("expected exactly %d AI calls after settling, got %d", want, got)
	}
}

func newEnrichmentNoteService(t *testing.T, debounce time.Duration, delay time.Duration) (*NoteService, *aiStub) {
	t.Helper()

	stub := &aiStub{delay: delay}
	srv := httptest.NewServer(stub.handler(t))
	t.Cleanup(srv.Close)

	aiClient := ai.NewClientWithConfig(ai.ClientConfig{
		Enabled: true,
		BaseURL: srv.URL + "/v1",
		Model:   "test-model",
		Timeout: 30 * time.Second,
	})
	return newNoteService(database.NewMemoryDB(), aiClient, debounce), stub
}

func ptr(s string) *string { return &s }

func TestEnrichmentCoalescesRapidUpdates(t *testing.T) {
	// Debounce must outlast the gap between the two sequential UpdateNote
	// DB round-trips so both land inside one debounce window.
	svc, stub := newEnrichmentNoteService(t, 150*time.Millisecond, 0)

	note, err := svc.CreateNote("Note", "original body", nil)
	if err != nil {
		t.Fatalf("CreateNote: %v", err)
	}
	// Creation enriches once (summary + concepts = 2 AI calls).
	waitUntilQuiet(t, stub, 2)

	// Two quick updates must coalesce into a single job (one more AI run).
	title2 := "Note two"
	if _, err := svc.UpdateNote(note.ID, &title2, ptr("body v2"), nil); err != nil {
		t.Fatalf("UpdateNote 1: %v", err)
	}
	title3 := "Note three"
	if _, err := svc.UpdateNote(note.ID, &title3, ptr("body v3"), nil); err != nil {
		t.Fatalf("UpdateNote 2: %v", err)
	}

	waitUntilQuiet(t, stub, 4)

	updated, err := svc.GetNote(note.ID)
	if err != nil {
		t.Fatalf("GetNote: %v", err)
	}
	if updated.Title != "Note three" || updated.Content != "body v3" {
		t.Fatalf("expected latest text persisted, got %q / %q", updated.Title, updated.Content)
	}
}

func TestEnrichmentSkippedWhenUnchanged(t *testing.T) {
	svc, stub := newEnrichmentNoteService(t, 20*time.Millisecond, 0)

	note, err := svc.CreateNote("Note", "original body", nil)
	if err != nil {
		t.Fatalf("CreateNote: %v", err)
	}
	waitUntilQuiet(t, stub, 2)

	// Save again with identical text; this must not call the AI.
	title := "Note"
	if _, err := svc.UpdateNote(note.ID, &title, ptr("original body"), nil); err != nil {
		t.Fatalf("UpdateNote: %v", err)
	}

	waitUntilQuiet(t, stub, 2)
}

func TestEnrichmentSingleFollowUpDuringRunningJob(t *testing.T) {
	// Delay makes each job long enough that both updates land while the
	// creation job is still running.
	svc, stub := newEnrichmentNoteService(t, 10*time.Millisecond, 150*time.Millisecond)

	note, err := svc.CreateNote("Note", "original body", nil)
	if err != nil {
		t.Fatalf("CreateNote: %v", err)
	}

	// Wait until the creation job is running (its first AI call started).
	waitForCalls(t, stub, 1)

	// Two updates while the job is running: they must coalesce into exactly
	// one pending follow-up, not one job per update.
	title2 := "Note two"
	if _, err := svc.UpdateNote(note.ID, &title2, ptr("body v2"), nil); err != nil {
		t.Fatalf("UpdateNote 1: %v", err)
	}
	title3 := "Note three"
	if _, err := svc.UpdateNote(note.ID, &title3, ptr("body v3"), nil); err != nil {
		t.Fatalf("UpdateNote 2: %v", err)
	}

	// Creation job (2 calls) + exactly one follow-up job (2 calls).
	waitUntilQuiet(t, stub, 4)
}

func TestEnrichmentWorkerRunsOnlyOneJobAtATime(t *testing.T) {
	svc, stub := newEnrichmentNoteService(t, 10*time.Millisecond, 30*time.Millisecond)

	note, err := svc.CreateNote("Note", "original body", nil)
	if err != nil {
		t.Fatalf("CreateNote: %v", err)
	}

	// Wait until the creation job is running, then queue a follow-up.
	waitForCalls(t, stub, 1)
	title2 := "Note two"
	if _, err := svc.UpdateNote(note.ID, &title2, ptr("body v2"), nil); err != nil {
		t.Fatalf("UpdateNote: %v", err)
	}

	// Both jobs run back to back; AI calls must never overlap.
	waitUntilQuiet(t, stub, 4)
	if max := stub.maxInFlight.Load(); max != 1 {
		t.Fatalf("worker ran %d AI calls in parallel, expected max 1", max)
	}
}

func TestFlushPendingEnrichmentRunsBeforeDebounce(t *testing.T) {
	// A debounce long enough that the natural timer would never fire during
	// the test: the flush must fire the pending job itself.
	svc, stub := newEnrichmentNoteService(t, time.Hour, 0)

	if _, err := svc.CreateNote("Note", "original body", nil); err != nil {
		t.Fatalf("CreateNote: %v", err)
	}

	svc.FlushPendingEnrichment(5 * time.Second)
	waitUntilQuiet(t, stub, 2)
}
