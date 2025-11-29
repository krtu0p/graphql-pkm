package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"regexp"
	"strings"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	model      string
	jinaToken string
}

func NewClient(apiKey, baseURL, model, jinaToken string) *Client {

	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		jinaToken: jinaToken,

	}
}

func (c *Client) IsEnabled() bool {
	return c.apiKey != ""
}

func (c *Client) SetModel(model string) {
	c.model = model
}

func (c *Client) makeRequest(endpoint string, payload interface{}) ([]byte, error) {
	if !c.IsEnabled() {
		return nil, fmt.Errorf("AI client is not enabled")
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/krtu0p/graphql-pkm")
	req.Header.Set("X-Title", "GraphQL PKM")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API Returned: %d, %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}


func (c *Client) GetEmbedding(text string) ([]float32, error) {
    url := "https://api.jina.ai/v1/embeddings"
    payload := map[string]interface{}{
        "input": []string{text},
        "model": "jina-embeddings-v3",      
        "dimensions": 1024,                 

    }

    jsonData, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.jinaToken)  

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to call Jina API: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read Jina response: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("Jina API error %d: %s", resp.StatusCode, string(body))
    }

    var respObj struct {
        Data []struct {
            Embedding []float32 `json:"embedding"`
        } `json:"data"`
    }
    if err := json.Unmarshal(body, &respObj); err != nil {
        return nil, fmt.Errorf("failed to unmarshal Jina embedding: %w", err)
    }
    if len(respObj.Data) == 0 {
        return nil, fmt.Errorf("no embedding returned from Jina")
    }
    return respObj.Data[0].Embedding, nil
}

func (c *Client) GenerateEmbedding(text string) ([]float64, error) {
	embed32, err := c.GetEmbedding(text)
	if err != nil {
		return nil, err
	}
	
	embed64 := make([]float64, len(embed32))
	for i, v := range embed32 {
		embed64[i] = float64(v)
	}
	return embed64, nil
}

func cleanAIContent(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "```json", "")
	s = strings.ReplaceAll(s, "```JSON", "")
	s = strings.ReplaceAll(s, "```", "")
	return strings.TrimSpace(s)
}

func extractJSON(s string) string {
	re := regexp.MustCompile(`\{[\s\S]*\}`)
	return re.FindString(s)
}

func formatNotesForAI(notes map[string]string) string {
	var sb strings.Builder
	for id, content := range notes {
		if len(content) > 500 {
			content = content[:500] + "..."
		}
		sb.WriteString(fmt.Sprintf("Note %s: %s\n\n", id, content))
	}
	return sb.String()
}

func (c *Client) GenerateSummary(title, content string) (string, error) {
	if !c.IsEnabled() {
		return "", nil 
	}

	prompt := fmt.Sprintf(`Summarize this note in 1-2 sentences:

Title: %s
Content: %s

Provide only the summary, no extra text.`, title, content)

	request := openRouterChatRequest{
		Model: c.model,
		Messages: []openRouterChatMessage{
			{
				Role:    "system",
				Content: "You are a concise summarization assistant. Provide brief, accurate summaries.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream:    false,
		MaxTokens: 150,
	}

	body, err := c.makeRequest("/chat/completions", request)
	if err != nil {
		return "", err
	}

	var response openRouterChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}

func (c *Client) ExtractKeyConcepts(title, content string) ([]string, error) {
	if !c.IsEnabled() {
		return []string{}, nil 
	}

	prompt := fmt.Sprintf(`Extract 3-5 key concepts or topics from this note:

Title: %s
Content: %s

Return only a JSON array of strings, like: ["concept1", "concept2", "concept3"]`, title, content)

	request := openRouterChatRequest{
		Model: c.model,
		Messages: []openRouterChatMessage{
			{
				Role:    "system",
				Content: "You are a concept extraction assistant. Return only valid JSON arrays.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream:    false,
		MaxTokens: 100,
	}

	body, err := c.makeRequest("/chat/completions", request)
	if err != nil {
		return []string{}, err
	}

	var response openRouterChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return []string{}, err
	}

	if len(response.Choices) == 0 {
		return []string{}, fmt.Errorf("no response from AI")
	}

	
	var concepts []string
	if err := json.Unmarshal([]byte(content), &concepts); err != nil {
		return extractConceptsFromText(content), nil
	}

	return concepts, nil
}

func extractConceptsFromText(text string) []string {
	text = strings.Trim(text, "[]")
	text = strings.ReplaceAll(text, "\"", "")
	parts := strings.Split(text, ",")
	
	var concepts []string
	for _, part := range parts {
		concept := strings.TrimSpace(part)
		if concept != "" {
			concepts = append(concepts, concept)
		}
	}
	return concepts
}