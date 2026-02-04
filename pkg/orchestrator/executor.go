package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AdrianWangs/ai-nexus/pkg/llm"
	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
)

func (o *Orchestrator) respondDirectly(ctx context.Context, req *protocol.TaskRequest) (*protocol.TaskResponse, error) {
	resp, err := o.llm.Chat(ctx, &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You are a helpful AI assistant. Be concise and helpful."},
			{Role: "user", Content: req.Message},
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

func (o *Orchestrator) respondDirectlyStream(
	ctx context.Context,
	req *protocol.TaskRequest,
	sendEvent func(*protocol.StreamEvent) error,
) error {
	chunkCh, errCh := o.llm.ChatStream(ctx, &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You are a helpful AI assistant. Be concise and helpful."},
			{Role: "user", Content: req.Message},
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

func (o *Orchestrator) executePlan(ctx context.Context, req *protocol.TaskRequest, plan *ExecutionPlan) ([]StepResult, error) {
	var results []StepResult
	for _, step := range plan.Steps {
		result, err := o.executeStep(ctx, req, step)
		if err != nil {
			result = StepResult{Agent: step.Agent, Task: step.Task, Success: false, Error: err.Error()}
		}
		results = append(results, result)
	}
	return results, nil
}

func (o *Orchestrator) executeStep(ctx context.Context, req *protocol.TaskRequest, step PlanStep) (StepResult, error) {
	agent, ok := o.GetAgent(step.Agent)
	if !ok {
		return StepResult{}, fmt.Errorf("agent not found: %s", step.Agent)
	}

	if o.config != nil && o.config.WebUIURL != "" {
		return o.executeStepViaWebUI(ctx, req, step, agent)
	}
	return o.executeStepDirect(ctx, req, step, agent)
}

func (o *Orchestrator) executeStepViaWebUI(ctx context.Context, req *protocol.TaskRequest, step PlanStep, agent *protocol.AgentCard) (StepResult, error) {
	routeReq := protocol.RouteRequest{
		SessionID:   req.SessionID,
		TraceID:     req.TraceID,
		From:        "@orchestrator",
		To:          "@" + step.Agent,
		Message:     step.Task,
		CallbackURL: req.CallbackURL,
	}

	data, _ := json.Marshal(routeReq)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.config.WebUIURL+"/api/route", bytes.NewReader(data))
	if err != nil {
		return StepResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return StepResult{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return StepResult{Agent: step.Agent, Task: step.Task, Content: string(body), Success: resp.StatusCode == http.StatusOK}, nil
}

func (o *Orchestrator) executeStepDirect(ctx context.Context, req *protocol.TaskRequest, step PlanStep, agent *protocol.AgentCard) (StepResult, error) {
	taskReq := protocol.TaskRequest{
		SessionID:   req.SessionID,
		TraceID:     req.TraceID,
		From:        "@orchestrator",
		To:          "@" + step.Agent,
		Message:     step.Task,
		CallbackURL: req.CallbackURL,
	}

	data, _ := json.Marshal(taskReq)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", agent.URL+"/a2a/tasks", bytes.NewReader(data))
	if err != nil {
		return StepResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return StepResult{}, err
	}
	defer resp.Body.Close()

	var taskResp protocol.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return StepResult{}, err
	}

	return StepResult{Agent: step.Agent, Task: step.Task, Content: taskResp.Content, Success: taskResp.Error == "", Error: taskResp.Error}, nil
}

