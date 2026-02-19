package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIConfig holds configuration for OpenAI client.
type OpenAIConfig struct {
	APIKey     string        `yaml:"api_key"`
	BaseURL    string        `yaml:"base_url"`
	Model      string        `yaml:"model"`
	Timeout    time.Duration `yaml:"timeout"`
	MaxRetries int           `yaml:"max_retries"`
}

// OpenAIClient implements Client for OpenAI API.
type OpenAIClient struct {
	config     OpenAIConfig
	httpClient *http.Client
}

// NewOpenAIClient creates a new OpenAI client.
func NewOpenAIClient(config OpenAIConfig) *OpenAIClient {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.Model == "" {
		config.Model = "gpt-4o-mini"
	}
	if config.Timeout == 0 {
		config.Timeout = 120 * time.Second
	}

	return &OpenAIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// openAIRequest is the request format for OpenAI API.
type openAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
	ToolChoice  any       `json:"tool_choice,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// openAIResponse is the response format from OpenAI API.
type openAIResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage Usage `json:"usage"`
}

// openAIStreamResponse is the streaming response format from OpenAI API.
type openAIStreamResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role      string     `json:"role,omitempty"`
			Content   string     `json:"content,omitempty"`
			ToolCalls []ToolCall `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

// Chat performs a non-streaming chat completion.
func (c *OpenAIClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = c.config.Model
	}

	openAIReq := openAIRequest{
		Model:       model,
		Messages:    req.Messages,
		Tools:       req.Tools,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      false,
	}

	if req.ToolChoice != "" {
		openAIReq.ToolChoice = req.ToolChoice
	}

	body, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &ChatResponse{
		ID:      openAIResp.ID,
		Model:   openAIResp.Model,
		Message: openAIResp.Choices[0].Message,
		Usage:   openAIResp.Usage,
	}, nil
}

// ChatStream performs a streaming chat completion.
func (c *OpenAIClient) ChatStream(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, <-chan error) {
	chunkCh := make(chan *StreamChunk, 100)
	errCh := make(chan error, 1)

	go func() {
		defer close(chunkCh)
		defer close(errCh)

		model := req.Model
		if model == "" {
			model = c.config.Model
		}

		openAIReq := openAIRequest{
			Model:       model,
			Messages:    req.Messages,
			Tools:       req.Tools,
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
			Stream:      true,
		}

		if req.ToolChoice != "" {
			openAIReq.ToolChoice = req.ToolChoice
		}

		body, err := json.Marshal(openAIReq)
		if err != nil {
			errCh <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			errCh <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
		httpReq.Header.Set("Accept", "text/event-stream")

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			errCh <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			errCh <- fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			if data == "[DONE]" {
				return
			}

			var streamResp openAIStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue
			}

			if len(streamResp.Choices) == 0 {
				continue
			}

			choice := streamResp.Choices[0]
			chunk := &StreamChunk{
				ID: streamResp.ID,
				Delta: struct {
					Content   string     `json:"content,omitempty"`
					ToolCalls []ToolCall `json:"tool_calls,omitempty"`
				}{
					Content:   choice.Delta.Content,
					ToolCalls: choice.Delta.ToolCalls,
				},
				FinishReason: choice.FinishReason,
			}

			select {
			case chunkCh <- chunk:
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errCh <- fmt.Errorf("failed to read stream: %w", err)
		}
	}()

	return chunkCh, errCh
}
