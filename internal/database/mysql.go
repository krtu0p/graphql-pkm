package database

import (
	"database/sql"
	"encoding/json"
	"graphql-pkm/internal/gql/models"
	"log"
)

type MySQLDB struct {
	db *sql.DB
}

func NewMySQLDB(dataSourceName string) (*MySQLDB, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}
	
	if err := db.Ping(); err != nil {
		return nil, err
	}
	
	if err := createTables(db); err != nil {
		return nil, err
	}
	
	log.Println("successfully connected to mysql database")
	return &MySQLDB{db: db}, nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS notes(
			id VARCHAR(36) PRIMARY KEY,
			title TEXT NOT NULL,
			content LONGTEXT NOT NULL,
			tags JSON,
			created_at DATETIME(6) NOT NULL,
			updated_at DATETUME(6) NOT NULL,
			summary text,
			key_concepts JSON,
			INDEX idx_created_at (created_at)
			INDEX idx_updated_at (updated_at)
			)
		`)
	
	if err != nil {
		return err
	}
	
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
		return err
	}
	
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
		return err
	}
	return nil
}

func (m *MySQLDB) CreateNote(note *models.Note) error {
	tagsJSON, _ := json.Marshal(note.Tags)
	keyConceptsJson, _ := json.Marshal(note.KeyConcepts)
	
	_, err := m.db.Exec(`
		INSERT INTO notes (id, title, content, tags, created_at, updated_at, summary, key_concepts)
		VALUES (?,?,?,?,?,?,?,?)
	`, note.ID, note.Title, tagsJSON, note.CreatedAt, note.UpdatedAt, note.Summary, keyConceptsJson)
	
	return err
	
}

func (m *MySQLDB) GetNote(id string) (*models.Note, error) {
	var note models.Note
	var tagsJson, keyConceptsJSON string
	
	err := m.db.QueryRow(`
		SELECT * FROM notes where id = ?
	`, id).Scan(&note.ID, &note.Title, &note.Content, &tagsJson, &note.CreatedAt, &note.UpdatedAt, &note.Summary, &keyConceptsJSON)	
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	if err != nil {
		return nil, err
	}
	
	json.Unmarshal([]byte(tagsJson), &note.Tags)
	json.Unmarshal([]byte(keyConceptsJSON), &note.KeyConcepts)
	
	return &note, nil
}

func (m *MySQLDB) GetAllNotes() {
	
}

func (m *MySQLDB) UpdateNote() {
	
}

func (m *MySQLDB) DeleteNote() {
	
}

func (m *MySQLDB) GetNotesByTag() {
	
}
