package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
	"github.com/redis/go-redis/v9"
)

const (
	// Key prefixes
	sessionPrefix  = "ai-nexus:session:"
	messagesPrefix = "ai-nexus:messages:"
	sessionsKey    = "ai-nexus:sessions"

	// Default TTL for sessions (7 days)
	defaultSessionTTL = 7 * 24 * time.Hour
)

// RedisConfig holds configuration for Redis store.
type RedisConfig struct {
	Addr     string        `yaml:"addr"`
	Password string        `yaml:"password"`
	DB       int           `yaml:"db"`
	TTL      time.Duration `yaml:"ttl"`
}

// RedisStore implements Store using Redis.
type RedisStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisStore creates a new Redis store.
func NewRedisStore(config RedisConfig) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	ttl := config.TTL
	if ttl == 0 {
		ttl = defaultSessionTTL
	}

	return &RedisStore{
		client: client,
		ttl:    ttl,
	}, nil
}

// CreateSession creates a new session.
func (s *RedisStore) CreateSession(ctx context.Context, session *protocol.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := sessionPrefix + session.ID
	pipe := s.client.Pipeline()
	pipe.Set(ctx, key, data, s.ttl)
	pipe.SAdd(ctx, sessionsKey, session.ID)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSession retrieves a session by ID.
func (s *RedisStore) GetSession(ctx context.Context, sessionID string) (*protocol.Session, error) {
	key := sessionPrefix + sessionID
	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, protocol.ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session protocol.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// UpdateSession updates an existing session.
func (s *RedisStore) UpdateSession(ctx context.Context, session *protocol.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := sessionPrefix + session.ID
	if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// DeleteSession deletes a session and its messages.
func (s *RedisStore) DeleteSession(ctx context.Context, sessionID string) error {
	sessionKey := sessionPrefix + sessionID
	messagesKey := messagesPrefix + sessionID

	pipe := s.client.Pipeline()
	pipe.Del(ctx, sessionKey)
	pipe.Del(ctx, messagesKey)
	pipe.SRem(ctx, sessionsKey, sessionID)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// ListSessions lists all sessions.
func (s *RedisStore) ListSessions(ctx context.Context) ([]*protocol.Session, error) {
	sessionIDs, err := s.client.SMembers(ctx, sessionsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list session IDs: %w", err)
	}

	if len(sessionIDs) == 0 {
		return []*protocol.Session{}, nil
	}

	// Batch get sessions
	keys := make([]string, len(sessionIDs))
	for i, id := range sessionIDs {
		keys[i] = sessionPrefix + id
	}

	values, err := s.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	sessions := make([]*protocol.Session, 0, len(values))
	for _, v := range values {
		if v == nil {
			continue
		}
		var session protocol.Session
		if err := json.Unmarshal([]byte(v.(string)), &session); err != nil {
			continue
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// AddMessage adds a message to a session.
func (s *RedisStore) AddMessage(ctx context.Context, sessionID string, message *protocol.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	key := messagesPrefix + sessionID
	if err := s.client.RPush(ctx, key, data).Err(); err != nil {
		return fmt.Errorf("failed to add message: %w", err)
	}

	// Extend TTL
	s.client.Expire(ctx, key, s.ttl)

	// Update session timestamp
	session, err := s.GetSession(ctx, sessionID)
	if err == nil {
		session.UpdatedAt = message.Timestamp
		s.UpdateSession(ctx, session)
	}

	return nil
}

// GetMessages retrieves all messages for a session.
func (s *RedisStore) GetMessages(ctx context.Context, sessionID string) ([]*protocol.Message, error) {
	key := messagesPrefix + sessionID
	values, err := s.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	messages := make([]*protocol.Message, 0, len(values))
	for _, v := range values {
		var msg protocol.Message
		if err := json.Unmarshal([]byte(v), &msg); err != nil {
			continue
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// UpdateMessage updates an existing message by ID.
func (s *RedisStore) UpdateMessage(ctx context.Context, sessionID string, message *protocol.Message) error {
	key := messagesPrefix + sessionID

	// Get all messages and find the one to update
	messages, err := s.GetMessages(ctx, sessionID)
	if err != nil {
		return err
	}

	// Find and update the message
	for i, msg := range messages {
		if msg.ID == message.ID {
			data, err := json.Marshal(message)
			if err != nil {
				return fmt.Errorf("failed to marshal message: %w", err)
			}
			if err := s.client.LSet(ctx, key, int64(i), data).Err(); err != nil {
				return fmt.Errorf("failed to update message: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("message not found: %s", message.ID)
}

// Close closes the Redis connection.
func (s *RedisStore) Close() error {
	return s.client.Close()
}
