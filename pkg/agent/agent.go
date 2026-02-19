package agent

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/AdrianWangs/ai-nexus/pkg/llm"
	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
	"github.com/AdrianWangs/ai-nexus/pkg/registry"
	"github.com/AdrianWangs/ai-nexus/pkg/tools"
)

// Agent represents an AI agent that can process tasks.
type Agent struct {
	config   *Config
	llm      llm.Client
	registry registry.Registry
	tools    *tools.Registry

	// Callbacks
	onTaskReceived func(ctx context.Context, req *protocol.TaskRequest) (*protocol.TaskResponse, error)

	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	cacheMu    sync.RWMutex
	agentCache map[string]*protocol.AgentCard
}

// New creates a new agent with the given configuration.
func New(config *Config) (*Agent, error) {
	ctx, cancel := context.WithCancel(context.Background())

	agent := &Agent{
		config: config,
		tools:  tools.NewRegistry(),
		ctx:    ctx,
		cancel: cancel,
		agentCache: make(map[string]*protocol.AgentCard),
	}

	// Initialize LLM client
	if config.LLM.APIKey != "" {
		agent.llm = NewLLMClient(config.LLM)
	}

	return agent, nil
}

// NewFromConfig creates a new agent from a config file.
func NewFromConfig(configPath string) (*Agent, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return New(config)
}

// Name returns the agent's name.
func (a *Agent) Name() string {
	return a.config.Name
}

// Config returns the agent's configuration.
func (a *Agent) Config() *Config {
	return a.config
}

// LLM returns the LLM client.
func (a *Agent) LLM() llm.Client {
	return a.llm
}

// Tools returns the tool registry.
func (a *Agent) Tools() *tools.Registry { return a.tools }

// RegisterTool registers a tool in the default capability.
func (a *Agent) RegisterTool(tool *tools.ToolDefinition) { a.tools.Register("default", tool) }

// RegisterToolFunc registers a function as a tool in the default capability.
func (a *Agent) RegisterToolFunc(name, description string, fn interface{}) error {
	return a.tools.RegisterFunc("default", name, description, fn)
}

// RegisterToolFuncWithCapability registers a function as a tool under a capability.
func (a *Agent) RegisterToolFuncWithCapability(capability, name, description string, fn interface{}) error {
	return a.tools.RegisterFunc(capability, name, description, fn)
}

// SetTaskHandler sets a custom task handler.
func (a *Agent) SetTaskHandler(handler func(ctx context.Context, req *protocol.TaskRequest) (*protocol.TaskResponse, error)) {
	a.onTaskReceived = handler
}

// InitRegistry initializes the registry connection.
func (a *Agent) InitRegistry() error {
	if a.config.Registry.Type != "etcd" && a.config.Registry.Type != "" {
		return fmt.Errorf("unsupported registry type: %s", a.config.Registry.Type)
	}

	if len(a.config.Registry.Endpoints) == 0 {
		log.Println("No registry endpoints configured, skipping registration")
		return nil
	}

	reg, err := registry.NewETCDRegistry(registry.ETCDConfig{
		Endpoints: a.config.Registry.Endpoints,
		Prefix:    a.config.Registry.Prefix,
		TTL:       a.config.Registry.TTL,
	})
	if err != nil {
		return fmt.Errorf("failed to create registry: %w", err)
	}

	a.registry = reg
	return nil
}

// Register registers the agent with the registry.
func (a *Agent) Register() error {
	if a.registry == nil {
		return nil
	}

	card := a.GetCard()
	return a.registry.Register(a.ctx, card)
}

// Deregister removes the agent from the registry.
func (a *Agent) Deregister() error {
	if a.registry == nil {
		return nil
	}

	return a.registry.Deregister(a.ctx, a.config.Name)
}

// GetCard returns the agent's card.
func (a *Agent) GetCard() *protocol.AgentCard {
	agentType := protocol.AgentTypeNormal
	if a.config.Type == "orchestrator" {
		agentType = protocol.AgentTypeOrchestrator
	}

	// Build URL
	url := fmt.Sprintf("http://%s:%d", a.config.Host, a.config.Port)

	// Convert tools
	tlist := a.tools.List()
	toolsOut := make([]protocol.Tool, 0, len(tlist))
	for _, t := range tlist {
		toolsOut = append(toolsOut, protocol.Tool{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  t.Parameters,
		})
	}

	// Convert skills
	skills := make([]protocol.Skill, 0, len(a.config.Skills))
	for _, s := range a.config.Skills {
		skills = append(skills, protocol.Skill{
			Name:        s.Name,
			Description: s.Description,
			Keywords:    s.Keywords,
		})
	}

	// Convert capabilities
	caps := make([]protocol.Capability, 0, len(a.config.Capabilities))
	for _, name := range a.config.Capabilities {
		caps = append(caps, protocol.Capability{Name: name})
	}
	if len(caps) == 0 {
		for _, c := range a.tools.CapabilityIndex() {
			caps = append(caps, protocol.Capability{Name: c.Name, Description: c.Description, Keywords: c.Keywords})
		}
	}

	return &protocol.AgentCard{
		Name:         a.config.Name,
		DisplayName:  a.config.DisplayName,
		Description:  a.config.Description,
		Version:      a.config.Version,
		URL:          url,
		Port:         a.config.Port,
		Capabilities: caps,
		Tools:        toolsOut,
		Skills:       skills,
		Type:         agentType,
		Status:       protocol.AgentStatusOnline,
	}
}

// Close closes the agent.
func (a *Agent) Close() error {
	a.cancel()

	if a.registry != nil {
		a.Deregister()
		a.registry.Close()
	}

	return nil
}
