package database

import "graphql-pkm/internal/models"

type Database interface {
	CreateNote(note *models.Note) error
	GetNote(id string) (*models.Note, error)
	GetAllNotes() ([]*models.Note, error)
	UpdateNote(note *models.Note) error
	DeleteNote(id string) error
	GetNotesByTag(tag string) ([]*models.Note, error)
	SearchNotes(query string) ([]*models.Note, error)
	CreateLink(link *models.Link) error
	GetLinksByNote(noteID string) ([]*models.Link, error)
	DeleteLink(linkID string) error
	GetBackLinksByNote(noteID string) ([]*models.Link, error)
}

var _ Database = (*MemoryDB)(nil)
var _ Database = (*MySQLDB)(nil)