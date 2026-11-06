package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type SQLiteEmbeddingsCache struct {
	db *SQLiteDB
}

func NewSQLiteEmbeddingsCache(sqliteDB *SQLiteDB) *SQLiteEmbeddingsCache {
	return &SQLiteEmbeddingsCache{db: sqliteDB}
}

func (ec *SQLiteEmbeddingsCache) Store(noteID string, embedding []float32) error {
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	now := time.Now()

	_, err = ec.db.db.Exec(`
		INSERT INTO embeddings_cache (note_id, embedding_json, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(note_id) DO UPDATE SET
			embedding_json = excluded.embedding_json,
			updated_at = excluded.updated_at
	`, noteID, string(embeddingJSON), now, now)

	if err != nil {
		return fmt.Errorf("failed to store embedding: %w", err)
	}

	log.Printf("✓ Cached embedding for note: %s (size: %d dimensions)", noteID, len(embedding))
	return nil
}

func (ec *SQLiteEmbeddingsCache) Get(noteID string) ([]float32, bool) {
	var embeddingJSON string

	err := ec.db.db.QueryRow(`
		SELECT embedding_json FROM embeddings_cache
		WHERE note_id = ?
	`, noteID).Scan(&embeddingJSON)

	if err == sql.ErrNoRows {
		log.Printf("⚠️  Cache miss for note: %s", noteID)
		return nil, false
	}

	if err != nil {
		log.Printf("❌ Error retrieving embedding for note %s: %v", noteID, err)
		return nil, false
	}

	var embedding []float32
	if err := json.Unmarshal([]byte(embeddingJSON), &embedding); err != nil {
		log.Printf("❌ Error unmarshaling embedding for note %s: %v", noteID, err)
		return nil, false
	}

	log.Printf("✓ Cache hit for note: %s (%d dimensions)", noteID, len(embedding))
	return embedding, true
}

func (ec *SQLiteEmbeddingsCache) Delete(noteID string) error {
	result, err := ec.db.db.Exec(`
		DELETE FROM embeddings_cache WHERE note_id = ?
	`, noteID)

	if err != nil {
		return fmt.Errorf("failed to delete embedding: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected > 0 {
		log.Printf("✓ Deleted embedding cache for note: %s", noteID)
	}

	return nil
}

func (ec *SQLiteEmbeddingsCache) Clear() error {
	result, err := ec.db.db.Exec(`DELETE FROM embeddings_cache`)
	if err != nil {
		return fmt.Errorf("failed to clear embedding cache: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	log.Printf("✓ Cleared embedding cache (%d embeddings removed)", rowsAffected)
	return nil
}

func (ec *SQLiteEmbeddingsCache) Count() (int, error) {
	var count int
	err := ec.db.db.QueryRow(`SELECT COUNT(*) FROM embeddings_cache`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count embeddings: %w", err)
	}
	return count, nil
}

func (ec *SQLiteEmbeddingsCache) GetCacheStats() (map[string]interface{}, error) {
	count, err := ec.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_embeddings": count,
	}

	if count == 0 {
		return stats, nil
	}

	var oldestCache, newestCache time.Time
	if err := ec.db.db.QueryRow(`SELECT created_at FROM embeddings_cache ORDER BY created_at ASC LIMIT 1`).Scan(&oldestCache); err == nil {
		stats["oldest_cache"] = oldestCache
	}
	if err := ec.db.db.QueryRow(`SELECT created_at FROM embeddings_cache ORDER BY created_at DESC LIMIT 1`).Scan(&newestCache); err == nil {
		stats["newest_cache"] = newestCache
	}

	return stats, nil
}
