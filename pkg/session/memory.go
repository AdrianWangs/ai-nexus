package session

import (
	"context"
	"sync"
	"time"

	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
)

// MemoryStore implements Store using in-memory storage.
// This is useful for testing and development.
type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*protocol.Session
	messages map[string][]*protocol.Message
}

// NewMemoryStore creates a new in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]*protocol.Session),
		messages: make(map[string][]*protocol.Message),
	}
}

// CreateSession creates a new session.
func (s *MemoryStore) CreateSession(ctx context.Context, session *protocol.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = session
	s.messages[session.ID] = make([]*protocol.Message, 0)
	return nil
}

// GetSession retrieves a session by ID.
func (s *MemoryStore) GetSession(ctx context.Context, sessionID string) (*protocol.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, protocol.ErrSessionNotFound
	}
	return session, nil
}

// UpdateSession updates an existing session.
func (s *MemoryStore) UpdateSession(ctx context.Context, session *protocol.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessions[session.ID]; !ok {
		return protocol.ErrSessionNotFound
	}
	s.sessions[session.ID] = session
	return nil
}

// DeleteSession deletes a session and its messages.
func (s *MemoryStore) DeleteSession(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
	delete(s.messages, sessionID)
	return nil
}

// ListSessions lists all sessions.
func (s *MemoryStore) ListSessions(ctx context.Context) ([]*protocol.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*protocol.Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// AddMessage adds a message to a session.
func (s *MemoryStore) AddMessage(ctx context.Context, sessionID string, message *protocol.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessions[sessionID]; !ok {
		return protocol.ErrSessionNotFound
	}

	s.messages[sessionID] = append(s.messages[sessionID], message)

	// Update session timestamp
	if session, ok := s.sessions[sessionID]; ok {
		session.UpdatedAt = time.Now()
	}

	return nil
}

// GetMessages retrieves all messages for a session.
func (s *MemoryStore) GetMessages(ctx context.Context, sessionID string) ([]*protocol.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages, ok := s.messages[sessionID]
	if !ok {
		return nil, protocol.ErrSessionNotFound
	}

	// Return a copy to prevent external modification
	result := make([]*protocol.Message, len(messages))
	copy(result, messages)
	return result, nil
}

// UpdateMessage updates an existing message by ID.
func (s *MemoryStore) UpdateMessage(ctx context.Context, sessionID string, message *protocol.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	messages, ok := s.messages[sessionID]
	if !ok {
		return protocol.ErrSessionNotFound
	}

	for i, msg := range messages {
		if msg.ID == message.ID {
			s.messages[sessionID][i] = message
			return nil
		}
	}

	return protocol.NewError(protocol.ErrCodeInternal, "message not found")
}

// Close is a no-op for memory store.
func (s *MemoryStore) Close() error {
	return nil
}
