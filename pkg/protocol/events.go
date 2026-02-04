package protocol

// StreamEventType defines the type of streaming event.
type StreamEventType string

const (
	// EventTypeMessageStart signals the start of a streamed assistant message.
	EventTypeMessageStart StreamEventType = "message_start"
	// EventTypeContentDelta is a content chunk (partial message).
	EventTypeContentDelta StreamEventType = "content_delta"
	// EventTypeMessageEnd is a complete message.
	EventTypeMessageEnd StreamEventType = "message_end"
	// EventTypeThinking is agent thinking/reasoning output.
	EventTypeThinking StreamEventType = "thinking"
	// EventTypeToolCall is when agent starts a tool call.
	EventTypeToolCall StreamEventType = "tool_call"
	// EventTypeToolResult is the result of a tool call.
	EventTypeToolResult StreamEventType = "tool_result"
	// EventTypeAgentCall is when an agent calls another agent.
	EventTypeAgentCall StreamEventType = "agent_call"
	// EventTypeAgentResult is the result from another agent.
	EventTypeAgentResult StreamEventType = "agent_result"
	// EventTypeStatus is a status update.
	EventTypeStatus StreamEventType = "status"
	// EventTypeDone signals the stream is complete.
	EventTypeDone StreamEventType = "done"
	// EventTypeError signals an error occurred.
	EventTypeError StreamEventType = "error"

	// Backward-compatible aliases (will be removed after migration).
	EventTypeContent StreamEventType = EventTypeContentDelta
	EventTypeMessage StreamEventType = EventTypeMessageEnd
)

// StreamEvent represents a single event in a streaming response.
type StreamEvent struct {
	Type      StreamEventType `json:"type"`
	SessionID string          `json:"session_id,omitempty"`
	TraceID   string          `json:"trace_id,omitempty"`
	From      string          `json:"from,omitempty"`
	To        string          `json:"to,omitempty"`

	// Content for text events
	Content string `json:"content,omitempty"`

	// For tool events
	ToolCall   *ToolCall   `json:"tool_call,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`

	// For agent call events
	TargetAgent string `json:"target_agent,omitempty"`

	// For complete message events
	Message *Message `json:"message,omitempty"`

	// For error events
	Error string `json:"error,omitempty"`

	// For status events
	Status string `json:"status,omitempty"`
}
