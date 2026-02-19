package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
	"github.com/AdrianWangs/ai-nexus/pkg/tools"
)

// Server is the HTTP server for an agent.
type Server struct {
	agent   *Agent
	handler *Handler
	server  *http.Server
}

// NewServer creates a new server for the agent.
func NewServer(agent *Agent) *Server {
	s := &Server{
		agent:   agent,
		handler: NewHandler(agent),
	}

	mux := http.NewServeMux()

	// A2A endpoints
	mux.HandleFunc("/a2a/tasks", s.handleTasks)
	mux.HandleFunc("/a2a/tasks/stream", s.handleTasksStream)
	mux.HandleFunc("/a2a/card", s.handleCard)
	mux.HandleFunc("/a2a/capabilities", s.handleCapabilities)
	mux.HandleFunc("/a2a/tools", s.handleTools)

	// Health check
	mux.HandleFunc("/health", s.handleHealth)

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", agent.config.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	return s
}

// handleTasks handles non-streaming task requests.
func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	resp, err := s.handler.HandleTask(ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleTasksStream handles streaming task requests.
func (s *Server) handleTasksStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

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

	ctx := r.Context()

	sendEvent := func(event *protocol.StreamEvent) error {
		data, err := json.Marshal(event)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
		return nil
	}

	if err := s.handler.HandleTaskStream(ctx, &req, sendEvent); err != nil {
		sendEvent(&protocol.StreamEvent{
			Type:  protocol.EventTypeError,
			Error: err.Error(),
		})
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

// handleCard returns the agent's card.
func (s *Server) handleCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.agent.GetCard())
}

// handleCapabilities returns the agent's capability index (for tool lazy-loading).
func (s *Server) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idx := s.agent.Tools().CapabilityIndex()
	out := make([]protocol.Capability, 0, len(idx))
	for _, c := range idx {
		out = append(out, protocol.Capability{Name: c.Name, Description: c.Description, Keywords: c.Keywords})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// handleTools returns tools under a capability (for tool lazy-loading).
func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	capability := r.URL.Query().Get("capability")
	var list []*tools.ToolDefinition
	if capability == "" {
		list = s.agent.Tools().List()
	} else {
		list = s.agent.Tools().ListByCapability(capability)
	}

	toolsOut := make([]protocol.Tool, 0, len(list))
	for _, t := range list {
		toolsOut = append(toolsOut, protocol.Tool{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  t.Parameters,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toolsOut)
}

// handleHealth handles health check requests.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"name":   s.agent.config.Name,
	})
}

// Run starts the server and blocks until shutdown.
func (s *Server) Run() error {
	// Initialize and register with registry
	if err := s.agent.InitRegistry(); err != nil {
		log.Printf("Warning: Failed to initialize registry: %v", err)
	} else if err := s.agent.Register(); err != nil {
		log.Printf("Warning: Failed to register agent: %v", err)
	} else if err := s.agent.StartDiscovery(); err != nil {
		log.Printf("Warning: Failed to start discovery: %v", err)
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Agent %s starting on port %d", s.agent.config.Name, s.agent.config.Port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.agent.Close()
	return s.server.Shutdown(ctx)
}

// RunAgent is a convenience function to create and run an agent.
func RunAgent(configPath string, setup func(*Agent) error) error {
	agent, err := NewFromConfig(configPath)
	if err != nil {
		return err
	}

	if setup != nil {
		if err := setup(agent); err != nil {
			return err
		}
	}

	server := NewServer(agent)
	return server.Run()
}
