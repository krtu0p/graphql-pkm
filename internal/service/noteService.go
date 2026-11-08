package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"graphql-pkm/internal/ai"
	"graphql-pkm/internal/database"
	"graphql-pkm/internal/models"
	"log"
	"strings"
	"sync"
	"time"
)

// enrichDebounceInterval delays each enrichment job so rapid consecutive
// edits coalesce into a single AI run. Tests shorten it via newNoteService.
const enrichDebounceInterval = 15 * time.Second

type enrichJob struct {
	noteID  string
	title   string
	content string
}

// pendingDispatch is a debounced enrichment that has not fired yet. Doubles
// as the timer callback's identity: a dispatch that is no longer the
// registered entry for its note, or is cancelled, is dropped. All field
// writes happen before the entry is published or under enrichMu.
type pendingDispatch struct {
	job       enrichJob
	cancelled bool
}

type NoteService struct {
	db       database.Database
	aiClient *ai.Client
	enrichCh chan enrichJob

	enrichMu       sync.Mutex
	enrichDebounce time.Duration
	enrichedHashes map[string]string    // note ID -> hash at last successful enrichment
	queuedNotes    map[string]bool      // note IDs with a job queued or running
	pendingJobs    map[string]enrichJob // at most one follow-up job per note
	pendingTimers  map[string]*pendingDispatch
}

func NewNoteService(db database.Database, aiClient *ai.Client) *NoteService {
	return newNoteService(db, aiClient, enrichDebounceInterval)
}

// newNoteService lets tests pass a short debounce interval. Shared logic with
// NewNoteService (also starts the enrichment worker).
func newNoteService(db database.Database, aiClient *ai.Client, debounce time.Duration) *NoteService {
	s := &NoteService{
		db:             db,
		aiClient:       aiClient,
		enrichCh:       make(chan enrichJob, 100),
		enrichDebounce: debounce,
		enrichedHashes: map[string]string{},
		queuedNotes:    map[string]bool{},
		pendingJobs:    map[string]enrichJob{},
		pendingTimers:  map[string]*pendingDispatch{},
	}
	go s.enrichmentWorker()
	return s
}

func (s *NoteService) enrichmentWorker() {
	for job := range s.enrichCh {
		s.processEnrichJob(job)
		s.finishEnrichJob(job)
	}
}

// enrichmentContentHash identifies the title/content pair a note was last
// successfully enriched with, so unchanged saves do not call the AI again.
func enrichmentContentHash(title, content string) string {
	sum := sha256.Sum256([]byte(title + "\x00" + content))
	return hex.EncodeToString(sum[:])
}

func (s *NoteService) enqueueEnrichment(noteID, title, content string) {
	if s.aiClient == nil || !s.aiClient.IsEnabled() {
		return
	}

	hash := enrichmentContentHash(title, content)

	s.enrichMu.Lock()
	if s.enrichedHashes[noteID] == hash {
		s.enrichMu.Unlock()
		log.Printf("AI enrichment skipped (unchanged) for note %s", noteID)
		return
	}
	if s.queuedNotes[noteID] {
		// A job for this note is queued or running; keep only the latest
		// content as a single pending follow-up.
		s.pendingJobs[noteID] = enrichJob{noteID: noteID, title: title, content: content}
		s.enrichMu.Unlock()
		log.Printf("AI enrichment coalesced for note %s", noteID)
		return
	}
	// Otherwise coalesce into the pending debounce timer: cancel the
	// previous one and reschedule with the newest text, so a burst of saves
	// leads to one run.
	if existing := s.pendingTimers[noteID]; existing != nil {
		existing.cancelled = true
		log.Printf("AI enrichment coalesced for note %s", noteID)
	}
	debounce := s.enrichDebounce
	entry := &pendingDispatch{job: enrichJob{noteID: noteID, title: title, content: content}}
	s.pendingTimers[noteID] = entry
	s.enrichMu.Unlock()

	time.AfterFunc(debounce, func() { s.dispatchEnrichment(entry) })
}

// dispatchEnrichment runs the debounced callback for a dispatch entry. It is
// a no-op when the entry was cancelled or replaced by a newer save.
func (s *NoteService) dispatchEnrichment(entry *pendingDispatch) {
	job := entry.job
	s.enrichMu.Lock()
	if s.pendingTimers[job.noteID] != entry || entry.cancelled {
		s.enrichMu.Unlock()
		return
	}
	delete(s.pendingTimers, job.noteID)
	s.enrichMu.Unlock()
	s.submitEnrichJob(job)
}

