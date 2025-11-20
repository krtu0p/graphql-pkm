package ai

import (
	"fmt"
	"encoding/json"
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
	return nil, nil
}

func (c *Client) extractConcepts(prompt string) ([]string, error) {
	return nil, nil
}

func (c *Client) conceptsToEmbedding(concepts []string) []float32 {
	return nil
}