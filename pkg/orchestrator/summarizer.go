package orchestrator

import (
	"context"
	"fmt"
	"strings"

	"github.com/AdrianWangs/ai-nexus/pkg/llm"
	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
)

func (o *Orchestrator) summarizeResults(ctx context.Context, req *protocol.TaskRequest, results []StepResult) (*protocol.TaskResponse, error) {
	var resultsText strings.Builder
	for _, r := range results {
		resultsText.WriteString(fmt.Sprintf("Agent %s:\n%s\n\n", r.Agent, r.Content))
	}

	systemPrompt := `You are an AI assistant. Summarize the following agent responses into a coherent answer for the user.
Be concise and focus on the key information. If there were errors, mention them briefly.`

	resp, err := o.llm.Chat(ctx, &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: fmt.Sprintf("User's original question: %s\n\nAgent responses:\n%s", req.Message, resultsText.String())},
		},
	})
	if err != nil {
		return nil, err
	}

	return &protocol.TaskResponse{
		SessionID: req.SessionID,
		TraceID:   req.TraceID,
		From:      "@orchestrator",
		To:        req.From,
		Content:   resp.Message.Content,
	}, nil
}

func (o *Orchestrator) summarizeResultsStream(
	ctx context.Context,
	req *protocol.TaskRequest,
	results []StepResult,
	sendEvent func(*protocol.StreamEvent) error,
) error {
	var resultsText strings.Builder
	for _, r := range results {
		resultsText.WriteString(fmt.Sprintf("Agent %s:\n%s\n\n", r.Agent, r.Content))
	}

	systemPrompt := `You are an AI assistant. Summarize the following agent responses into a coherent answer for the user.
Be concise and focus on the key information. If there were errors, mention them briefly.`

	chunkCh, errCh := o.llm.ChatStream(ctx, &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: fmt.Sprintf("User's original question: %s\n\nAgent responses:\n%s", req.Message, resultsText.String())},
		},
		Stream: true,
	})

	for {
		select {
		case chunk, ok := <-chunkCh:
			if !ok {
				_ = sendEvent(&protocol.StreamEvent{Type: protocol.EventTypeDone, SessionID: req.SessionID})
				return nil
			}
			if chunk.Delta.Content != "" {
				_ = sendEvent(&protocol.StreamEvent{
					Type:      protocol.EventTypeContentDelta,
					SessionID: req.SessionID,
					From:      "@orchestrator",
					To:        req.From,
					Content:   chunk.Delta.Content,
				})
			}
		case err := <-errCh:
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