// submitEnrichJob claims the queue slot for a note and pushes the job to the
// worker. Dropped when the text was already successfully enriched, already
// queued/running (kept as a single pending follow-up), or the channel is full.
func (s *NoteService) submitEnrichJob(job enrichJob) {
	s.enrichMu.Lock()
	if s.enrichedHashes[job.noteID] == enrichmentContentHash(job.title, job.content) {
		s.enrichMu.Unlock()
		return
	}
	if s.queuedNotes[job.noteID] {
		s.pendingJobs[job.noteID] = job
		s.enrichMu.Unlock()
		return
	}
	s.queuedNotes[job.noteID] = true
	s.enrichMu.Unlock()

	select {
	case s.enrichCh <- job:
	default:
		s.enrichMu.Lock()
		delete(s.queuedNotes, job.noteID)
		s.enrichMu.Unlock()
		log.Printf("Warning: AI enrichment queue full, skipping note %s", job.noteID)
	}
}

// finishEnrichJob releases the note's queue slot after a run and starts at
// most one pending follow-up job collected while this job was running.
func (s *NoteService) finishEnrichJob(job enrichJob) {
	s.enrichMu.Lock()
	delete(s.queuedNotes, job.noteID)
	pending, hasPending := s.pendingJobs[job.noteID]
	if hasPending {
		delete(s.pendingJobs, job.noteID)
	}
	s.enrichMu.Unlock()

	if hasPending {
		s.enqueueEnrichment(pending.noteID, pending.title, pending.content)
	}
}

