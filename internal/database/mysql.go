package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"graphql-pkm/internal/models"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLDB struct {
	db *sql.DB
}

func NewMySQLDB(dataSourceName string) (*MySQLDB, error) {
	log.Printf("Connecting to MySQL database...")
	
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	log.Println("✓ Database connection established")
	
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}
	
	log.Println("✓ Database tables verified/created successfully")
	return &MySQLDB{db: db}, nil
}

func createTables(db *sql.DB) error {
	log.Println("Creating/verifying database tables...")
	
	// Create notes table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS notes(
			id VARCHAR(36) PRIMARY KEY,
			title TEXT NOT NULL,
			content LONGTEXT NOT NULL,
			tags JSON,
			created_at DATETIME(6) NOT NULL,
			updated_at DATETIME(6) NOT NULL,
			summary TEXT,
			key_concepts JSON,
			INDEX idx_created_at (created_at),
			INDEX idx_updated_at (updated_at)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create notes table: %w", err)
	}
	log.Println("  ✓ notes table ready")
	
	// Create links table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			id VARCHAR(36) PRIMARY KEY,
			source_note_id VARCHAR(36) NOT NULL,
			target_note_id VARCHAR(36) NOT NULL,
			description TEXT,
			created_at DATETIME(6) NOT NULL,
			FOREIGN KEY (source_note_id) REFERENCES notes(id) ON DELETE CASCADE,
			FOREIGN KEY (target_note_id) REFERENCES notes(id) ON DELETE CASCADE,
			INDEX idx_source_note (source_note_id),
			INDEX idx_target_note (target_note_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create links table: %w", err)
	}
	log.Println("  ✓ links table ready")
	
	// Create embeddings_cache table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS embeddings_cache (
			note_id VARCHAR(36) PRIMARY KEY,
			embedding_json JSON NOT NULL,
			created_at DATETIME(6) NOT NULL,
			updated_at DATETIME(6) NOT NULL,
			FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create embeddings_cache table: %w", err)
	}
	log.Println("  ✓ embeddings_cache table ready")
	
	return nil
}

func (m *MySQLDB) Close() error {
	if m.db != nil {
		log.Println("Closing MySQL database connection")
		return m.db.Close()
	}
	return nil
}

