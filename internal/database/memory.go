package database

import (
	"fmt"
	"graphql-pkm/internal/models"
	"sync"
)

type MemoryDB struct {
	notes map[string]*models.Note
	links map[string]*models.Link
	mu    sync.RWMutex
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		notes: make(map[string]*models.Note),
		links: make(map[string]*models.Link),
	}
}

func (db *MemoryDB) createNote(note *models.Note) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.notes[note.ID] = note
	return nil
}

func (db *MemoryDB) updateNote(note *models.Note) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	if _, exists := db.notes[note.ID]; !exists {
		return fmt.Errorf("note not found")
	}
	
	db.notes[note.ID] = note
	return nil
}


func (db *MemoryDB) getNote(id string) (*models.Note, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	note, exists := db.notes[id]
	if !exists {
		return nil, nil
	}
	return note, nil
}

func (db *MemoryDB) deleteNote(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	if _, exists := db.notes[id]; !exists {
		return fmt.Errorf("note not found to delete")
	}
	
	delete(db.notes,id)
	return nil
}


func (db *MemoryDB) getAllNotes() ([]*models.Note, error){
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	notes := make([]*models.Note, 0, len(db.notes))
	for _, note := range db.notes {
		notes = append(notes, note)
	}
	
	return notes, nil
}


func (db *MemoryDB) getNotesByTag(note *models.Note) error {
	return nil
}


func (db *MemoryDB) createLink(note *models.Note) error {
	return nil
}


func (db *MemoryDB) UpdateNote(note *models.Note) error {
	return nil
}


func (db *MemoryDB) getLinksByNote(note *models.Note) error {
	return nil
}

func (db *MemoryDB) searchNotes(note *models.Note) error {
	return nil
}
