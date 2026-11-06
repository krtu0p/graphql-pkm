package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type MySQLEmbeddingsCache struct {
	db    *MySQLDB
	model string
}

func NewMySQLEmbeddingsCache(mysqlDB *MySQLDB, model ...string) *MySQLEmbeddingsCache {
	m := ""
	if len(model) > 0 {
		m = model[0]
	}
	return &MySQLEmbeddingsCache{db: mysqlDB, model: m}
}

func (ec *MySQLEmbeddingsCache) Store(noteID string, embedding []float32) error {
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	now := time.Now()
	dims := len(embedding)

	_, err = ec.db.db.Exec(`
		INSERT INTO embeddings_cache (note_id, model, dimensions, embedding_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			model = VALUES(model),
			dimensions = VALUES(dimensions),
			embedding_json = VALUES(embedding_json),
			updated_at = VALUES(updated_at)
	`, noteID, ec.model, dims, embeddingJSON, now, now)

	if err != nil {
		return fmt.Errorf("failed to store embedding: %w", err)
	}

	log.Printf("✓ Cached embedding for note: %s (model: %s, size: %d dimensions)", noteID, ec.model, dims)
	return nil
}

func (ec *MySQLEmbeddingsCache) Get(noteID string) ([]float32, bool) {
	var embeddingJSON string
	var cachedModel string
	var cachedDims int

	err := ec.db.db.QueryRow(`
		SELECT embedding_json, model, dimensions FROM embeddings_cache
		WHERE note_id = ?
	`, noteID).Scan(&embeddingJSON, &cachedModel, &cachedDims)

	if err == sql.ErrNoRows {
		log.Printf("⚠️  Cache miss for note: %s", noteID)
		return nil, false
	}

	if err != nil {
		log.Printf("❌ Error retrieving embedding for note %s: %v", noteID, err)
		return nil, false
	}

	if ec.model != "" && cachedModel != ec.model {
		log.Printf("⚠️  Cache miss for note %s: cached model %q != configured model %q", noteID, cachedModel, ec.model)
		return nil, false
	}

	var embedding []float32
	if err := json.Unmarshal([]byte(embeddingJSON), &embedding); err != nil {
		log.Printf("❌ Error unmarshaling embedding for note %s: %v", noteID, err)
		return nil, false
	}

	if cachedDims > 0 && len(embedding) != cachedDims {
		log.Printf("⚠️  Cache miss for note %s: dimension mismatch (%d != %d)", noteID, len(embedding), cachedDims)
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
