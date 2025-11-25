package resolvers

import (
    "context"
    "graphql-pkm/internal/models"
)

func (r *queryResolver) SemanticSearch(ctx context.Context, query string) ([]*models.SearchResult, error) {
    if query == "" {
        return []*models.SearchResult{}, nil
    }
    return r.searchService.SemanticSearch(query)
}

func (r *queryResolver) SmartSearch(ctx context.Context, query string) (*models.SmartSearchResponse, error) {
    if query == "" {
        return &models.SmartSearchResponse{
            Results:     []*models.SearchResult{},
            Explanation: "Empty query",
        }, nil
    }
    return r.searchService.SmartSearch(query)
}

func (r *queryResolver) HybridSearch(ctx context.Context, query string) ([]*models.SearchResult, error) {
    if query == "" {
        return []*models.SearchResult{}, nil
    }
    return r.searchService.HybridSearch(query)
}