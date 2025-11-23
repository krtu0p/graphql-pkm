package database

import "sync"

type EmbeddingsCache struct {
	embeddings map[string][]float32
	mu		   sync.RWMutex
}

func newEmbeddingsCache() *EmbeddingsCache {
	return &EmbeddingsCache{
		embeddings: make(map[string][]float32),
	}
}

func (ec *EmbeddingsCache) Store(noteID string, embedding []float32) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	ec.embeddings[noteID] = embedding
}

func (ec *EmbeddingsCache) Get(noteID string) ([]float32, bool) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	
	embedding, exists := ec.embeddings[noteID]
	
	return  embedding, exists
}

func (ec *EmbeddingsCache) Delete(noteID string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	delete(ec.embeddings, noteID)
}