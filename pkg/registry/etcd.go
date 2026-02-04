package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	// DefaultPrefix is the default key prefix for agent registration.
	DefaultPrefix = "/ai-nexus/agents/"
	// DefaultTTL is the default lease TTL in seconds.
	DefaultTTL = 30
)

// ETCDConfig holds configuration for ETCD registry.
type ETCDConfig struct {
	Endpoints   []string      `yaml:"endpoints"`
	Prefix      string        `yaml:"prefix"`
	TTL         int64         `yaml:"ttl"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
	Username    string        `yaml:"username"`
	Password    string        `yaml:"password"`
}

// ETCDRegistry implements Registry using ETCD.
type ETCDRegistry struct {
	client  *clientv3.Client
	config  ETCDConfig
	leaseID clientv3.LeaseID

	mu     sync.RWMutex
	cancel context.CancelFunc
}

// NewETCDRegistry creates a new ETCD registry.
func NewETCDRegistry(config ETCDConfig) (*ETCDRegistry, error) {
	if config.Prefix == "" {
		config.Prefix = DefaultPrefix
	}
	if config.TTL == 0 {
		config.TTL = DefaultTTL
	}
	if config.DialTimeout == 0 {
		config.DialTimeout = 5 * time.Second
	}

	clientConfig := clientv3.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: config.DialTimeout,
	}
	if config.Username != "" {
		clientConfig.Username = config.Username
		clientConfig.Password = config.Password
	}

	client, err := clientv3.New(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &ETCDRegistry{
		client: client,
		config: config,
	}, nil
}

// Register registers an agent with a lease for automatic cleanup.
func (r *ETCDRegistry) Register(ctx context.Context, card *protocol.AgentCard) error {
	// Create a lease
	lease, err := r.client.Grant(ctx, r.config.TTL)
	if err != nil {
		return fmt.Errorf("failed to create lease: %w", err)
	}
	r.leaseID = lease.ID

	// Start keep-alive
	keepAliveCh, err := r.client.KeepAlive(ctx, r.leaseID)
	if err != nil {
		return fmt.Errorf("failed to start keep-alive: %w", err)
	}

	// Consume keep-alive responses in background
	go func() {
		for range keepAliveCh {
			// Keep consuming to prevent blocking
		}
	}()

	// Serialize agent card
	data, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("failed to marshal agent card: %w", err)
	}

	// Put with lease
	key := r.config.Prefix + card.Name
	_, err = r.client.Put(ctx, key, string(data), clientv3.WithLease(r.leaseID))
	if err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}

	return nil
}

// Deregister removes an agent from the registry.
func (r *ETCDRegistry) Deregister(ctx context.Context, name string) error {
	key := r.config.Prefix + name
	_, err := r.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to deregister agent: %w", err)
	}

	// Revoke lease if exists
	if r.leaseID != 0 {
		_, _ = r.client.Revoke(ctx, r.leaseID)
	}

	return nil
}

// GetAgent retrieves an agent by name.
func (r *ETCDRegistry) GetAgent(ctx context.Context, name string) (*protocol.AgentCard, error) {
	key := r.config.Prefix + name
	resp, err := r.client.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, protocol.ErrAgentNotFound
	}

	var card protocol.AgentCard
	if err := json.Unmarshal(resp.Kvs[0].Value, &card); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent card: %w", err)
	}

	return &card, nil
}

// ListAgents lists all registered agents.
func (r *ETCDRegistry) ListAgents(ctx context.Context) ([]*protocol.AgentCard, error) {
	resp, err := r.client.Get(ctx, r.config.Prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	agents := make([]*protocol.AgentCard, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var card protocol.AgentCard
		if err := json.Unmarshal(kv.Value, &card); err != nil {
			continue // Skip invalid entries
		}
		agents = append(agents, &card)
	}

	return agents, nil
}

// Watch watches for registry changes.
func (r *ETCDRegistry) Watch(ctx context.Context, callback func(WatchEvent)) error {
	watchCtx, cancel := context.WithCancel(ctx)
	r.mu.Lock()
	r.cancel = cancel
	r.mu.Unlock()

	watchCh := r.client.Watch(watchCtx, r.config.Prefix, clientv3.WithPrefix())

	for watchResp := range watchCh {
		for _, event := range watchResp.Events {
			var eventType EventType
			switch event.Type {
			case clientv3.EventTypePut:
				if event.IsCreate() {
					eventType = EventTypeAdd
				} else {
					eventType = EventTypeUpdate
				}
			case clientv3.EventTypeDelete:
				eventType = EventTypeDelete
			}

			name := ""
			if event.Kv != nil {
				key := string(event.Kv.Key)
				name = strings.TrimPrefix(key, r.config.Prefix)
			}

			var card *protocol.AgentCard
			if event.Type != clientv3.EventTypeDelete {
				card = &protocol.AgentCard{}
				if err := json.Unmarshal(event.Kv.Value, card); err != nil {
					continue
				}
				if name == "" {
					name = card.Name
				}
			}

			callback(WatchEvent{Type: eventType, Name: name, Agent: card})
		}
	}

	return nil
}

// Close closes the registry connection.
func (r *ETCDRegistry) Close() error {
	r.mu.Lock()
	if r.cancel != nil {
		r.cancel()
	}
	r.mu.Unlock()

	return r.client.Close()
}
