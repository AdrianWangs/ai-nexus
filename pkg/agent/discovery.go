package agent

import (
	"log"

	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
	"github.com/AdrianWangs/ai-nexus/pkg/registry"
)

// StartDiscovery loads agents from registry and keeps a local cache up-to-date via watch.
// It is safe to call multiple times; subsequent calls will reload cache.
func (a *Agent) StartDiscovery() error {
	if a.registry == nil {
		return nil
	}

	agents, err := a.registry.ListAgents(a.ctx)
	if err != nil {
		return err
	}

	a.cacheMu.Lock()
	for _, ag := range agents {
		a.agentCache[ag.Name] = ag
	}
	a.cacheMu.Unlock()

	go func() {
		err := a.registry.Watch(a.ctx, func(event registry.WatchEvent) {
			switch event.Type {
			case registry.EventTypeAdd, registry.EventTypeUpdate:
				if event.Agent == nil {
					return
				}
				a.cacheMu.Lock()
				a.agentCache[event.Agent.Name] = event.Agent
				a.cacheMu.Unlock()
			case registry.EventTypeDelete:
				name := event.Name
				if name == "" && event.Agent != nil {
					name = event.Agent.Name
				}
				if name == "" {
					return
				}
				a.cacheMu.Lock()
				delete(a.agentCache, name)
				a.cacheMu.Unlock()
			}
		})
		if err != nil && a.ctx.Err() == nil {
			log.Printf("[agent:%s] registry watch stopped: %v", a.config.Name, err)
		}
	}()

	return nil
}

// GetAgent returns an agent card from local cache.
func (a *Agent) GetAgent(name string) (*protocol.AgentCard, bool) {
	a.cacheMu.RLock()
	defer a.cacheMu.RUnlock()
	ag, ok := a.agentCache[name]
	return ag, ok
}

// ListAgents returns all agent cards from local cache.
func (a *Agent) ListAgents() []*protocol.AgentCard {
	a.cacheMu.RLock()
	defer a.cacheMu.RUnlock()
	out := make([]*protocol.AgentCard, 0, len(a.agentCache))
	for _, ag := range a.agentCache {
		out = append(out, ag)
	}
	return out
}

