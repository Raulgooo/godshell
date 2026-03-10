package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client interface {
	Chat(messages []Message, tools []Tool) (*Message, error)
}

type OpenAIClient struct {
	APIKey     string
	Endpoint   string
	Model      string
	HTTPClient *http.Client
}

func NewOpenAIClient(apiKey, endpoint, model string) *OpenAIClient {
	return &OpenAIClient{
		APIKey:     apiKey,
		Endpoint:   endpoint,
		Model:      model,
		HTTPClient: &http.Client{},
	}
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *OpenAIClient) Chat(messages []Message, tools []Tool) (*Message, error) {
	reqBody := chatRequest{
		Model:    c.Model,
		Messages: messages,
		Tools:    tools,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, err
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from LLM")
	}

	return &chatResp.Choices[0].Message, nil
}