// FlushPendingEnrichment fires pending debounced jobs immediately and waits
// (bounded by the timeout) for queued and running jobs to finish, so saves
// made shortly before shutdown are still enriched. Call it after the HTTP
// server has stopped; it is best effort and logs on timeout.
func (s *NoteService) FlushPendingEnrichment(timeout time.Duration) {
	if s.aiClient == nil || !s.aiClient.IsEnabled() {
		return
	}

	deadline := time.Now().Add(timeout)
	// Only exit after two consecutive empty rounds, so a debounce timer
	// registered between snapshots (e.g. by finishEnrichJob handling a
	// pending follow-up) is still picked up.
	emptyRound := false
	for {
		// Cancel everything pending so the debounce callback no-ops, and
		// dispatch the collected jobs without waiting for their timers.
		s.enrichMu.Lock()
		busy := len(s.queuedNotes)
		pendingFollowUps := len(s.pendingJobs)
		var immediate []enrichJob
		for id, entry := range s.pendingTimers {
			if !entry.cancelled {
				immediate = append(immediate, entry.job)
			}
			entry.cancelled = true
			delete(s.pendingTimers, id)
		}
		s.enrichMu.Unlock()

		for _, job := range immediate {
			s.submitEnrichJob(job)
		}

		if busy == 0 && pendingFollowUps == 0 && len(immediate) == 0 {
			if emptyRound {
				return
			}
			emptyRound = true
		} else {
			emptyRound = false
		}
		if time.Now().After(deadline) {
			s.enrichMu.Lock()
			remaining := len(s.queuedNotes) + len(s.pendingJobs) + len(s.pendingTimers)
			s.enrichMu.Unlock()
			log.Printf("Warning: AI enrichment flush timed out, %d job(s) still pending", remaining)
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func (s *NoteService) processEnrichJob(job enrichJob) {
	if s.aiClient == nil || !s.aiClient.IsEnabled() {
		return
	}

	note, err := s.db.GetNote(job.noteID)
	if err != nil {
		log.Printf("Warning: Failed to fetch note %s for AI enrichment: %v", job.noteID, err)
		return
	}
	if note == nil {
		return
	}

	log.Printf("Generating AI metadata in background for note %s", job.noteID)

	var newSummary *string
	summary, err := s.aiClient.GenerateSummary(job.title, job.content)
	if err != nil {
		log.Printf("Warning: Failed to generate summary for note %s: %v", job.noteID, err)
	} else if summary != "" {
		newSummary = &summary
	}

	var newConcepts []string
	hasConcepts := false
	concepts, err := s.aiClient.ExtractKeyConcepts(job.title, job.content)
	if err != nil {
		log.Printf("Warning: Failed to extract key concepts for note %s: %v", job.noteID, err)
	} else {
		newConcepts = concepts
		hasConcepts = true
	}

	if newSummary == nil && !hasConcepts {
		return
	}

	if newSummary != nil {
		note.Summary = newSummary
	}
	if hasConcepts {
		note.KeyConcepts = newConcepts
	}
	note.UpdatedAt = time.Now()

	if err := s.db.UpdateNote(note); err != nil {
		log.Printf("Warning: Failed to update note %s with AI metadata: %v", job.noteID, err)
	} else {
		log.Printf("✓ Completed AI metadata enrichment for note %s", job.noteID)
		// Remember the enriched text so unchanged saves skip the AI.
		s.enrichMu.Lock()
		s.enrichedHashes[job.noteID] = enrichmentContentHash(job.title, job.content)
		s.enrichMu.Unlock()
	}
}

func generateId() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate ID: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}

func (s *NoteService) CreateNote(title, content string, tags []string) (*models.Note, error) {
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}

	id, err := generateId()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	note := &models.Note{
		ID:          id,
		Title:       strings.TrimSpace(title),
		Content:     strings.TrimSpace(content),
		Tags:        s.normalizeTags(tags),
		CreatedAt:   now,
		UpdatedAt:   now,
		KeyConcepts: []string{},
	}

	err = s.db.CreateNote(note)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	s.enqueueEnrichment(note.ID, note.Title, note.Content)

	return note, nil
}

func (s *NoteService) GetAllNotes() ([]*models.Note, error) {
	return s.db.GetAllNotes()
}

func (s *NoteService) GetNote(id string) (*models.Note, error) {
	return s.db.GetNote(id)
}

func (s *NoteService) SearchNotes(query string) ([]*models.Note, error) {
	if strings.TrimSpace(query) == "" {
		return []*models.Note{}, nil
	}
	return s.db.SearchNotes(query)
}

func (s *NoteService) GetNotesByTag(tag string) ([]*models.Note, error) {
	return s.db.GetNotesByTag(tag)
}

func (s *NoteService) UpdateNote(id string, title, content *string, tags []string) (*models.Note, error) {
	note, err := s.db.GetNote(id)
	if err != nil {
		return nil, err
	}

	if note == nil {
		return nil, fmt.Errorf("note does not exist")
	}

	if title != nil {
		note.Title = strings.TrimSpace(*title)
	}

	if content != nil {
		note.Content = strings.TrimSpace(*content)
	}

	if tags != nil {
		note.Tags = s.normalizeTags(tags)
	}

	note.UpdatedAt = time.Now()

	err = s.db.UpdateNote(note)
	if err != nil {
		return nil, err
	}

	if title != nil || content != nil {
		s.enqueueEnrichment(note.ID, note.Title, note.Content)
	}

	return note, nil
}

func (s *NoteService) DeleteNote(id string) error {
	// Drop any debounce timer, pending follow-up and hash for the note.
	s.enrichMu.Lock()
	if entry := s.pendingTimers[id]; entry != nil {
		entry.cancelled = true
		delete(s.pendingTimers, id)
	}
	delete(s.pendingJobs, id)
	delete(s.enrichedHashes, id)
	s.enrichMu.Unlock()

	return s.db.DeleteNote(id)
}

func (s *NoteService) GetLinks(noteID string) ([]*models.Link, error) {
	return s.db.GetLinksByNote(noteID)
}

func (s *NoteService) CreateLink(sourceNoteID, targetNoteID, description string) (*models.Link, error) {
	sourceNote, err := s.db.GetNote(sourceNoteID)
	if err != nil {
		return nil, err
	}

	if sourceNote == nil {
		return nil, fmt.Errorf("source note does not exist")
	}

	targetNote, err := s.db.GetNote(targetNoteID)
	if err != nil {
		return nil, err
	}

	if targetNote == nil {
		return nil, fmt.Errorf("target note does not exist")
	}

	if sourceNoteID == targetNoteID {
		return nil, fmt.Errorf("cannot link a note to itself")
	}

	id, err := generateId()
	if err != nil {
		return nil, err
	}

	link := &models.Link{
		ID:           id,
		SourceNoteID: sourceNoteID,
		TargetNoteID: targetNoteID,
		Description:  &description,
		CreatedAt:    time.Now(),
	}

	err = s.db.CreateLink(link)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (s *NoteService) DeleteLink(linkID string) error {
	return s.db.DeleteLink(linkID)
}

func (s *NoteService) GetBackLinks(noteID string) ([]*models.Link, error) {
	return s.db.GetBackLinksByNote(noteID)
}

func (s *NoteService) normalizeTags(tags []string) []string {
	var normalized []string
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			normalized = append(normalized, strings.ToLower(trimmed))
		}
	}
	return normalized
}
