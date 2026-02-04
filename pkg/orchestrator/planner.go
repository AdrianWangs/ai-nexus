package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/AdrianWangs/ai-nexus/pkg/llm"
	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
)

func (o *Orchestrator) plan(ctx context.Context, req *protocol.TaskRequest) (*ExecutionPlan, error) {
	agents := o.GetAgents()
	if len(agents) == 0 {
		return &ExecutionPlan{DirectResponse: true}, nil
	}

	var agentDescs strings.Builder
	for _, agent := range agents {
		agentDescs.WriteString(fmt.Sprintf("- %s: %s\n", agent.Name, agent.Description))
		if len(agent.Capabilities) > 0 {
			agentDescs.WriteString("  Capabilities: ")
			for i, c := range agent.Capabilities {
				if i > 0 {
					agentDescs.WriteString(", ")
				}
				agentDescs.WriteString(c.Name)
			}
			agentDescs.WriteString("\n")
		}
		if len(agent.Tools) > 0 {
			agentDescs.WriteString("  Tools: ")
			for i, tool := range agent.Tools {
				if i > 0 {
					agentDescs.WriteString(", ")
				}
				agentDescs.WriteString(tool.Name)
			}
			agentDescs.WriteString("\n")
		}
		if len(agent.Skills) > 0 {
			agentDescs.WriteString("  Skills: ")
			for i, skill := range agent.Skills {
				if i > 0 {
					agentDescs.WriteString(", ")
				}
				agentDescs.WriteString(skill.Name)
			}
			agentDescs.WriteString("\n")
		}
	}

	systemPrompt := fmt.Sprintf(`You are an AI orchestrator. Analyze the user's request and determine which agents to call.

Available agents:
%s

Respond with a JSON object:
{
  "direct_response": true/false,
  "steps": [
    {"agent": "agent-name", "task": "specific task description", "order": 1}
  ],
  "reasoning": "brief explanation"
}

If the request is a simple greeting or general question, set direct_response to true with empty steps.
If agents are needed, list them in order of execution.`, agentDescs.String())

	resp, err := o.llm.Chat(ctx, &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: req.Message},
		},
		Temperature: 0.3,
	})
	if err != nil {
		return nil, err
	}

	// Parse the response
	var plan ExecutionPlan
	content := resp.Message.Content

	// Extract JSON from response
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		content = content[start : end+1]
	}

	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		log.Printf("Failed to parse plan: %v, content: %s", err, content)
		return &ExecutionPlan{DirectResponse: true}, nil
	}

	if plan.DirectResponse {
		plan.Steps = nil
	}
	return &plan, nil
}

