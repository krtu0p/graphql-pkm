package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
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

type clientSettings struct {
	enabled      bool
	apiKey       string
	baseURL      string
	model        string
	embedBaseURL string
	embedAPIKey  string
	embedModel   string
}

type Client struct {
	mu         sync.RWMutex
	settings   clientSettings
	httpClient *http.Client
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

	return &Client{
		settings:   normalizeClientSettings(cfg),
		httpClient: &http.Client{Timeout: timeout},
	}
}

// UpdateConfig replaces the client configuration at runtime. Shared by all
// services that hold this client, so new settings apply to in-flight workers.
func (c *Client) UpdateConfig(cfg ClientConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.settings = normalizeClientSettings(cfg)
	if cfg.Timeout > 0 {
		c.httpClient = &http.Client{Timeout: cfg.Timeout}
	}
}

func normalizeClientSettings(cfg ClientConfig) clientSettings {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	embedBaseURL := strings.TrimRight(strings.TrimSpace(cfg.EmbedBaseURL), "/")
	if embedBaseURL == "" {
		embedBaseURL = baseURL
	}
	embedAPIKey := strings.TrimSpace(cfg.EmbedAPIKey)
	apiKey := strings.TrimSpace(cfg.APIKey)
	if embedAPIKey == "" {
		embedAPIKey = apiKey
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "llama3.2"
	}
	embedModel := strings.TrimSpace(cfg.EmbedModel)
	if embedModel == "" {
		embedModel = "nomic-embed-text"
	}

	return clientSettings{
		enabled:      cfg.Enabled && baseURL != "",
		apiKey:       apiKey,
		baseURL:      baseURL,
		model:        model,
		embedBaseURL: embedBaseURL,
		embedAPIKey:  embedAPIKey,
		embedModel:   embedModel,
	}
}

func (c *Client) snapshot() (clientSettings, *http.Client) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.settings, c.httpClient
}

func (c *Client) IsEnabled() bool {
	if c == nil {
		return false
	}
	settings, _ := c.snapshot()
	return settings.enabled && settings.baseURL != ""
}

func (c *Client) SetModel(model string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.settings.model = model
}

func (c *Client) EmbedModel() string {
	settings, _ := c.snapshot()
	return settings.embedModel
}

func (c *Client) modelName() string {
	settings, _ := c.snapshot()
	return settings.model
}

func (c *Client) makeRequest(endpoint string, payload interface{}) ([]byte, error) {
	settings, httpClient := c.snapshot()
	if !settings.enabled || settings.baseURL == "" {
		return nil, fmt.Errorf("AI client is not enabled")
	}
	return c.doRequest(httpClient, http.MethodPost, settings.baseURL, settings.apiKey, endpoint, payload)
}

func (c *Client) makeRequestTo(baseURL, apiKey, endpoint string, payload interface{}) ([]byte, error) {
	settings, httpClient := c.snapshot()
	if !settings.enabled || settings.baseURL == "" {
		return nil, fmt.Errorf("AI client is not enabled")
	}
	return c.doRequest(httpClient, http.MethodPost, baseURL, apiKey, endpoint, payload)
}

func (c *Client) doRequest(httpClient *http.Client, method, baseURL, apiKey, endpoint string, payload interface{}) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("HTTP-Referer", "https://github.com/krtu0p/graphql-pkm")
	req.Header.Set("X-Title", "GraphQL PKM")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("API Returned: %d, %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return io.ReadAll(resp.Body)
}

// ListModels calls {baseURL}/models with the supplied values. The values are
// not persisted; the provided client is only used for its HTTP transport.
func (c *Client) ListModels(baseURL, apiKey string) ([]string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	_, httpClient := c.snapshot()
	body, err := c.doRequest(httpClient, http.MethodGet, baseURL, strings.TrimSpace(apiKey), "/models", nil)
	if err != nil {
		return nil, err
	}

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Models []struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse models response: %w", err)
	}

	seen := make(map[string]bool)
	ids := make([]string, 0, len(payload.Data)+len(payload.Models))
	for _, model := range payload.Data {
		if model.ID != "" && !seen[model.ID] {
			seen[model.ID] = true
			ids = append(ids, model.ID)
		}
	}
	for _, model := range payload.Models {
		id := model.ID
		if id == "" {
			id = model.Name
		}
		if id != "" && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}

	sort.Strings(ids)
	return ids, nil
}

func (c *Client) GetEmbedding(text string) ([]float32, error) {
	settings, _ := c.snapshot()
	reqPayload := openrouterEmbbedingRequest{
		Model: settings.embedModel,
		Input: []string{text},
	}

	body, err := c.makeRequestTo(settings.embedBaseURL, settings.embedAPIKey, "/embeddings", reqPayload)
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

// noteLanguageRule tells AI text fields to follow the note language instead
// of the prompt language.
const noteLanguageRule = "Write your answer in the language of the note contents (the body text, not the title and not these instructions). Do not translate or mix languages."

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

Provide only the summary, no extra text.

%s`, title, content, noteLanguageRule)

	lowTemp := 0.2
	request := openRouterChatRequest{
		Model: c.modelName(),
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

Return only a JSON array of strings, like: ["concept1", "concept2", "concept3"]

%s
Keep the JSON array format exactly as specified.`, title, content, noteLanguageRule)

	lowTemp := 0.2
	request := openRouterChatRequest{
		Model: c.modelName(),
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
	return c.AskWithMaxTokens(prompt, 0)
}

// AskWithMaxTokens sends a user prompt and caps the completion length.
// maxTokens of 0 omits max_tokens from the request so the server decides.
// Some reasoning models spend tokens on hidden reasoning before the visible
// answer, so callers showing text to the user pass a generous budget.
func (c *Client) AskWithMaxTokens(prompt string, maxTokens int) (string, error) {
	if !c.IsEnabled() {
		return "", fmt.Errorf("AI client not enabled")
	}

	lowTemp := 0.2
	request := openRouterChatRequest{
		Model: c.modelName(),
		Messages: []openRouterChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream:      false,
		MaxTokens:   maxTokens,
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
