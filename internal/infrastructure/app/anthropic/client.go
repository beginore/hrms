package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	apiURL        = "https://api.anthropic.com/v1/messages"
	apiVersion    = "2023-06-01"
	defaultModel  = "claude-haiku-4-5-20251001"
	defaultMaxOut = 1024
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     strings.TrimSpace(apiKey),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ─── Request / Response types ─────────────────────────────────────────────────

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
}

type response struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ─── Complete ─────────────────────────────────────────────────────────────────

func (c *Client) Complete(ctx context.Context, system, userPrompt string) (string, error) {
	body, err := json.Marshal(request{
		Model:     defaultModel,
		MaxTokens: defaultMaxOut,
		System:    system,
		Messages:  []Message{{Role: "user", Content: userPrompt}},
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var result response
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("anthropic error: %s", result.Error.Message)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from anthropic")
	}
	return result.Content[0].Text, nil
}
