package orchestrator

import (
	"context"
	"log"
	"sync"

	"github.com/AdrianWangs/ai-nexus/pkg/llm"
	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
	"github.com/AdrianWangs/ai-nexus/pkg/registry"
)

// Orchestrator handles intelligent request routing and multi-agent coordination.
type Orchestrator struct {
	config   *Config
	llm      llm.Client
	registry registry.Registry

	agents map[string]*protocol.AgentCard
	mu     sync.RWMutex
}

// New creates a new orchestrator.
func New(config *Config, llmClient llm.Client) *Orchestrator {
	return &Orchestrator{
		config: config,
		llm:    llmClient,
		agents: make(map[string]*protocol.AgentCard),
	}
}

// SetRegistry sets the registry for agent discovery.
func (o *Orchestrator) SetRegistry(reg registry.Registry) {
	o.registry = reg
}

// LoadAgents loads agents from the registry.
func (o *Orchestrator) LoadAgents(ctx context.Context) error {
	if o.registry == nil {
		return nil
	}

	agents, err := o.registry.ListAgents(ctx)
	if err != nil {
		return err
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	for _, agent := range agents {
		// Skip self
		if agent.Type == protocol.AgentTypeOrchestrator {
			continue
		}
		o.agents[agent.Name] = agent
		log.Printf("Loaded agent: %s at %s", agent.Name, agent.URL)
	}

	return nil
}

// WatchAgents watches for agent changes.
func (o *Orchestrator) WatchAgents(ctx context.Context) error {
	if o.registry == nil {
		return nil
	}

	return o.registry.Watch(ctx, func(event registry.WatchEvent) {
		o.mu.Lock()
		defer o.mu.Unlock()

		switch event.Type {
		case registry.EventTypeAdd, registry.EventTypeUpdate:
			if event.Agent != nil && event.Agent.Type != protocol.AgentTypeOrchestrator {
				o.agents[event.Agent.Name] = event.Agent
				log.Printf("Agent %s: %s", event.Type, event.Agent.Name)
			}
		case registry.EventTypeDelete:
			name := event.Name
			if name == "" && event.Agent != nil {
				name = event.Agent.Name
			}
			if name != "" {
				delete(o.agents, name)
				log.Printf("Agent removed: %s", name)
			}
		}
	})
}

// GetAgents returns all known agents.
func (o *Orchestrator) GetAgents() []*protocol.AgentCard {
	o.mu.RLock()
	defer o.mu.RUnlock()

	agents := make([]*protocol.AgentCard, 0, len(o.agents))
	for _, agent := range o.agents {
		agents = append(agents, agent)
	}
	return agents
}

// GetAgent returns an agent by name.
func (o *Orchestrator) GetAgent(name string) (*protocol.AgentCard, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	agent, ok := o.agents[name]
	return agent, ok
}

// ProcessRequest processes a request and routes it to appropriate agents.
func (o *Orchestrator) ProcessRequest(ctx context.Context, req *protocol.TaskRequest) (*protocol.TaskResponse, error) {
	plan, err := o.plan(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(plan.Steps) == 0 {
		return o.respondDirectly(ctx, req)
	}

	results, err := o.executePlan(ctx, req, plan)
	if err != nil {
		return nil, err
	}

	return o.summarizeResults(ctx, req, results)
}

// ProcessRequestStream processes a request with streaming response.
func (o *Orchestrator) ProcessRequestStream(
	ctx context.Context,
	req *protocol.TaskRequest,
	sendEvent func(*protocol.StreamEvent) error,
) error {
	_ = sendEvent(&protocol.StreamEvent{Type: protocol.EventTypeThinking, Status: "Analyzing your request..."})

	plan, err := o.plan(ctx, req)
	if err != nil {
		return err
	}

	if len(plan.Steps) == 0 {
		return o.respondDirectlyStream(ctx, req, sendEvent)
	}

	var results []StepResult
	for _, step := range plan.Steps {
		_ = sendEvent(&protocol.StreamEvent{
			Type:        protocol.EventTypeAgentCall,
			TargetAgent: step.Agent,
			Status:      "Calling " + step.Agent + ": " + step.Task,
		})

		result, err := o.executeStep(ctx, req, step)
		if err != nil {
			result = StepResult{Agent: step.Agent, Success: false, Error: err.Error()}
		}
		results = append(results, result)

		_ = sendEvent(&protocol.StreamEvent{
			Type:        protocol.EventTypeAgentResult,
			TargetAgent: step.Agent,
			Content:     result.Content,
		})
	}

	return o.summarizeResultsStream(ctx, req, results, sendEvent)
}