func (m *MySQLDB) CreateNote(note *models.Note) error {
	tagsJSON, _ := json.Marshal(note.Tags)
	keyConceptsJSON, _ := json.Marshal(note.KeyConcepts)
	
	_, err := m.db.Exec(`
		INSERT INTO notes (id, title, content, tags, created_at, updated_at, summary, key_concepts)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, note.ID, note.Title, note.Content, tagsJSON, note.CreatedAt, note.UpdatedAt, note.Summary, keyConceptsJSON)
	
	return err
}

func (m *MySQLDB) GetNote(id string) (*models.Note, error) {
	var note models.Note
	var tagsJSON, keyConceptsJSON sql.NullString
	
	err := m.db.QueryRow(`
		SELECT id, title, content, tags, created_at, updated_at, summary, key_concepts
		FROM notes WHERE id = ?
	`, id).Scan(&note.ID, &note.Title, &note.Content, &tagsJSON, &note.CreatedAt, &note.UpdatedAt, &note.Summary, &keyConceptsJSON)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	if err != nil {
		return nil, err
	}
	
	if tagsJSON.Valid {
		json.Unmarshal([]byte(tagsJSON.String), &note.Tags)
	}
	if keyConceptsJSON.Valid {
		json.Unmarshal([]byte(keyConceptsJSON.String), &note.KeyConcepts)
	}
	
	return &note, nil
}

func (m *MySQLDB) GetAllNotes() ([]*models.Note, error) {
	rows, err := m.db.Query(`
		SELECT id, title, content, tags, created_at, updated_at, summary, key_concepts
		FROM notes ORDER BY created_at DESC
	`)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var notes []*models.Note
	for rows.Next() {
		var note models.Note
		var tagsJSON, keyConceptsJSON sql.NullString
		
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &tagsJSON, &note.CreatedAt, &note.UpdatedAt, &note.Summary, &keyConceptsJSON)
		if err != nil {
			return nil, err
		}
		
		if tagsJSON.Valid {
			json.Unmarshal([]byte(tagsJSON.String), &note.Tags)
		}
		if keyConceptsJSON.Valid {
			json.Unmarshal([]byte(keyConceptsJSON.String), &note.KeyConcepts)
		}
		notes = append(notes, &note)
	}
	
	return notes, nil
}

func (m *MySQLDB) UpdateNote(note *models.Note) error {
	tagsJSON, _ := json.Marshal(note.Tags)
	keyConceptsJSON, _ := json.Marshal(note.KeyConcepts)
	
	_, err := m.db.Exec(`
		UPDATE notes
		SET title = ?, content = ?, tags = ?, updated_at = ?, summary = ?, key_concepts = ?
		WHERE id = ?
	`, note.Title, note.Content, tagsJSON, note.UpdatedAt, note.Summary, keyConceptsJSON, note.ID)
	
	return err
}

func (m *MySQLDB) DeleteNote(id string) error {
	_, err := m.db.Exec("DELETE FROM notes WHERE id = ?", id)
	return err
}

func (m *MySQLDB) GetNotesByTag(tag string) ([]*models.Note, error) {
	rows, err := m.db.Query(`
		SELECT id, title, content, tags, created_at, updated_at, summary, key_concepts
		FROM notes WHERE JSON_CONTAINS(tags, ?)
		ORDER BY created_at DESC
	`, fmt.Sprintf(`"%s"`, tag))
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var notes []*models.Note
	for rows.Next() {
		var note models.Note
		var tagsJSON, keyConceptsJSON sql.NullString
		
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &tagsJSON, &note.CreatedAt, &note.UpdatedAt, &note.Summary, &keyConceptsJSON)
		if err != nil {
			return nil, err
		}
		
		if tagsJSON.Valid {
			json.Unmarshal([]byte(tagsJSON.String), &note.Tags)
		}
		if keyConceptsJSON.Valid {
			json.Unmarshal([]byte(keyConceptsJSON.String), &note.KeyConcepts)
		}
		notes = append(notes, &note)
	}
	
	return notes, nil
}

func (m *MySQLDB) SearchNotes(query string) ([]*models.Note, error) {
	searchTerm := "%" + strings.ToLower(query) + "%"
	
	rows, err := m.db.Query(`
		SELECT id, title, content, tags, created_at, updated_at, summary, key_concepts
		FROM notes 
		WHERE LOWER(title) LIKE ? OR LOWER(content) LIKE ?
		ORDER BY created_at DESC
	`, searchTerm, searchTerm)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var notes []*models.Note
	for rows.Next() {
		var note models.Note
		var tagsJSON, keyConceptsJSON sql.NullString
		
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &tagsJSON, &note.CreatedAt, &note.UpdatedAt, &note.Summary, &keyConceptsJSON)
		if err != nil {
			return nil, err
		}
		
		if tagsJSON.Valid {
			json.Unmarshal([]byte(tagsJSON.String), &note.Tags)
		}
		if keyConceptsJSON.Valid {
			json.Unmarshal([]byte(keyConceptsJSON.String), &note.KeyConcepts)
		}
		
		notes = append(notes, &note)
	}
	
	return notes, nil
}

func (m *MySQLDB) CreateLink(link *models.Link) error {
	_, err := m.db.Exec(`
		INSERT INTO links (id, source_note_id, target_note_id, description, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, link.ID, link.SourceNoteID, link.TargetNoteID, link.Description, link.CreatedAt)
	return err
}

func (m *MySQLDB) GetLinksByNote(noteID string) ([]*models.Link, error) {
	rows, err := m.db.Query(`
		SELECT id, source_note_id, target_note_id, description, created_at
		FROM links
		WHERE source_note_id = ?
	`, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var links []*models.Link
	for rows.Next() {
		var link models.Link
		err := rows.Scan(&link.ID, &link.SourceNoteID, &link.TargetNoteID, &link.Description, &link.CreatedAt)
		if err != nil {
			return nil, err
		}
		links = append(links, &link)
	}
	
	return links, nil
}