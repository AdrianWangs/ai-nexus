package main

// HTTP handlers for WebUI.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
	"github.com/google/uuid"
)

// Handlers holds all HTTP handlers.
type Handlers struct {
	discovery *Discovery
	sessions  *SessionManager
	router    *Router
	sse       *SSEManager
	config    *Config
}

// NewHandlers creates new handlers.
func NewHandlers(discovery *Discovery, sessions *SessionManager, router *Router, sse *SSEManager, config *Config) *Handlers {
	return &Handlers{
		discovery: discovery,
		sessions:  sessions,
		router:    router,
		sse:       sse,
		config:    config,
	}
}

// HandleAgents returns the list of online agents.
func (h *Handlers) HandleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agents := h.discovery.ListAgents()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// HandleSessions returns the list of sessions.
func (h *Handlers) HandleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessions, err := h.sessions.ListSessions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// HandleSession handles individual session operations.
func (h *Handlers) HandleSession(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}
	sessionID := parts[0]

	switch r.Method {
	case http.MethodGet:
		// Check if requesting messages or stream
		if len(parts) > 1 && parts[1] == "messages" {
			h.handleSessionMessages(w, r, sessionID)
			return
		}
		if len(parts) > 1 && parts[1] == "stream" {
			h.handleSessionStream(w, r, sessionID)
			return
		}

		sess, err := h.sessions.GetSession(r.Context(), sessionID)
		if err != nil {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sess)

	case http.MethodPost:
		// Post a message to session
		if len(parts) > 1 && parts[1] == "messages" {
			h.handlePostMessage(w, r, sessionID)
			return
		}
		http.Error(w, "Invalid path", http.StatusBadRequest)

	case http.MethodDelete:
		if err := h.sessions.DeleteSession(r.Context(), sessionID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSessionMessages returns messages for a session.
func (h *Handlers) handleSessionMessages(w http.ResponseWriter, r *http.Request, sessionID string) {
	messages, err := h.sessions.GetMessages(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// handlePostMessage handles posting a message to a session.
func (h *Handlers) handlePostMessage(w http.ResponseWriter, r *http.Request, sessionID string) {
	var req struct {
		Content     string `json:"content"`
		TargetAgent string `json:"target_agent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get or create session
	sess, err := h.sessions.GetOrCreateSession(ctx, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine target agent
	targetAgent := h.parseTargetAgent(req.Content, req.TargetAgent)
	if targetAgent == "" {
		// Default to orchestrator if available
		agents := h.discovery.ListAgents()
		for _, agent := range agents {
			if agent.Type == protocol.AgentTypeOrchestrator {
				targetAgent = agent.Name
				break
			}
		}
		if targetAgent == "" && len(agents) > 0 {
			targetAgent = agents[0].Name
		}
	}

	if targetAgent == "" {
		http.Error(w, "No available agents", http.StatusServiceUnavailable)
		return
	}

	// Add user message to session
	userMsg := &protocol.Message{
		ID:        uuid.New().String(),
		SessionID: sess.ID,
		From:      "user",
		To:        "@" + targetAgent,
		Content:   req.Content,
		Type:      protocol.MessageTypeUser,
		Timestamp: time.Now(),
	}
	h.sessions.AddMessage(ctx, sess.ID, userMsg)

	// Route to agent asynchronously
	go func() {
		var contentBuffer strings.Builder
		err := h.router.RouteToAgent(context.Background(), sess.ID, targetAgent, req.Content, "user", func(event *protocol.StreamEvent) error {
			// Accumulate content
			if event.Type == protocol.EventTypeContentDelta {
				contentBuffer.WriteString(event.Content)
			}

			// Broadcast to SSE subscribers
			h.sse.Broadcast(sess.ID, event)
			return nil
		})

		if err != nil {
			h.sse.Broadcast(sess.ID, &protocol.StreamEvent{
				Type:  protocol.EventTypeError,
				Error: err.Error(),
			})
			return
		}

		// Save complete message if we have content
		if contentBuffer.Len() > 0 {
			h.sessions.AddMessage(context.Background(), sess.ID, &protocol.Message{
				ID:        uuid.New().String(),
				SessionID: sess.ID,
				From:      targetAgent,
				To:        "user",
				Content:   contentBuffer.String(),
				Type:      protocol.MessageTypeReply,
				Timestamp: time.Now(),
			})
		}

		// Send done event
		h.sse.Broadcast(sess.ID, &protocol.StreamEvent{
			Type: protocol.EventTypeDone,
		})
	}()

	// Return immediately with accepted status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "accepted",
		"session": sess.ID,
		"target":  targetAgent,
	})
}

// handleSessionStream handles SSE stream for a session.
func (h *Handlers) handleSessionStream(w http.ResponseWriter, r *http.Request, sessionID string) {
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

	// Create event channel
	eventChan := make(chan *protocol.StreamEvent, 100)
	h.sse.Subscribe(sessionID, eventChan)
	defer h.sse.Unsubscribe(sessionID, eventChan)

	// Send connected event
	sendEvent(w, flusher, &protocol.StreamEvent{
		Type:      protocol.EventTypeStatus,
		SessionID: sessionID,
		Status:    "connected",
	})

	// Listen for events
	for {
		select {
		case event := <-eventChan:
			if err := sendEvent(w, flusher, event); err != nil {
				return
			}
			if event.Type == protocol.EventTypeDone {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}

// HandleCreateSession creates a new session.
func (h *Handlers) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Title == "" {
		req.Title = "New Chat"
	}

	sess, err := h.sessions.CreateSession(r.Context(), req.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sess)
}

// ChatRequest represents a chat request from the frontend.
type ChatRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
	Target    string `json:"target"` // Optional: @agent-name
}

// HandleChat handles chat requests with streaming response.
func (h *Handlers) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get or create session
	sess, err := h.sessions.GetOrCreateSession(ctx, req.SessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine target agent
	targetAgent := h.parseTargetAgent(req.Message, req.Target)
	if targetAgent == "" {
		// Default to orchestrator if available
		agents := h.discovery.ListAgents()
		for _, agent := range agents {
			if agent.Type == protocol.AgentTypeOrchestrator {
				targetAgent = agent.Name
				break
			}
		}
		if targetAgent == "" && len(agents) > 0 {
			targetAgent = agents[0].Name
		}
	}

	if targetAgent == "" {
		http.Error(w, "No available agents", http.StatusServiceUnavailable)
		return
	}

	// Add user message to session
	userMsg := &protocol.Message{
		ID:        uuid.New().String(),
		SessionID: sess.ID,
		From:      "user",
		To:        "@" + targetAgent,
		Content:   req.Message,
		Type:      protocol.MessageTypeUser,
		Timestamp: time.Now(),
	}
	h.sessions.AddMessage(ctx, sess.ID, userMsg)

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

	// Send session ID
	sendEvent(w, flusher, &protocol.StreamEvent{
		Type:      protocol.EventTypeStatus,
		SessionID: sess.ID,
		Status:    "started",
	})

	// Route to agent
	var contentBuffer strings.Builder
	err = h.router.RouteToAgent(ctx, sess.ID, targetAgent, req.Message, "user", func(event *protocol.StreamEvent) error {
		// Accumulate content
		if event.Type == protocol.EventTypeContentDelta {
			contentBuffer.WriteString(event.Content)
		}

		// Broadcast to other SSE subscribers
		h.sse.Broadcast(sess.ID, event)

		// Send to current response
		return sendEvent(w, flusher, event)
	})

	if err != nil {
		sendEvent(w, flusher, &protocol.StreamEvent{
			Type:  protocol.EventTypeError,
			Error: err.Error(),
		})
	}

	// Save complete message if we have content
	if contentBuffer.Len() > 0 {
		h.sessions.AddMessage(ctx, sess.ID, &protocol.Message{
			ID:        uuid.New().String(),
			SessionID: sess.ID,
			From:      "@" + targetAgent,
			To:        "user",
			Content:   contentBuffer.String(),
			Type:      protocol.MessageTypeReply,
			Timestamp: time.Now(),
		})
	}

	// Send done
	sendEvent(w, flusher, &protocol.StreamEvent{
		Type:      protocol.EventTypeDone,
		SessionID: sess.ID,
	})

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

// HandleCallback handles callbacks from agents.
func (h *Handlers) HandleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg protocol.CallbackMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.router.HandleCallback(r.Context(), &msg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast to SSE subscribers
	h.sse.Broadcast(msg.SessionID, &protocol.StreamEvent{
		Type:      protocol.EventTypeMessageEnd,
		SessionID: msg.SessionID,
		From:      msg.From,
		To:        msg.To,
		Content:   msg.Content,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HandleRoute handles route requests from agents.
func (h *Handlers) HandleRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.RouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

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

	// Handle the route request
	err := h.router.HandleRouteRequest(ctx, &req, func(event *protocol.StreamEvent) error {
		// Broadcast to SSE subscribers
		h.sse.Broadcast(req.SessionID, event)

		// Send to response
		return sendEvent(w, flusher, event)
	})

	if err != nil {
		sendEvent(w, flusher, &protocol.StreamEvent{
			Type:  protocol.EventTypeError,
			Error: err.Error(),
		})
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

// HandleSSE handles SSE subscriptions for a session.
func (h *Handlers) HandleSSE(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	h.sse.ServeHTTP(w, r, sessionID)
}

// parseTargetAgent extracts the target agent from the message or target field.
func (h *Handlers) parseTargetAgent(message, target string) string {
	// Check explicit target
	if target != "" {
		return strings.TrimPrefix(target, "@")
	}

	// Check for @mention in message
	words := strings.Fields(message)
	for _, word := range words {
		if strings.HasPrefix(word, "@") {
			agentName := strings.TrimPrefix(word, "@")
			if _, ok := h.discovery.GetAgent(agentName); ok {
				return agentName
			}
		}
	}

	return ""
}

// sendEvent sends an SSE event.
func sendEvent(w http.ResponseWriter, flusher http.Flusher, event *protocol.StreamEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	eventName := string(event.Type)

	// Send SSE event with proper format: event: xxx\ndata: yyy\n\n
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventName, data)
	flusher.Flush()
	return nil
}
