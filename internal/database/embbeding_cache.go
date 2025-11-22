package database

import "sync"

type embeddingsCache struct {
	embeddings map[string][]float32
	mu		   sync.RWMutex
}

func newEmbeddingsCache() *embeddingsCache {
	return &embeddingsCache{
		embeddings: make(map[string][]float32),
	}
}

func (ec *embeddingsCache) Store(noteID string, embedding []float32) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	ec.embeddings[noteID] = embedding
}

func (ec *embeddingsCache) Get(noteID string) ([]float32, bool) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	
	embedding, exists := ec.embeddings[noteID]
	
	return  embedding, exists
}

func (ec *embeddingsCache) Delete(noteID string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	delete(ec.embeddings, noteID)
}