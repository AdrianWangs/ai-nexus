// Package session provides session storage and management.
package session

import (
	"context"

	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
)

// Store is the interface for session storage.
type Store interface {
	// CreateSession creates a new session.
	CreateSession(ctx context.Context, session *protocol.Session) error

	// GetSession retrieves a session by ID.
	GetSession(ctx context.Context, sessionID string) (*protocol.Session, error)

	// UpdateSession updates an existing session.
	UpdateSession(ctx context.Context, session *protocol.Session) error

	// DeleteSession deletes a session.
	DeleteSession(ctx context.Context, sessionID string) error

	// ListSessions lists all sessions.
	ListSessions(ctx context.Context) ([]*protocol.Session, error)

	// AddMessage adds a message to a session.
	AddMessage(ctx context.Context, sessionID string, message *protocol.Message) error

	// GetMessages retrieves all messages for a session.
	GetMessages(ctx context.Context, sessionID string) ([]*protocol.Message, error)

	// UpdateMessage updates an existing message (e.g., for streaming content).
	UpdateMessage(ctx context.Context, sessionID string, message *protocol.Message) error

	// Close closes the store connection.
	Close() error
}
