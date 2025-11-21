package ai

import (
	"encoding/json"
	"fmt"
)

func (c *Client) getEmbbeding(text string) ([]float32, error) {
	if c.supportsEmbeddings() {
		return c.getDirectEmbedding(text)
	}
	
	return c.getSemanticEmbedding(text)
}

func (c *Client) supportsEmbeddings() bool {
	return c.model == "text-embedding-ada-002"
}

func (c *Client) getDirectEmbedding(text string) ([]float32, error) {
	request := openrouterEmbbedingRequest{
		Model: c.model,
		Input: []string{text},
	}
	
	body, err := c.makeRequest("/embeddings", request)
	if err != nil {
		return nil, err
	}
	
	var response openrouterEmbbedingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("Failed to Parse embedding response: %w", err)
	}
	
	if len(response.Data) == 0 {
		return nil, fmt.Errorf("No embedding data found in response")
	}
	
	return response.Data[0].Embedding, nil
}

func (c *Client) getSemanticEmbedding(text string) ([]float32, error) {
	prompt := fmt.Sprintf(`Convert The following text into a semantic representation for search.
		Return ONLY a json array of 10-15 key concepts of topics from this text: %s`, text)
	
	concepts, err := c.extractConcepts(prompt)
	if err != nil {
		return nil, err
	}
	
	return c.conceptsToEmbedding(concepts), nil
}

func (c *Client) extractConcepts(prompt string) ([]string, error) {
	
	request := openRouterChatRequest{
		Model: c.model,
		Messages: []openRouterChatMessage{
			{
				Role: "system",
				Content: "You are a semantic analysis tool. Extract key concepts and return as JSON array",
			},
			{
				Role: "user",
				Content: prompt, 
			},
			
		},
		Stream: false,
	}
	
	body, err := c.makeRequest("/chat/completions", request)
	if err != nil {
		return nil, err
	}
	
	var response openRouterChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse chat response: %w", err)
	}
	
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}
	
	var concepts []string
	content := response.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(content),&concepts); err != nil {
		return []string{content}, nil
	}
	
	return concepts, nil
}

func (c *Client) conceptsToEmbedding(concepts []string) []float32 {
	embedding := make([]float32, 384) // 384 = standard embedded size
	for i, concept := range concepts {
		if i >= len(embedding) {
			break
		}
		
		embedding[i] = float32(len(concept)) / 100.0
	}
	return embedding
}