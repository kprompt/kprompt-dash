package amazonbedrock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is an Amazon Bedrock API client compatible with the Enterprise Deferred provider interface.
type Client struct {
	Endpoint   string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new Amazon Bedrock client with the given endpoint and API key.
func NewClient(endpoint, apiKey string) *Client {
	return &Client{
		Endpoint:   strings.TrimRight(endpoint, "/"),
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// CompletionRequest represents a request to the Bedrock completion API.
type CompletionRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature float64  `json:"temperature,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

// CompletionResponse represents a response from the Bedrock completion API.
type CompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// Complete sends a completion request to Amazon Bedrock and returns the generated text.
func (c *Client) Complete(ctx context.Context, req CompletionRequest) (string, error) {
	url := fmt.Sprintf("%s/completions", c.Endpoint)

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal completion request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request to Bedrock failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Bedrock API error: status %d, body: %s", resp.StatusCode, string(b))
	}

	var completionResp CompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completionResp); err != nil {
		return "", fmt.Errorf("failed to decode Bedrock response: %w", err)
	}

	if len(completionResp.Choices) == 0 {
		return "", errors.New("no choices returned from Bedrock")
	}

	return completionResp.Choices[0].Text, nil
}
