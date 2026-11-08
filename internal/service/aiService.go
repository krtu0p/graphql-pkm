package service

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"

	"graphql-pkm/internal/ai"
	"graphql-pkm/internal/database"
	"graphql-pkm/internal/models"
)

type AIService struct {
	aiClient          *ai.Client
	embeddingsCache   database.EmbeddingsCache
	noteService       *NoteService
	semanticThreshold float64
}

func NewAIService(aiClient *ai.Client, cache database.EmbeddingsCache, noteService *NoteService) *AIService {
	return &AIService{
		aiClient:          aiClient,
		embeddingsCache:   cache,
		noteService:       noteService,
		semanticThreshold: 0.6,
	}
}

func (s *AIService) SetSemanticThreshold(threshold float64) {
	if threshold > 0 {
		s.semanticThreshold = threshold
	}
}

func (s *AIService) SemanticSearch(query string) ([]*models.SearchResult, error) {
	if s.aiClient == nil || !s.aiClient.IsEnabled() {
		return s.fallbackSearch(query)
	}

	queryEmbedding, err := s.aiClient.GetEmbedding(query)
	if err != nil || len(queryEmbedding) == 0 {
		log.Printf("Warning: Semantic search falling back to keyword search (embedding unavailable: %v)", err)
		return s.fallbackSearch(query)
	}

	allNotes, err := s.noteService.GetAllNotes()
	if err != nil {
		return nil, err
	}

	threshold := s.semanticThreshold
	if threshold <= 0 {
		threshold = 0.6
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
			_ = s.embeddingsCache.Store(note.ID, embedding)
		}

		if len(embedding) != len(queryEmbedding) {
			continue
		}

		similarity := cosineSimilarity(queryEmbedding, embedding)
		if similarity > threshold {
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
	if s.aiClient == nil || !s.aiClient.IsEnabled() {
		fallback, _ := s.fallbackSearch(query)
		return &models.SmartSearchResponse{
			Results:     fallback,
			Explanation: "AI features are disabled; showing keyword search results.",
			Gaps:        []string{},
			Connections: []string{},
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
		log.Printf("Warning: SmartSearch falling back to keyword search (AI error: %v)", err)
		fallback, _ := s.fallbackSearch(query)
		return &models.SmartSearchResponse{
			Results:     fallback,
			Explanation: "AI unavailable; showing keyword search results.",
			Gaps:        []string{},
			Connections: []string{},
		}, nil
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
		if result == nil || result.Note == nil {
			continue
		}
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

func (s *AIService) AskAI(prompt string) (string, error) {
	if s.aiClient == nil || !s.aiClient.IsEnabled() {
		return "", fmt.Errorf("AI client not enabled")
	}
	return s.aiClient.Ask(prompt)
}

// noRelationshipToken is the marker the relationship prompt asks the model to
// emit as its first line when two notes are unrelated. It is stripped from the
// answer and must never reach the UI.
const noRelationshipToken = "NO_RELATIONSHIP"

// noRelationshipFallback is returned when the AI answer contains nothing but
// the marker line.
const noRelationshipFallback = "No meaningful relationship was found between these notes."

// relationshipMaxTokens leaves room for models that spend hidden reasoning
// tokens before their visible answer.
const relationshipMaxTokens = 2048

// buildCheckRelationshipPrompt renders the AI prompt for comparing two notes.
func buildCheckRelationshipPrompt(noteA, noteB *models.Note) string {
	return fmt.Sprintf(`Decide whether these two notes from a personal knowledge base have a meaningful relationship.

Note A: %s
%s

Note B: %s
%s

Only report a relationship if the notes share the same concrete topic, if one note depends on or directly builds on the other, or if one directly applies to the other. Abstract similarities do not count as a relationship (for example: "both involve planning", "both mention preparation", "both are about saving money").

Output rules:
- Answer in plain text only: no markdown, no headings, no lists (never use "**", "#", bullet dashes, or numbered lists).
- If there is a relationship: answer in at most 3 short sentences and never use the token NO_RELATIONSHIP.
- If there is no meaningful relationship: the first line of your answer must be exactly NO_RELATIONSHIP (nothing else on that line), followed by one short sentence describing what each note is about.
- Write your answer in the language of the note contents (the body text, not the titles and not these instructions). If the two notes are in different languages, use the language of Note A. Do not translate or mix languages.`, noteA.Title, noteA.Content, noteB.Title, noteB.Content)
}

// stripNoRelationshipLine inspects the first non-empty line of an answer and,
// if it is the NO_RELATIONSHIP marker (leading/trailing whitespace and casing
// are tolerated), removes that line and returns the remaining text trimmed of
// leading blank lines. A plain relationship answer passes through unchanged.
func stripNoRelationshipLine(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.EqualFold(trimmed, noRelationshipToken) {
			return strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
		}
		break
	}
	return strings.TrimSpace(text)
}

func (s *AIService) CheckRelationship(noteA, noteB *models.Note) (string, error) {
	if noteA == nil || noteB == nil {
		return "", fmt.Errorf("both notes must exist")
	}
	if s.aiClient == nil || !s.aiClient.IsEnabled() {
		return "AI features are disabled.", nil
	}

	answer, err := s.aiClient.AskWithMaxTokens(buildCheckRelationshipPrompt(noteA, noteB), relationshipMaxTokens)
	if err != nil {
		return "", err
	}

	result := stripNoRelationshipLine(answer)
	if result == "" {
		return noRelationshipFallback, nil
	}
	return result, nil
}
