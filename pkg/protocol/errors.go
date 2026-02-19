package protocol

import "fmt"

// Error codes for the protocol.
const (
	ErrCodeInvalidRequest  = "INVALID_REQUEST"
	ErrCodeAgentNotFound   = "AGENT_NOT_FOUND"
	ErrCodeSessionNotFound = "SESSION_NOT_FOUND"
	ErrCodeToolNotFound    = "TOOL_NOT_FOUND"
	ErrCodeToolExecFailed  = "TOOL_EXEC_FAILED"
	ErrCodeLLMError        = "LLM_ERROR"
	ErrCodeCallbackFailed  = "CALLBACK_FAILED"
	ErrCodeRouteError      = "ROUTE_ERROR"
	ErrCodeTimeout         = "TIMEOUT"
	ErrCodeInternal        = "INTERNAL_ERROR"
)

// ProtocolError represents an error in the protocol.
type ProtocolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *ProtocolError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewError creates a new protocol error.
func NewError(code, message string) *ProtocolError {
	return &ProtocolError{Code: code, Message: message}
}

// NewErrorWithDetails creates a new protocol error with details.
func NewErrorWithDetails(code, message, details string) *ProtocolError {
	return &ProtocolError{Code: code, Message: message, Details: details}
}

// Common errors.
var (
	ErrInvalidRequest  = NewError(ErrCodeInvalidRequest, "Invalid request")
	ErrAgentNotFound   = NewError(ErrCodeAgentNotFound, "Agent not found")
	ErrSessionNotFound = NewError(ErrCodeSessionNotFound, "Session not found")
	ErrToolNotFound    = NewError(ErrCodeToolNotFound, "Tool not found")
)
