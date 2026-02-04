// Package llm provides LLM client implementations.
package llm

import (
	"context"
)

// Message represents a message in a conversation.
type Message struct {
	Role       string     `json:"role"` // "system", "user", "assistant", "tool"
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolCall represents a tool/function call from the LLM.
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "function"
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"` // JSON string
	} `json:"function"`
}

// Tool represents a tool definition for the LLM.
type Tool struct {
	Type     string       `json:"type"` // "function"
	Function ToolFunction `json:"function"`
}

// ToolFunction describes a function that can be called.
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ChatRequest represents a chat completion request.
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
	ToolChoice  string    `json:"tool_choice,omitempty"` // "auto", "none", or specific tool
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatResponse represents a chat completion response.
type ChatResponse struct {
	ID      string  `json:"id"`
	Model   string  `json:"model"`
	Message Message `json:"message"`
	Usage   Usage   `json:"usage,omitempty"`
}

// Usage represents token usage information.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk represents a streaming response chunk.
type StreamChunk struct {
	ID    string `json:"id"`
	Delta struct {
		Content   string     `json:"content,omitempty"`
		ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	} `json:"delta"`
	FinishReason string `json:"finish_reason,omitempty"`
}

// Client is the interface for LLM clients.
type Client interface {
	// Chat performs a non-streaming chat completion.
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// ChatStream performs a streaming chat completion.
	ChatStream(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, <-chan error)
}
