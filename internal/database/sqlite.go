package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"graphql-pkm/internal/models"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

type SQLiteDB struct {
	db *sql.DB
}

func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	log.Printf("Opening SQLite database at %s...", dbPath)

	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	if err := createSQLiteTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create sqlite tables: %w", err)
	}

	log.Println("✓ SQLite database ready")
	return &SQLiteDB{db: db}, nil
}

func createSQLiteTables(db *sql.DB) error {
	log.Println("Creating/verifying SQLite tables...")

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS notes (
			id VARCHAR(36) PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			tags TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			summary TEXT,
			key_concepts TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_created_at ON notes(created_at);
		CREATE INDEX IF NOT EXISTS idx_updated_at ON notes(updated_at);
	`)
	if err != nil {
		return fmt.Errorf("failed to create notes table: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			id VARCHAR(36) PRIMARY KEY,
			source_note_id VARCHAR(36) NOT NULL,
			target_note_id VARCHAR(36) NOT NULL,
			description TEXT,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (source_note_id) REFERENCES notes(id) ON DELETE CASCADE,
			FOREIGN KEY (target_note_id) REFERENCES notes(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_source_note ON links(source_note_id);
		CREATE INDEX IF NOT EXISTS idx_target_note ON links(target_note_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to create links table: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS embeddings_cache (
			note_id VARCHAR(36) PRIMARY KEY,
			model VARCHAR(255) NOT NULL DEFAULT '',
			dimensions INT NOT NULL DEFAULT 0,
			embedding_json TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create embeddings_cache table: %w", err)
	}

	// Ensure model and dimensions columns exist on pre-existing databases.
	_, _ = db.Exec(`ALTER TABLE embeddings_cache ADD COLUMN model VARCHAR(255) NOT NULL DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE embeddings_cache ADD COLUMN dimensions INT NOT NULL DEFAULT 0`)

	return nil
}

func (s *SQLiteDB) Close() error {
	if s.db != nil {
		log.Println("Closing SQLite database connection")
		return s.db.Close()
	}
	return nil
}

func (s *SQLiteDB) CreateNote(note *models.Note) error {
	tagsJSON, _ := json.Marshal(note.Tags)
	keyConceptsJSON, _ := json.Marshal(note.KeyConcepts)

	_, err := s.db.Exec(`
		INSERT INTO notes (id, title, content, tags, created_at, updated_at, summary, key_concepts)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, note.ID, note.Title, note.Content, string(tagsJSON), note.CreatedAt, note.UpdatedAt, note.Summary, string(keyConceptsJSON))

	return err
}

func (s *SQLiteDB) GetNote(id string) (*models.Note, error) {
	var note models.Note
	var tagsJSON, keyConceptsJSON sql.NullString

	err := s.db.QueryRow(`
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

func (s *SQLiteDB) GetAllNotes() ([]*models.Note, error) {
	rows, err := s.db.Query(`
		SELECT id, title, content, tags, created_at, updated_at, summary, key_concepts
		FROM notes ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSQLiteNotes(rows)
}

func (s *SQLiteDB) UpdateNote(note *models.Note) error {
	tagsJSON, _ := json.Marshal(note.Tags)
	keyConceptsJSON, _ := json.Marshal(note.KeyConcepts)

	_, err := s.db.Exec(`
		UPDATE notes
		SET title = ?, content = ?, tags = ?, updated_at = ?, summary = ?, key_concepts = ?
		WHERE id = ?
	`, note.Title, note.Content, string(tagsJSON), note.UpdatedAt, note.Summary, string(keyConceptsJSON), note.ID)

	return err
}

func (s *SQLiteDB) DeleteNote(id string) error {
	_, err := s.db.Exec("DELETE FROM notes WHERE id = ?", id)
	return err
}

func (s *SQLiteDB) GetNotesByTag(tag string) ([]*models.Note, error) {
	rows, err := s.db.Query(`
		SELECT id, title, content, tags, created_at, updated_at, summary, key_concepts
		FROM notes
		WHERE EXISTS (SELECT 1 FROM json_each(notes.tags) WHERE value = ?)
		ORDER BY created_at DESC
	`, tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSQLiteNotes(rows)
}

func (s *SQLiteDB) SearchNotes(query string) ([]*models.Note, error) {
	searchTerm := "%" + strings.ToLower(query) + "%"

	rows, err := s.db.Query(`
		SELECT id, title, content, tags, created_at, updated_at, summary, key_concepts
		FROM notes
		WHERE LOWER(title) LIKE ? OR LOWER(content) LIKE ?
		ORDER BY created_at DESC
	`, searchTerm, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSQLiteNotes(rows)
}

func (s *SQLiteDB) CreateLink(link *models.Link) error {
	_, err := s.db.Exec(`
		INSERT INTO links (id, source_note_id, target_note_id, description, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, link.ID, link.SourceNoteID, link.TargetNoteID, link.Description, link.CreatedAt)
	return err
}

func (s *SQLiteDB) GetLinksByNote(noteID string) ([]*models.Link, error) {
	rows, err := s.db.Query(`
		SELECT id, source_note_id, target_note_id, description, created_at
		FROM links
		WHERE source_note_id = ?
	`, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSQLiteLinks(rows)
}

func (s *SQLiteDB) DeleteLink(id string) error {
	result, err := s.db.Exec("DELETE FROM links WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no link found")
	}

	return nil
}

func (s *SQLiteDB) GetBackLinksByNote(noteID string) ([]*models.Link, error) {
	rows, err := s.db.Query(`
		SELECT id, source_note_id, target_note_id, description, created_at
		FROM links
		WHERE target_note_id = ?
		ORDER BY created_at DESC
	`, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSQLiteLinks(rows)
}

func scanSQLiteNotes(rows *sql.Rows) ([]*models.Note, error) {
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
	return notes, rows.Err()
}

func scanSQLiteLinks(rows *sql.Rows) ([]*models.Link, error) {
	var links []*models.Link
	for rows.Next() {
		var link models.Link
		var description sql.NullString

		err := rows.Scan(&link.ID, &link.SourceNoteID, &link.TargetNoteID, &description, &link.CreatedAt)
		if err != nil {
			return nil, err
		}

		if description.Valid {
			link.Description = &description.String
		}

		links = append(links, &link)
	}
	return links, rows.Err()
}
