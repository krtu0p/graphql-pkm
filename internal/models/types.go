package models

import (
    "time"
)

type Note struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Content     string    `json:"content"`
    Tags        []string  `json:"tags"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
    Summary     *string   `json:"summary,omitempty"`
    KeyConcepts []string  `json:"keyConcepts"`
}

type Link struct {
    ID           string    `json:"id"`
    SourceNoteID string    `json:"sourceNoteId"`
    TargetNoteID string    `json:"targetNoteId"` // FIXED: lowercase 'd'
    Description  *string   `json:"description,omitempty"`
    CreatedAt    time.Time `json:"createdAt"`
}

type Backlink struct { // FIXED: lowercase 'l'
    ID          string    `json:"id"`
    SourceNote  *Note     `json:"sourceNote"`
    Description *string   `json:"description,omitempty"`
    CreatedAt   time.Time `json:"createdAt"`
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
    Connections []string        `json:"connections"` // FIXED: plural
}

type CreateNoteInput struct {
    Title   string   `json:"title"`
    Content string   `json:"content"`
    Tags    []string `json:"tags"`
}

type UpdateNoteInput struct {
    Title   *string  `json:"title,omitempty"`
    Content *string  `json:"content,omitempty"`
    Tags    []string `json:"tags,omitempty"`
}