package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type ClientConfig struct {
	Enabled      bool
	BaseURL      string
	APIKey       string
	Model        string
	EmbedBaseURL string
	EmbedAPIKey  string
	EmbedModel   string
	Timeout      time.Duration
}

type Client struct {
	enabled      bool
	apiKey       string
	baseURL      string
	httpClient   *http.Client
	model        string
	embedBaseURL string
	embedAPIKey  string
	embedModel   string
}

func NewClient(apiKey, baseURL, model, jinaToken string) *Client {
	embedBase := baseURL
	embedKey := apiKey
	if jinaToken != "" {
		embedKey = jinaToken
	}
	return NewClientWithConfig(ClientConfig{
		Enabled:      baseURL != "",
		BaseURL:      baseURL,
		APIKey:       apiKey,
		Model:        model,
		EmbedBaseURL: embedBase,
		EmbedAPIKey:  embedKey,
		EmbedModel:   "nomic-embed-text",
	})
}

func NewClientWithConfig(cfg ClientConfig) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	embedBaseURL := strings.TrimRight(strings.TrimSpace(cfg.EmbedBaseURL), "/")
	if embedBaseURL == "" {
		embedBaseURL = baseURL
	}
	embedAPIKey := cfg.EmbedAPIKey
	if embedAPIKey == "" {
		embedAPIKey = cfg.APIKey
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "llama3.2"
	}
	embedModel := strings.TrimSpace(cfg.EmbedModel)
	if embedModel == "" {
		embedModel = "nomic-embed-text"
	}

	return &Client{
		enabled:      cfg.Enabled && baseURL != "",
		apiKey:       cfg.APIKey,
		baseURL:      baseURL,
		model:        model,
		embedBaseURL: embedBaseURL,
		embedAPIKey:  embedAPIKey,
		embedModel:   embedModel,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) IsEnabled() bool {
	return c != nil && c.enabled && c.baseURL != ""
}

func (c *Client) SetModel(model string) {
	c.model = model
}

func (c *Client) EmbedModel() string {
	return c.embedModel
}

func (c *Client) makeRequest(endpoint string, payload interface{}) ([]byte, error) {
	return c.makeRequestTo(c.baseURL, c.apiKey, endpoint, payload)
}

func (c *Client) makeRequestTo(baseURL, apiKey, endpoint string, payload interface{}) ([]byte, error) {
	if !c.IsEnabled() {
		return nil, fmt.Errorf("AI client is not enabled")
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
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
	reqPayload := openrouterEmbbedingRequest{
		Model: c.embedModel,
		Input: []string{text},
	}

	body, err := c.makeRequestTo(c.embedBaseURL, c.embedAPIKey, "/embeddings", reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to call embeddings API: %w", err)
	}

	var respObj openrouterEmbbedingResponse
	if err := json.Unmarshal(body, &respObj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
	}
	if len(respObj.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
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

	lowTemp := 0.2
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
		Stream:      false,
		MaxTokens:   150,
		Temperature: &lowTemp,
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

	lowTemp := 0.2
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
		Stream:      false,
		MaxTokens:   100,
		Temperature: &lowTemp,
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

	raw := cleanAIContent(response.Choices[0].Message.Content)

	var concepts []string
	if err := json.Unmarshal([]byte(raw), &concepts); err != nil {
		return extractConceptsFromText(raw), nil
	}

	return concepts, nil
}

func extractConceptsFromText(text string) []string {
	if start := strings.Index(text, "["); start != -1 {
		if end := strings.LastIndex(text, "]"); end > start {
			text = text[start+1 : end]
		}
	}
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

func (c *Client) Ask(prompt string) (string, error) {
	if !c.IsEnabled() {
		return "", fmt.Errorf("AI client not enabled")
	}

	lowTemp := 0.2
	request := openRouterChatRequest{
		Model: c.model,
		Messages: []openRouterChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream:      false,
		Temperature: &lowTemp,
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
