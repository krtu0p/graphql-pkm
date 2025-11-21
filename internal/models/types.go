package models

import (
	"time"
)


type Note struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Tags      []string `json:"tags"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Summary   string `json:"summary"`
	KeyConcepts []string `json:"keyconcepts"`
}

type Link struct {
	ID        string `json:"id"`
	SourceNoteID string `json:"sourceNoteId"`
	TargetNoteID string `json:"targetNoteId"`
	Description string `json:"description"`
	CreatedAt time.Time `json:"createdAt"`
}

type backLink struct {
	ID        string `json:"id"`
	SourceNoteID string `json:"sourceNoteId"`
	TargetNoteID string `json:"targetNoteId"`
	Description string `json:"description"`
	CreatedAt time.Time `json:"createdAt"`
}

type searchResult struct {
	NoteID		*Note `json:"note"`
	Score		float64 `json:"score"`
	Reason		string `json:"reason"`
	MatchType	string `json:"matchType"`
}

type smartSearchResponse struct {
	Results		[]searchResult `json:"results"`
	Explanation	string `json:"explanation"`
	Gaps		[]string `json:"gaps"`
	Connections	[]string `json:"connections"`
}

type createNoteInput struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	Tags      []string `json:"tags"`
}

type updateNoteInput struct {
	Title     *string `json:"title"`
	Content   *string `json:"content"`
	Tags      []string `json:"tags"`
}

