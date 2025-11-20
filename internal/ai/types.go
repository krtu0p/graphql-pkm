package ai

type openrouterEmbbedingRequest struct {
	Model string `json:"model"`
	Input []string `json:"input"`
}

type openrouterEmbbedingResponse struct {
	Data []struct {
		Object string `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index int `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
}

type openRouterChatRequest struct {
	Model string	`json:"model"`
	Messages []openRouterChatMessage `json:"messages"`
	Stream bool `json:"stream"`
}

type openRouterChatMessage struct {
	Role string `json:"role"`
	Content string `json:"content"`
}

type openRouterChatResponse struct {
	Choices []struct {
		Message struct {
			Role string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Model string `json:"model"`
}

type smartSearchResult struct { 
	RelevantNotes []struct {
		NoteID string `json:"noteId"`
		Reason string `json:"reason"`
		Score float64 `json:"score"`
	} `json:"relevantNotes"`
	Explanation string `json:"explanation"`
	Gaps []string `json:"gaps"`
	Connections []string `json:"connections"`	
}