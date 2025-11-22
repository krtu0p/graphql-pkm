package service

import (
	"crypto/rand"
	"fmt"
	"graphql-pkm/internal/database"
	"graphql-pkm/internal/models"
	"strings"
	"time"

)

type noteService struct {
	db  *database.MemoryDB
}

func newNoteService(db *database.MemoryDB) *noteService {
	return &noteService{db: db}
}


func generateId() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("Failed to generate ID: %w", err)
	}
	
	return fmt.Sprintf("%x", b), nil
}

func (s *noteService) createNote(title, content string, tags []string) (*models.Note, error) {
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("Tittle cannot be empty")
	}
	
	id, err := generateId()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	note := &models.Note{
		ID:			id,
		Title:		strings.TrimSpace(title),
		Content:	strings.TrimSpace(content),
		Tags: 		s.normalizeTags(tags),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	
	err = s.db.CreateNote(note)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}
	
	return note, nil
	
}

func (s *noteService) GetAllNotes() ([]*models.Note, error) {
	return s.db.GetAllNotes()
}

func (s *noteService) GetNote(id string) (*models.Note, error) {
	return s.db.GetNote(id)
}

func (s *noteService) GetNotesByTags(tag string) ([]*models.Note, error) {
	return  s.db.GetNotesByTag(tag)
}

func (s *noteService) UpdateNote(id string, title, content *string, tags []string) (*models.Note, error) {
	note, err := s.db.GetNote(id)
	if err != nil {
		return nil, err
	}
	
	if note == nil {
		return nil, fmt.Errorf("note doesn't exist")
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
	
	return note, nil
}

func (s *noteService) DeleteNote(id string) error {
	return s.db.DeleteNote(id)
}

func (s *noteService) GetLinks(noteID string) ([]*models.Link, error) {
	return s.db.GetLinksByNote(noteID)
}

func (s *noteService) normalizeTags(tags []string) []string {
	var normalized []string
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			normalized = append(normalized, strings.ToLower(trimmed))
		}
	}
	
	return normalized
}
