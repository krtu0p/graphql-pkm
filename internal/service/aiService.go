package service

import (
    "fmt"
    "math"
    "sort"

    "graphql-pkm/internal/ai"
    "graphql-pkm/internal/database"
    "graphql-pkm/internal/models"
)

type AIService struct {
    aiClient        *ai.Client
    embeddingsCache *database.MySQLEmbeddingsCache
    noteService     *NoteService
}

func NewAIService(aiClient *ai.Client, cache *database.MySQLEmbeddingsCache, noteService *NoteService) *AIService {
    return &AIService{
        aiClient:        aiClient,
        embeddingsCache: cache,
        noteService:     noteService,
    }
}


func (s *AIService) SemanticSearch(query string) ([]*models.SearchResult, error) {
    if !s.aiClient.IsEnabled() {
        return s.fallbackSearch(query)
    }

    queryEmbedding, err := s.aiClient.GetEmbedding(query)
    if err != nil {
        return nil, err
    }
    if len(queryEmbedding) == 0 {
        return nil, fmt.Errorf("query embedding empty")
    }

    allNotes, err := s.noteService.GetAllNotes()
    if err != nil {
        return nil, err
    }

    var results []*models.SearchResult

    for _, note := range allNotes {
        embedding, exists := s.embeddingsCache.Get(note.ID)


        if !exists {
            text := note.Title + " " + note.Content
            embedding, err = s.aiClient.GetEmbedding(text)
            if err != nil || len(embedding) == 0 {
                continue
            }
            s.embeddingsCache.Store(note.ID, embedding)
        }

        if len(embedding) != len(queryEmbedding) {
            continue
        }


        similarity := cosineSimilarity(queryEmbedding, embedding)
        if similarity > 0.6 {
            results = append(results, &models.SearchResult{
                Note:      note,
                Score:     similarity,
                Reason:    "Conceptually similar",
                MatchType: "semantic",
            })
        }
    }

    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })

    return results, nil
}

func (s *AIService) SmartSearch(query string) (*models.SmartSearchResponse, error) {
    if !s.aiClient.IsEnabled() {
        return &models.SmartSearchResponse{
            Results:     []*models.SearchResult{},
            Explanation: "AI features are disabled",
        }, nil
    }

    contextNotes, _ := s.noteService.SearchNotes(query)
    if len(contextNotes) > 10 {
        contextNotes = contextNotes[:10]
    }

    contextMap := make(map[string]string)
    for _, note := range contextNotes {
        contextMap[note.ID] = note.Title + ": " + note.Content
    }

    aiResult, err := s.aiClient.SmartSearch(query, contextMap)
    if err != nil {
        return nil, err
    }

    var results []*models.SearchResult
    for _, aiNote := range aiResult.RelevantNotes {
        note, err := s.noteService.GetNote(aiNote.NoteID)
        if err == nil && note != nil {
            results = append(results, &models.SearchResult{
                Note:      note,
                Score:     aiNote.Score,
                Reason:    aiNote.Reason,
                MatchType: "ai",
            })
        }
    }

    return &models.SmartSearchResponse{
        Results:     results,
        Explanation: aiResult.Explanation,
        Gaps:        aiResult.Gaps,
        Connections: aiResult.Connections,
    }, nil
}


func (s *AIService) HybridSearch(query string) ([]*models.SearchResult, error) {
    var allResults []*models.SearchResult

    keywordNotes, _ := s.noteService.SearchNotes(query)
    for _, note := range keywordNotes {
        allResults = append(allResults, &models.SearchResult{
            Note:      note,
            Score:     0.8,
            Reason:    "Keyword match",
            MatchType: "keyword",
        })
    }

    semanticResults, _ := s.SemanticSearch(query)
    allResults = append(allResults, semanticResults...)

    return s.deduplicateResults(allResults), nil
}


func (s *AIService) deduplicateResults(results []*models.SearchResult) []*models.SearchResult {
    seen := make(map[string]bool)
    var unique []*models.SearchResult

    for _, result := range results {
        if !seen[result.Note.ID] {
            seen[result.Note.ID] = true
            unique = append(unique, result)
        }
    }

    sort.Slice(unique, func(i, j int) bool {
        return unique[i].Score > unique[j].Score
    })

    return unique
}

func (s *AIService) fallbackSearch(query string) ([]*models.SearchResult, error) {
    notes, err := s.noteService.SearchNotes(query)
    if err != nil {
        return nil, err
    }

    var results []*models.SearchResult
    for _, note := range notes {
        results = append(results, &models.SearchResult{
            Note:      note,
            Score:     1.0,
            Reason:    "Keyword match (AI unavailable)",
            MatchType: "keyword",
        })
    }

    return results, nil
}

func cosineSimilarity(a, b []float32) float64 {
    if len(a) == 0 || len(b) == 0 {
        return 0
    }
    if len(a) != len(b) {
        return 0
    }

    var dot, normA, normB float64
    for i := 0; i < len(a); i++ {
        dot += float64(a[i] * b[i])
        normA += float64(a[i] * a[i])
        normB += float64(b[i] * b[i])
    }

    if normA == 0 || normB == 0 {
        return 0
    }
    return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
