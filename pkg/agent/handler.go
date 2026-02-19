package agent

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

	"github.com/AdrianWangs/ai-nexus/pkg/llm"
	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
)

// Handler handles incoming A2A requests.
type Handler struct {
	agent *Agent
}

// NewHandler creates a new handler for the agent.
func NewHandler(agent *Agent) *Handler {
	return &Handler{agent: agent}
}

// HandleTask handles a task request.
func (h *Handler) HandleTask(ctx context.Context, req *protocol.TaskRequest) (*protocol.TaskResponse, error) {
	// If custom handler is set, use it
	if h.agent.onTaskReceived != nil {
		return h.agent.onTaskReceived(ctx, req)
	}

	// Default handling: use LLM with tools
	return h.processWithLLM(ctx, req)
}

// HandleTaskStream handles a streaming task request.
func (h *Handler) HandleTaskStream(ctx context.Context, req *protocol.TaskRequest, sendEvent func(*protocol.StreamEvent) error) error {
	// Build messages for LLM
	messages := []llm.Message{
		{
			Role:    "system",
			Content: h.buildSystemPrompt(),
		},
		{
			Role:    "user",
			Content: req.Message,
		},
	}

	// Get LLM tools
	tools := h.agent.tools.ToLLMTools("")

	// Signal stream start to downstream (e.g. WebUI)
	_ = sendEvent(&protocol.StreamEvent{
		Type:      protocol.EventTypeMessageStart,
		SessionID: req.SessionID,
		TraceID:   req.TraceID,
		From:      "@" + h.agent.config.Name,
		To:        req.From,
	})

	// Stream from LLM
	chunkCh, errCh := h.agent.llm.ChatStream(ctx, &llm.ChatRequest{
		Messages: messages,
		Tools:    tools,
		Stream:   true,
	})

	var contentBuffer strings.Builder
	var toolCalls []llm.ToolCall

	for {
		select {
		case chunk, ok := <-chunkCh:
			if !ok {
				// Stream ended
				goto processToolCalls
			}

			// Handle content
			if chunk.Delta.Content != "" {
				contentBuffer.WriteString(chunk.Delta.Content)
				sendEvent(&protocol.StreamEvent{
					Type:      protocol.EventTypeContent,
					SessionID: req.SessionID,
					From:      "@" + h.agent.config.Name,
					To:        req.From,
					Content:   chunk.Delta.Content,
				})
			}

			// Collect tool calls
			if len(chunk.Delta.ToolCalls) > 0 {
				toolCalls = append(toolCalls, chunk.Delta.ToolCalls...)
			}

		case err := <-errCh:
			if err != nil {
				sendEvent(&protocol.StreamEvent{
					Type:  protocol.EventTypeError,
					Error: err.Error(),
				})
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

processToolCalls:
	// Execute tool calls if any
	if len(toolCalls) > 0 {
		for _, tc := range toolCalls {
			// Send tool call event
			sendEvent(&protocol.StreamEvent{
				Type: protocol.EventTypeToolCall,
				ToolCall: &protocol.ToolCall{
					ID:   tc.ID,
					Name: tc.Function.Name,
				},
			})

			// Execute tool
			result, err := h.agent.tools.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}

			// Send tool result event
			sendEvent(&protocol.StreamEvent{
				Type: protocol.EventTypeToolResult,
				ToolResult: &protocol.ToolResult{
					ToolCallID: tc.ID,
					Content:    result,
				},
			})

			// Continue conversation with tool result
			messages = append(messages, llm.Message{
				Role:      "assistant",
				ToolCalls: toolCalls,
			})
			messages = append(messages, llm.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}

		// Get final response after tool calls
		resp, err := h.agent.llm.Chat(ctx, &llm.ChatRequest{
			Messages: messages,
		})
		if err != nil {
			return err
		}

		// Send final content
		sendEvent(&protocol.StreamEvent{
			Type:      protocol.EventTypeContent,
			SessionID: req.SessionID,
			From:      "@" + h.agent.config.Name,
			To:        req.From,
			Content:   resp.Message.Content,
		})
	}

	// Send done event
	sendEvent(&protocol.StreamEvent{
		Type:      protocol.EventTypeDone,
		SessionID: req.SessionID,
	})

	// If callback URL is set, send the result
	if req.CallbackURL != "" {
		go h.sendCallback(req, contentBuffer.String(), nil)
	}

	return nil
}

// processWithLLM processes a request using the LLM.
func (h *Handler) processWithLLM(ctx context.Context, req *protocol.TaskRequest) (*protocol.TaskResponse, error) {
	messages := []llm.Message{
		{
			Role:    "system",
			Content: h.buildSystemPrompt(),
		},
		{
			Role:    "user",
			Content: req.Message,
		},
	}

	tools := h.agent.tools.ToLLMTools("")

	resp, err := h.agent.llm.Chat(ctx, &llm.ChatRequest{
		Messages: messages,
		Tools:    tools,
	})
	if err != nil {
		return nil, err
	}

	// Handle tool calls
	if len(resp.Message.ToolCalls) > 0 {
		var toolResults []protocol.ToolResult

		for _, tc := range resp.Message.ToolCalls {
			result, execErr := h.agent.tools.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
			if execErr != nil {
				result = fmt.Sprintf("Error: %v", execErr)
			}

			toolResults = append(toolResults, protocol.ToolResult{
				ToolCallID: tc.ID,
				Content:    result,
			})

			// Add to messages for final response
			messages = append(messages, llm.Message{
				Role:      "assistant",
				ToolCalls: []llm.ToolCall{tc},
			})
			messages = append(messages, llm.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}

		// Get final response
		finalResp, err := h.agent.llm.Chat(ctx, &llm.ChatRequest{
			Messages: messages,
		})
		if err != nil {
			return nil, err
		}

		response := &protocol.TaskResponse{
			SessionID:   req.SessionID,
			TraceID:     req.TraceID,
			From:        "@" + h.agent.config.Name,
			To:          req.From,
			Content:     finalResp.Message.Content,
			ToolResults: toolResults,
		}

		// Send callback if URL is set
		if req.CallbackURL != "" {
			go h.sendCallback(req, response.Content, toolResults)
		}

		return response, nil
	}

	response := &protocol.TaskResponse{
		SessionID: req.SessionID,
		TraceID:   req.TraceID,
		From:      "@" + h.agent.config.Name,
		To:        req.From,
		Content:   resp.Message.Content,
	}

	// Send callback if URL is set
	if req.CallbackURL != "" {
		go h.sendCallback(req, response.Content, nil)
	}

	return response, nil
}

// sendCallback sends a callback message to the WebUI.
func (h *Handler) sendCallback(req *protocol.TaskRequest, content string, toolResults []protocol.ToolResult) {
	msg := protocol.CallbackMessage{
		SessionID:   req.SessionID,
		TraceID:     req.TraceID,
		From:        "@" + h.agent.config.Name,
		To:          req.From,
		Content:     content,
		Type:        protocol.MessageTypeReply,
		IsFinal:     true,
		ToolResults: toolResults,
	}

	data, _ := json.Marshal(msg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", req.CallbackURL, bytes.NewReader(data))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		fmt.Printf("Failed to send callback: %v\n", err)
		return
	}
	defer resp.Body.Close()
}

// buildSystemPrompt builds the system prompt for the LLM.
func (h *Handler) buildSystemPrompt() string {
	prompt := fmt.Sprintf("You are %s, an AI assistant.", h.agent.config.Name)
	if h.agent.config.Description != "" {
		prompt += " " + h.agent.config.Description
	}
	return prompt
}

// SendTaskToAgent sends a task to another agent.
func SendTaskToAgent(ctx context.Context, agentURL string, req *protocol.TaskRequest) (*protocol.TaskResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", agentURL+"/a2a/tasks", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("agent error (status %d): %s", resp.StatusCode, string(body))
	}

	var taskResp protocol.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return nil, err
	}

	return &taskResp, nil
}

// SendTaskStreamToAgent sends a streaming task to another agent.
func SendTaskStreamToAgent(ctx context.Context, agentURL string, req *protocol.TaskRequest, onEvent func(*protocol.StreamEvent) error) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", agentURL+"/a2a/tasks/stream", bytes.NewReader(data))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("agent error (status %d): %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" || !strings.HasPrefix(line, "data: ") {
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

		if err := onEvent(&event); err != nil {
			return err
		}
	}

	return scanner.Err()
}
