package llm

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

type Client struct {
	baseURL   string
	apiKey    string
	model     string
	maxTokens int32
	http      *http.Client
}

func NewClient(baseURL, apiKey, model string, maxTokens int32, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		http:      &http.Client{Timeout: timeout},
	}
}

func (c *Client) Enabled() bool {
	return c.apiKey != "" && c.baseURL != "" && c.model != ""
}

func (c *Client) Model() string { return c.model }

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string    `json:"model"`
	Messages       []message `json:"messages"`
	Temperature    float64   `json:"temperature"`
	MaxTokens      int32     `json:"max_tokens,omitempty"`
	ResponseFormat any       `json:"response_format,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int32 `json:"prompt_tokens"`
		CompletionTokens int32 `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

type ChatResult struct {
	Content   string
	TokensIn  int32
	TokensOut int32
}

func (c *Client) ChatJSON(ctx context.Context, system, user string) (*ChatResult, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("llm client is not configured")
	}

	body := chatRequest{
		Model: c.model,
		Messages: []message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature:    0,
		MaxTokens:      c.maxTokens,
		ResponseFormat: map[string]string{"type": "json_object"},
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var parsed chatResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal response (status=%d): %w", resp.StatusCode, err)
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("llm error (status=%d): %s", resp.StatusCode, parsed.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("llm http status %d: %s", resp.StatusCode, truncate(string(payload), 200))
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("llm returned no choices")
	}

	return &ChatResult{
		Content:   stripCodeFence(parsed.Choices[0].Message.Content),
		TokensIn:  parsed.Usage.PromptTokens,
		TokensOut: parsed.Usage.CompletionTokens,
	}, nil
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
