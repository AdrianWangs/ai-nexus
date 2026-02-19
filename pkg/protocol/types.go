// Package protocol defines the core types for the AI-Nexus messaging system.
package protocol

import (
	"time"
)

// MessageType defines the type of message in a session.
type MessageType string

const (
	// MessageTypeUser is a message from the end user.
	MessageTypeUser MessageType = "user"
	// MessageTypeReply is a direct reply to the user.
	MessageTypeReply MessageType = "reply"
	// MessageTypeInternal is an internal message between agents (collapsible in UI).
	MessageTypeInternal MessageType = "internal"
	// MessageTypeSystem is a system message (errors, status updates).
	MessageTypeSystem MessageType = "system"
)

// Message represents a single message in a session.
type Message struct {
	ID        string      `json:"id"`
	SessionID string      `json:"session_id"`
	From      string      `json:"from"` // "user" or "@agent-name"
	To        string      `json:"to"`   // "@agent-name" or "user"
	Content   string      `json:"content"`
	Type      MessageType `json:"type"`
	Timestamp time.Time   `json:"timestamp"`

	// Optional fields for tool calls
	ToolCalls   []ToolCall   `json:"tool_calls,omitempty"`
	ToolResults []ToolResult `json:"tool_results,omitempty"`

	// Metadata for additional context
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall represents a tool/function call made by an agent.
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult represents the result of a tool call.
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error,omitempty"`
}

// Session represents a conversation session (group chat).
type Session struct {
	ID           string    `json:"id"`
	Title        string    `json:"title,omitempty"`
	Participants []string  `json:"participants"` // List of agent names involved
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TaskRequest is the request sent to an agent to process a task.
type TaskRequest struct {
	// Session identification
	SessionID string `json:"session_id"`
	TraceID   string `json:"trace_id,omitempty"`

	// Message details
	From    string `json:"from"`
	To      string `json:"to,omitempty"`
	Message string `json:"message"`

	// Callback configuration
	CallbackURL string `json:"callback_url,omitempty"`

	// Streaming configuration
	Stream bool `json:"stream,omitempty"`

	// Shared context between agents
	SharedContext map[string]interface{} `json:"shared_context,omitempty"`
}

// TaskResponse is the response from an agent (non-streaming).
type TaskResponse struct {
	SessionID string `json:"session_id"`
	TraceID   string `json:"trace_id,omitempty"`
	From      string `json:"from"`
	To        string `json:"to"`
	Content   string `json:"content"`

	// Tool calls if any
	ToolCalls   []ToolCall   `json:"tool_calls,omitempty"`
	ToolResults []ToolResult `json:"tool_results,omitempty"`

	// Error information
	Error string `json:"error,omitempty"`
}

// CallbackMessage is sent by agents to the WebUI callback endpoint.
type CallbackMessage struct {
	SessionID string      `json:"session_id"`
	TraceID   string      `json:"trace_id,omitempty"`
	From      string      `json:"from"`
	To        string      `json:"to"`
	Content   string      `json:"content"`
	Type      MessageType `json:"type"`

	// For streaming
	IsPartial bool `json:"is_partial,omitempty"`
	IsFinal   bool `json:"is_final,omitempty"`

	// Tool information
	ToolCalls   []ToolCall   `json:"tool_calls,omitempty"`
	ToolResults []ToolResult `json:"tool_results,omitempty"`

	// Error information
	Error string `json:"error,omitempty"`
}

// RouteRequest is sent by agents to request routing to another agent.
type RouteRequest struct {
	SessionID   string `json:"session_id"`
	TraceID     string `json:"trace_id,omitempty"`
	From        string `json:"from"`
	To          string `json:"to"` // Target agent name
	Message     string `json:"message"`
	CallbackURL string `json:"callback_url,omitempty"`

	// SharedContext to pass to the target agent
	SharedContext map[string]interface{} `json:"shared_context,omitempty"`
}

// RouteResponse is the response from a route request.
type RouteResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}
