package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type MySQLEmbeddingsCache struct {
	db *MySQLDB
}

func NewMySQLEmbeddingsCache(mysqlDB *MySQLDB) *MySQLEmbeddingsCache {
	return &MySQLEmbeddingsCache{db: mysqlDB}
}

func (ec *MySQLEmbeddingsCache) Store(noteID string, embedding []float32) error {
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}
	
	now := time.Now()
	
	_, err = ec.db.db.Exec(`
		INSERT INTO embeddings_cache (note_id, embedding_json, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			embedding_json = VALUES(embedding_json),
			updated_at = VALUES(updated_at)
	`, noteID, embeddingJSON, now, now)
	
	if err != nil {
		return fmt.Errorf("failed to store embedding: %w", err)
	}
	
	log.Printf("✓ Cached embedding for note: %s (size: %d dimensions)", noteID, len(embedding))
	return nil
}

func (ec *MySQLEmbeddingsCache) Get(noteID string) ([]float32, bool) {
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

func (ec *MySQLEmbeddingsCache) Delete(noteID string) error {
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

func (ec *MySQLEmbeddingsCache) Clear() error {
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

func (ec *MySQLEmbeddingsCache) Count() (int, error) {
	var count int
	err := ec.db.db.QueryRow(`SELECT COUNT(*) FROM embeddings_cache`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count embeddings: %w", err)
	}
	return count, nil
}

func (ec *MySQLEmbeddingsCache) GetCacheStats() (map[string]interface{}, error) {
	var count int
	var oldestCache, newestCache sql.NullTime
	
	err := ec.db.db.QueryRow(`
		SELECT 
			COUNT(*) as count,
			MIN(created_at) as oldest,
			MAX(created_at) as newest
		FROM embeddings_cache
	`).Scan(&count, &oldestCache, &newestCache)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get cache stats: %w", err)
	}
	
	stats := map[string]interface{}{
		"total_embeddings": count,
	}
	
	if oldestCache.Valid {
		stats["oldest_cache"] = oldestCache.Time
	}
	if newestCache.Valid {
		stats["newest_cache"] = newestCache.Time
	}
	
	return stats, nil
}