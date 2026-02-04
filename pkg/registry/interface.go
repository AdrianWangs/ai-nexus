// Package registry provides service registration and discovery.
package registry

import (
	"context"

	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
)

// EventType defines the type of watch event.
type EventType string

const (
	EventTypeAdd    EventType = "add"
	EventTypeUpdate EventType = "update"
	EventTypeDelete EventType = "delete"
)

// WatchEvent represents a registry change event.
type WatchEvent struct {
	Type  EventType
	Name  string
	Agent *protocol.AgentCard
}

// Registry is the interface for service registration and discovery.
type Registry interface {
	// Register registers an agent in the registry.
	Register(ctx context.Context, card *protocol.AgentCard) error

	// Deregister removes an agent from the registry.
	Deregister(ctx context.Context, name string) error

	// GetAgent retrieves an agent by name.
	GetAgent(ctx context.Context, name string) (*protocol.AgentCard, error)

	// ListAgents lists all registered agents.
	ListAgents(ctx context.Context) ([]*protocol.AgentCard, error)

	// Watch watches for registry changes.
	Watch(ctx context.Context, callback func(WatchEvent)) error

	// Close closes the registry connection.
	Close() error
}
