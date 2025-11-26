package database

import (
	"encoding/json"
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
		return err
	}
	
	now := time.Now()
	
	_, err = ec.db.db.Exec(`
		INSERT INTO embeddings_cache (note_id, embedding_json, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			embedding_json = VALUES(embedding_json),
			updated_at = VALUES(updated_at)
	`, noteID, embeddingJSON, now, now)
	
	return err
}

func (ec *MySQLEmbeddingsCache) Get(noteID string) ([]float32, bool) {
	var embeddingJSON string
	var embedding []float32
	
	err := ec.db.db.QueryRow(`
		SELECT embedding_json FROM embeddings_cache
		WHERE note_id = ?
	`, noteID).Scan(&embeddingJSON)
	
	if err != nil {
		return nil, false
	}
	
	if err := json.Unmarshal([]byte(embeddingJSON), &embedding); err != nil {
		return nil, false
	}
	
	return embedding, true
}

func (ec *MySQLEmbeddingsCache) Delete(noteID string) {
	ec.db.db.Exec(`
		DELETE FROM embeddings_cache WHERE note_id = ?
	`, noteID)
}