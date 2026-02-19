package main

// Agent discovery for WebUI.

import (
	"context"
	"log"
	"sync"

	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
	"github.com/AdrianWangs/ai-nexus/pkg/registry"
)

// Discovery manages agent discovery from the registry.
type Discovery struct {
	registry registry.Registry
	agents   map[string]*protocol.AgentCard
	mu       sync.RWMutex

	onChange func(event registry.WatchEvent)
}

// NewDiscovery creates a new discovery service.
func NewDiscovery(config RegistryConfig) (*Discovery, error) {
	if len(config.Endpoints) == 0 {
		return &Discovery{
			agents: make(map[string]*protocol.AgentCard),
		}, nil
	}

	reg, err := registry.NewETCDRegistry(registry.ETCDConfig{
		Endpoints: config.Endpoints,
		Prefix:    config.Prefix,
		TTL:       config.TTL,
	})
	if err != nil {
		return nil, err
	}

	return &Discovery{
		registry: reg,
		agents:   make(map[string]*protocol.AgentCard),
	}, nil
}

// Start starts the discovery service.
func (d *Discovery) Start(ctx context.Context) error {
	if d.registry == nil {
		return nil
	}

	// Load existing agents
	agents, err := d.registry.ListAgents(ctx)
	if err != nil {
		return err
	}

	d.mu.Lock()
	for _, agent := range agents {
		d.agents[agent.Name] = agent
		log.Printf("Discovered agent: %s at %s", agent.Name, agent.URL)
	}
	d.mu.Unlock()

	// Start watching for changes
	go func() {
		err := d.registry.Watch(ctx, func(event registry.WatchEvent) {
			d.handleEvent(event)
		})
		if err != nil && ctx.Err() == nil {
			log.Printf("Watch error: %v", err)
		}
	}()

	return nil
}

// handleEvent handles a registry event.
func (d *Discovery) handleEvent(event registry.WatchEvent) {
	d.mu.Lock()
	defer d.mu.Unlock()

	switch event.Type {
	case registry.EventTypeAdd, registry.EventTypeUpdate:
		if event.Agent != nil {
			d.agents[event.Agent.Name] = event.Agent
			log.Printf("Agent %s: %s at %s", event.Type, event.Agent.Name, event.Agent.URL)
		}
	case registry.EventTypeDelete:
		name := event.Name
		if name == "" && event.Agent != nil {
			name = event.Agent.Name
		}
		if name != "" {
			delete(d.agents, name)
			log.Printf("Agent removed: %s", name)
		}
	}

	if d.onChange != nil {
		d.onChange(event)
	}
}

// SetOnChange sets the change callback.
func (d *Discovery) SetOnChange(fn func(event registry.WatchEvent)) {
	d.onChange = fn
}

// GetAgent retrieves an agent by name.
func (d *Discovery) GetAgent(name string) (*protocol.AgentCard, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	agent, ok := d.agents[name]
	return agent, ok
}

// ListAgents returns all known agents.
func (d *Discovery) ListAgents() []*protocol.AgentCard {
	d.mu.RLock()
	defer d.mu.RUnlock()

	agents := make([]*protocol.AgentCard, 0, len(d.agents))
	for _, agent := range d.agents {
		agents = append(agents, agent)
	}
	return agents
}

// Close closes the discovery service.
func (d *Discovery) Close() error {
	if d.registry != nil {
		return d.registry.Close()
	}
	return nil
}
