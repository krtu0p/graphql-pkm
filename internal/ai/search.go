package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)


func (c *Client) smartSearch(query string, contextNotes map[string]string) (*smartSearchResult, error){
	if !c.isEnabled() {
		return nil, fmt.Errorf("AI client is not enabled")
	}
	
	contextStr := formatNotesForAI(contextNotes)
	
	prompt := fmt.Sprintf(`
    USER QUERY: "%s"

    AVAILABLE NOTES:
    %s

    Analyze this query against the available notes and provide:
    1. Which notes are most relevant (include note IDs and specific reasons)
    2. An overall explanation of the relevance
    3. Any apparent gaps in the knowledge base
    4. Suggested connections between ideas

    Respond in this exact JSON format:
    {
        "relevantNotes": [
            {
                "noteId": "note_id_here",
                "reason": "Why this note is relevant",
                "score": 0.95
            }
        ],
        "explanation": "Overall analysis...",
        "gaps": ["gap1", "gap2"],
        "connections": ["connection1", "connection2"]
    }
    `, query, contextStr)
	
	request := openRouterChatRequest{
		Model: c.model,
		Messages: []openRouterChatMessage{
			{
				Role: "system",
				Content: "You are a knowledge managament assistant. Always response with valid JSON",
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
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}
	
	content := response.Choices[0].Message.Content
	
	content = strings.TrimSpace(content)
	content = strings.ReplaceAll(content, "```json", "")
	content = strings.ReplaceAll(content, "```JSON", "")
	content = strings.ReplaceAll(content, "```", "")
	content = strings.TrimSpace(content)
	
	var result smartSearchResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("error parsing ai json response")
	}
	return &result, nil
}

func formatNotesForAI(notes map[string]string) string {
	var sb strings.Builder
	for id, content := range notes {
		if len(content) > 500 {
			content = content[:500] + "..."
		}
		sb.WriteString(fmt.Sprintf("Notes %s: %s\n", id, content))
	}
	return sb.String()
}