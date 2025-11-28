package database

type EmbeddingsCache interface {
	Store(noteID string, embedding []float32) error
	Get(noteID string) ([]float32, bool)
	Delete(noteID string) error
	Clear() error
	Count() (int, error)
	GetCacheStats() (map[string]interface{}, error)
}

var _ EmbeddingsCache = (*MySQLEmbeddingsCache)(nil)