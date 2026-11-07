package models

import "time"

type Note struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Summary     *string   `json:"summary,omitempty"`
	KeyConcepts []string  `json:"keyConcepts"`
	Links       []*Link   `json:"links"`
	Backlinks   []*Link   `json:"backlinks"`
}

type Link struct {
	ID           string    `json:"id"`
	SourceNoteID string    `json:"sourceNoteId"`
	TargetNoteID string    `json:"targetNoteId"`
	Description  *string   `json:"description,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Backlink struct {
	ID          string    `json:"id"`
	SourceNote  *Note     `json:"sourceNote"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type CreateNoteInput struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags,omitempty"`
}

type UpdateNoteInput struct {
	Title   *string  `json:"title,omitempty"`
	Content *string  `json:"content,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

type SearchResult struct {
	Note      *Note   `json:"note"`
	Score     float64 `json:"score"`
	Reason    string  `json:"reason"`
	MatchType string  `json:"matchType"`
}

type SmartSearchResponse struct {
	Results     []*SearchResult `json:"results"`
	Explanation string          `json:"explanation"`
	Gaps        []string        `json:"gaps"`
	Connections []string        `json:"connections"`
}

type AISettings struct {
	Enabled      bool   `json:"enabled"`
	Provider     string `json:"provider"`
	BaseURL      string `json:"baseUrl" gqlgen:"baseUrl"`
	Model        string `json:"model"`
	HasAPIKey    bool   `json:"hasApiKey" gqlgen:"hasApiKey"`
	EmbedBaseURL string `json:"embedBaseUrl" gqlgen:"embedBaseUrl"`
	EmbedModel   string `json:"embedModel"`
}

type SaveAISettingsInput struct {
	Enabled      bool    `json:"enabled"`
	Provider     string  `json:"provider"`
	BaseURL      string  `json:"baseUrl" gqlgen:"baseUrl"`
	Model        string  `json:"model"`
	APIKey       *string `json:"apiKey,omitempty" gqlgen:"apiKey"`
	EmbedBaseURL string  `json:"embedBaseUrl" gqlgen:"embedBaseUrl"`
	EmbedModel   string  `json:"embedModel"`
}

type ListAIModelsInput struct {
	Provider *string `json:"provider,omitempty"`
	BaseURL  string  `json:"baseUrl" gqlgen:"baseUrl"`
	APIKey   *string `json:"apiKey,omitempty" gqlgen:"apiKey"`
}

type AIModelList struct {
	Models []string `json:"models"`
	Error  *string  `json:"error,omitempty"`
}
