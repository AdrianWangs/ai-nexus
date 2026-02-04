package main

// Session manager for WebUI.

import (
	"context"
	"time"

	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
	"github.com/AdrianWangs/ai-nexus/pkg/session"
	"github.com/google/uuid"
)

// SessionManager manages chat sessions.
type SessionManager struct {
	store session.Store
}

// NewSessionManager creates a new session manager.
func NewSessionManager(config SessionConfig) (*SessionManager, error) {
	var store session.Store
	var err error

	switch config.Type {
	case "redis":
		store, err = session.NewRedisStore(session.RedisConfig{
			Addr:     config.Addr,
			Password: config.Password,
			DB:       config.DB,
			TTL:      config.TTL,
		})
	default:
		store = session.NewMemoryStore()
	}

	if err != nil {
		return nil, err
	}

	return &SessionManager{store: store}, nil
}

// CreateSession creates a new session.
func (m *SessionManager) CreateSession(ctx context.Context, title string) (*protocol.Session, error) {
	sess := &protocol.Session{
		ID:           uuid.New().String(),
		Title:        title,
		Participants: []string{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := m.store.CreateSession(ctx, sess); err != nil {
		return nil, err
	}

	return sess, nil
}

// GetSession retrieves a session by ID.
func (m *SessionManager) GetSession(ctx context.Context, sessionID string) (*protocol.Session, error) {
	return m.store.GetSession(ctx, sessionID)
}

// UpdateSession updates a session.
func (m *SessionManager) UpdateSession(ctx context.Context, sess *protocol.Session) error {
	sess.UpdatedAt = time.Now()
	return m.store.UpdateSession(ctx, sess)
}

// DeleteSession deletes a session.
func (m *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	return m.store.DeleteSession(ctx, sessionID)
}

// ListSessions lists all sessions.
func (m *SessionManager) ListSessions(ctx context.Context) ([]*protocol.Session, error) {
	return m.store.ListSessions(ctx)
}

// AddMessage adds a message to a session.
func (m *SessionManager) AddMessage(ctx context.Context, sessionID string, msg *protocol.Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	msg.SessionID = sessionID

	// Update participants
	sess, err := m.store.GetSession(ctx, sessionID)
	if err == nil {
		updated := false
		for _, p := range []string{msg.From, msg.To} {
			if p != "" && p != "user" {
				found := false
				for _, existing := range sess.Participants {
					if existing == p {
						found = true
						break
					}
				}
				if !found {
					sess.Participants = append(sess.Participants, p)
					updated = true
				}
			}
		}
		if updated {
			m.store.UpdateSession(ctx, sess)
		}
	}

	return m.store.AddMessage(ctx, sessionID, msg)
}

// GetMessages retrieves all messages for a session.
func (m *SessionManager) GetMessages(ctx context.Context, sessionID string) ([]*protocol.Message, error) {
	return m.store.GetMessages(ctx, sessionID)
}

// UpdateMessage updates an existing message.
func (m *SessionManager) UpdateMessage(ctx context.Context, sessionID string, msg *protocol.Message) error {
	return m.store.UpdateMessage(ctx, sessionID, msg)
}

// Close closes the session manager.
func (m *SessionManager) Close() error {
	return m.store.Close()
}

// GetOrCreateSession gets an existing session or creates a new one.
func (m *SessionManager) GetOrCreateSession(ctx context.Context, sessionID string) (*protocol.Session, error) {
	if sessionID == "" {
		return m.CreateSession(ctx, "New Chat")
	}

	sess, err := m.GetSession(ctx, sessionID)
	if err == nil {
		return sess, nil
	}

	// Create new session with the given ID
	sess = &protocol.Session{
		ID:           sessionID,
		Title:        "New Chat",
		Participants: []string{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := m.store.CreateSession(ctx, sess); err != nil {
		return nil, err
	}

	return sess, nil
}
