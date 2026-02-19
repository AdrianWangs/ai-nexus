package main

// Message router for WebUI.

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

	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
	"github.com/google/uuid"
)

// Router handles message routing between agents.
type Router struct {
	discovery *Discovery
	sessions  *SessionManager
	config    *Config
}

// NewRouter creates a new message router.
func NewRouter(discovery *Discovery, sessions *SessionManager, config *Config) *Router {
	return &Router{
		discovery: discovery,
		sessions:  sessions,
		config:    config,
	}
}

// RouteToAgent routes a message to an agent and returns the streaming response.
func (r *Router) RouteToAgent(
	ctx context.Context,
	sessionID string,
	agentName string,
	message string,
	from string,
	onEvent func(*protocol.StreamEvent) error,
) error {
	// Get agent info
	agent, ok := r.discovery.GetAgent(agentName)
	if !ok {
		return fmt.Errorf("agent not found: %s", agentName)
	}

	// Create task request
	traceID := uuid.New().String()
	callbackURL := fmt.Sprintf("http://%s:%d/api/callback", r.config.Host, r.config.Port)

	req := &protocol.TaskRequest{
		SessionID:   sessionID,
		TraceID:     traceID,
		From:        from,
		To:          "@" + agentName,
		Message:     message,
		CallbackURL: callbackURL,
		Stream:      true,
	}

	// Add outgoing message to session
	r.sessions.AddMessage(ctx, sessionID, &protocol.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		From:      from,
		To:        "@" + agentName,
		Content:   message,
		Type:      protocol.MessageTypeInternal,
		Timestamp: time.Now(),
	})

	// Send streaming request to agent
	return r.sendStreamingRequest(ctx, agent.URL+"/a2a/tasks/stream", req, onEvent)
}

// sendStreamingRequest sends a streaming request to an agent.
func (r *Router) sendStreamingRequest(
	ctx context.Context,
	url string,
	req *protocol.TaskRequest,
	onEvent func(*protocol.StreamEvent) error,
) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: 0} // No timeout for streaming
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("agent error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse SSE stream
	scanner := bufio.NewScanner(resp.Body)
	var contentBuffer strings.Builder

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
			break
		}

		var event protocol.StreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		// Accumulate content for final message
		if event.Type == protocol.EventTypeContentDelta {
			contentBuffer.WriteString(event.Content)
		}

		// Forward event to caller
		if err := onEvent(&event); err != nil {
			return err
		}

		// If done, save the complete message
		if event.Type == protocol.EventTypeDone {
			if contentBuffer.Len() > 0 {
				r.sessions.AddMessage(ctx, req.SessionID, &protocol.Message{
					ID:        uuid.New().String(),
					SessionID: req.SessionID,
					From:      event.From,
					To:        req.From,
					Content:   contentBuffer.String(),
					Type:      protocol.MessageTypeReply,
					Timestamp: time.Now(),
				})
			}
		}
	}

	return scanner.Err()
}

// HandleCallback handles a callback from an agent.
func (r *Router) HandleCallback(ctx context.Context, msg *protocol.CallbackMessage) error {
	// Add message to session
	message := &protocol.Message{
		ID:          uuid.New().String(),
		SessionID:   msg.SessionID,
		From:        msg.From,
		To:          msg.To,
		Content:     msg.Content,
		Type:        msg.Type,
		ToolCalls:   msg.ToolCalls,
		ToolResults: msg.ToolResults,
		Timestamp:   time.Now(),
	}

	return r.sessions.AddMessage(ctx, msg.SessionID, message)
}

// HandleRouteRequest handles a route request from an agent.
func (r *Router) HandleRouteRequest(
	ctx context.Context,
	req *protocol.RouteRequest,
	onEvent func(*protocol.StreamEvent) error,
) error {
	// Parse target agent name
	targetAgent := strings.TrimPrefix(req.To, "@")

	// Add the routing message to session
	r.sessions.AddMessage(ctx, req.SessionID, &protocol.Message{
		ID:        uuid.New().String(),
		SessionID: req.SessionID,
		From:      req.From,
		To:        req.To,
		Content:   req.Message,
		Type:      protocol.MessageTypeInternal,
		Timestamp: time.Now(),
	})

	// Route to target agent
	return r.RouteToAgent(ctx, req.SessionID, targetAgent, req.Message, req.From, onEvent)
}
