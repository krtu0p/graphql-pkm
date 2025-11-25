package service

import "graphql-pkm/internal/models"

type SearchService struct {
	noteService *NoteService
	aiService *AIService
}

func NewSearchService(noteService *NoteService, aiService *AIService) *SearchService {
	return &SearchService{
		noteService: noteService,
		aiService: aiService,
	}
}

func (s *SearchService) SearchNotes(query string) ([]*models.Note, error) {
	return s.noteService.SearchNotes(query)
}

func (s *SearchService) SemanticSearch(query string) ([]*models.SearchResult, error) {
	return s.aiService.SemanticSearch(query)
}

func (s *SearchService) SmartSearch(query string) (*models.SmartSearchResponse, error) {
	return s.aiService.SmartSearch(query)
}

func (s *SearchService) HybridSearch(query string) ([]*models.SearchResult, error) {
	return s.aiService.HybridSearch(query)
}