package service

import (
	"strings"
	"crypto/rand"
	"fmt"
	"graphql-pkm/internal/database"
	"graphql-pkm/internal/models"
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

func (s *noteService) createNote(title, content string, tags []string) ([]*models.Note, error) {
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
	
}

func (s *noteService) normalizeTags(tags []string) []string {
	var normalized []string
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			normalized = append(normalized, strings.ToLower(trimmed))
		}
	}
}
