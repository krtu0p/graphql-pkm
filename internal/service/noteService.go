package service

import (
	"crypto/rand"
	"fmt"
	"graphql-pkm/internal/ai"
	"graphql-pkm/internal/database"
	"graphql-pkm/internal/models"
	"log"
	"strings"
	"time"
)

type enrichJob struct {
	noteID  string
	title   string
	content string
}

type NoteService struct {
	db       database.Database
	aiClient *ai.Client
	enrichCh chan enrichJob
}

func NewNoteService(db database.Database, aiClient *ai.Client) *NoteService {
	s := &NoteService{
		db:       db,
		aiClient: aiClient,
		enrichCh: make(chan enrichJob, 100),
	}
	go s.enrichmentWorker()
	return s
}

func (s *NoteService) enrichmentWorker() {
	for job := range s.enrichCh {
		s.processEnrichJob(job)
	}
}

func (s *NoteService) enqueueEnrichment(noteID, title, content string) {
	if s.aiClient == nil || !s.aiClient.IsEnabled() {
		return
	}
	select {
	case s.enrichCh <- enrichJob{noteID: noteID, title: title, content: content}:
	default:
		log.Printf("Warning: AI enrichment queue full, skipping note %s", noteID)
	}
}

func (s *NoteService) processEnrichJob(job enrichJob) {
	if s.aiClient == nil || !s.aiClient.IsEnabled() {
		return
	}
	log.Printf("Generating AI metadata in background for note: %s", job.title)

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

	note, err := s.db.GetNote(job.noteID)
	if err != nil {
		log.Printf("Warning: Failed to fetch note %s for AI enrichment: %v", job.noteID, err)
		return
	}
	if note == nil {
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
		log.Printf("✓ Completed AI metadata enrichment for note: %s", job.noteID)
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
