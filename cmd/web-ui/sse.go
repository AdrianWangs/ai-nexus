package main

// SSE manager for WebUI.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
)

// SSEManager manages SSE connections for real-time updates.
type SSEManager struct {
	mu      sync.RWMutex
	clients map[string]map[chan *protocol.StreamEvent]bool // sessionID -> channels
}

// NewSSEManager creates a new SSE manager.
func NewSSEManager() *SSEManager {
	return &SSEManager{
		clients: make(map[string]map[chan *protocol.StreamEvent]bool),
	}
}

// Subscribe subscribes to a session's events with a provided channel.
func (m *SSEManager) Subscribe(sessionID string, ch chan *protocol.StreamEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.clients[sessionID] == nil {
		m.clients[sessionID] = make(map[chan *protocol.StreamEvent]bool)
	}
	m.clients[sessionID][ch] = true
}

// SubscribeNew subscribes to a session's events and returns a new channel.
func (m *SSEManager) SubscribeNew(sessionID string) chan *protocol.StreamEvent {
	ch := make(chan *protocol.StreamEvent, 100)
	m.Subscribe(sessionID, ch)
	return ch
}

// Unsubscribe unsubscribes from a session's events.
func (m *SSEManager) Unsubscribe(sessionID string, ch chan *protocol.StreamEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.clients[sessionID] != nil {
		delete(m.clients[sessionID], ch)
		close(ch)

		if len(m.clients[sessionID]) == 0 {
			delete(m.clients, sessionID)
		}
	}
}

// Broadcast sends an event to all subscribers of a session.
func (m *SSEManager) Broadcast(sessionID string, event *protocol.StreamEvent) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if clients, ok := m.clients[sessionID]; ok {
		for ch := range clients {
			select {
			case ch <- event:
			default:
				// Channel full, skip
			}
		}
	}
}

// ServeHTTP handles SSE connections for a session.
func (m *SSEManager) ServeHTTP(w http.ResponseWriter, r *http.Request, sessionID string) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Subscribe to events
	ch := m.SubscribeNew(sessionID)
	defer m.Unsubscribe(sessionID, ch)

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\ndata: {\"type\":\"connected\",\"session_id\":\"%s\"}\n\n", sessionID)
	flusher.Flush()

	// Stream events
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return
			}

			data, err := json.Marshal(event)
			if err != nil {
				continue
			}

			// Send with event type for proper event listener matching
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
			flusher.Flush()

		case <-r.Context().Done():
			return
		}
	}
}
