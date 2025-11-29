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
<<<<<<< HEAD
	hfToken string
}

func NewClient(apiKey, baseURL, model, hfToken string) *Client {
=======
	jinaToken string
}

func NewClient(apiKey, baseURL, model, jinaToken string) *Client {
>>>>>>> 3bca154 (+semantic search fix and refine with jina)
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
<<<<<<< HEAD
		hfToken: hfToken,
=======
		jinaToken: jinaToken,
>>>>>>> 3bca154 (+semantic search fix and refine with jina)
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

<<<<<<< HEAD
// Add these methods to the END of your client.go file


// GetEmbedding generates a semantic embedding using Hugging Face (free)
=======

>>>>>>> 3bca154 (+semantic search fix and refine with jina)
func (c *Client) GetEmbedding(text string) ([]float32, error) {
    url := "https://api.jina.ai/v1/embeddings"
    payload := map[string]interface{}{
        "input": []string{text},
<<<<<<< HEAD
        "model": "jina-embeddings-v3",       // bisa diganti model lain
        "dimensions": 1024,                  // sesuai model
=======
        "model": "jina-embeddings-v3",      
        "dimensions": 1024,                 
>>>>>>> 3bca154 (+semantic search fix and refine with jina)
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
<<<<<<< HEAD
    req.Header.Set("Authorization", "Bearer "+c.hfToken)  // atau c.apiKey kalau itu token Jina
=======
    req.Header.Set("Authorization", "Bearer "+c.jinaToken)  
>>>>>>> 3bca154 (+semantic search fix and refine with jina)

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

<<<<<<< HEAD
    // response JSON bentuk:
    // { "data": [ { "embedding": [float,...] } ] }
=======
>>>>>>> 3bca154 (+semantic search fix and refine with jina)
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

<<<<<<< HEAD


// GenerateEmbedding converts float32 to float64 for compatibility
=======
>>>>>>> 3bca154 (+semantic search fix and refine with jina)
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

