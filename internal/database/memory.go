package database

import (
	"fmt"
	"graphql-pkm/internal/models"
	"strings"
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

func (db *MemoryDB) CreateNote(note *models.Note) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.notes[note.ID] = note
	return nil
}

func (db *MemoryDB) UpdateNote(note *models.Note) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	if _, exists := db.notes[note.ID]; !exists {
		return fmt.Errorf("note not found")
	}
	
	db.notes[note.ID] = note
	return nil
}


func (db *MemoryDB) GetNote(id string) (*models.Note, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	note, exists := db.notes[id]
	if !exists {
		return nil, nil
	}
	return note, nil
}

func (db *MemoryDB) DeleteNote(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	if _, exists := db.notes[id]; !exists {
		return fmt.Errorf("note not found to delete")
	}
	
	delete(db.notes,id)
	return nil
}


func (db *MemoryDB) GetAllNotes() ([]*models.Note, error){
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	notes := make([]*models.Note, 0, len(db.notes))
	for _, note := range db.notes {
		notes = append(notes, note)
	}
	
	return notes, nil
}


func (db *MemoryDB) GetNotesByTag(tag string) ([]*models.Note, error) {
	db.mu.RLock()
	defer db.mu.Unlock()
	
	var results []*models.Note
	for _, note := range db.notes {
		for _, noteTag := range note.Tags {
			if noteTag == tag {
				results = append(results, note)
				break
			}
		}
	}
	
	return results, nil
}


func (db *MemoryDB) CreateLink(link *models.Link) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	db.links[link.ID] = link
	return nil
}


func (db *MemoryDB) GetLinksByNote(noteID string) ([]*models.Link, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	var links []*models.Link
	for _, link := range db.links {
		if link.SourceNoteID == noteID {
			links = append(links, link)
		}
	}
	
	return links, nil
}

func (db *MemoryDB) SearchNotes(query string) ([]*models.Note, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	var results []*models.Note
	query = strings.ToLower(query)
	
	for _, note := range db.notes {
		if strings.Contains(strings.ToLower(note.Title), query) ||
		strings.Contains(strings.ToLower(note.Content), query) {
			results = append(results, note)
		}
	}
	
	return results, nil
}
