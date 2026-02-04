package main

// cmd entrypoint for orchestrator

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AdrianWangs/ai-nexus/pkg/agent"
	"github.com/AdrianWangs/ai-nexus/pkg/llm"
	"github.com/AdrianWangs/ai-nexus/pkg/orchestrator"
	"github.com/AdrianWangs/ai-nexus/pkg/protocol"
	"github.com/AdrianWangs/ai-nexus/pkg/registry"
)

func main() {
	configPath := flag.String("config", "./config.yaml", "Path to config file")
	flag.Parse()

	// Load config
	config, err := agent.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create LLM client
	llmClient := llm.NewOpenAIClient(llm.OpenAIConfig{
		APIKey:  config.LLM.APIKey,
		BaseURL: config.LLM.BaseURL,
		Model:   config.LLM.Model,
		Timeout: config.LLM.Timeout,
	})

	// Create orchestrator
	orc := orchestrator.New(&orchestrator.Config{
		WebUIURL: os.Getenv("WEBUI_URL"),
	}, llmClient)

	// Initialize registry
	var reg registry.Registry
	if len(config.Registry.Endpoints) > 0 {
		reg, err = registry.NewETCDRegistry(registry.ETCDConfig{
			Endpoints: config.Registry.Endpoints,
			Prefix:    config.Registry.Prefix,
			TTL:       config.Registry.TTL,
		})
		if err != nil {
			log.Printf("Warning: Failed to create registry: %v", err)
		} else {
			orc.SetRegistry(reg)
		}
	}

	// Create HTTP server
	mux := http.NewServeMux()

	// A2A endpoints
	mux.HandleFunc("/a2a/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req protocol.TaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		resp, err := orc.ProcessRequest(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/a2a/tasks/stream", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req protocol.TaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
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

		sendEvent := func(event *protocol.StreamEvent) error {
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			return nil
		}

		if err := orc.ProcessRequestStream(r.Context(), &req, sendEvent); err != nil {
			sendEvent(&protocol.StreamEvent{
				Type:  protocol.EventTypeError,
				Error: err.Error(),
			})
		}

		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	})

	mux.HandleFunc("/a2a/card", func(w http.ResponseWriter, r *http.Request) {
		caps := make([]protocol.Capability, 0, len(config.Capabilities))
		for _, name := range config.Capabilities {
			caps = append(caps, protocol.Capability{Name: name})
		}
		card := &protocol.AgentCard{
			Name:         config.Name,
			DisplayName:  config.DisplayName,
			Description:  config.Description,
			Version:      config.Version,
			URL:          fmt.Sprintf("http://%s:%d", config.Host, config.Port),
			Port:         config.Port,
			Type:         protocol.AgentTypeOrchestrator,
			Capabilities: caps,
			Status:       protocol.AgentStatusOnline,
		}

		// Add skills
		for _, s := range config.Skills {
			card.Skills = append(card.Skills, protocol.Skill{
				Name:        s.Name,
				Description: s.Description,
				Keywords:    s.Keywords,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(card)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "name": config.Name})
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register with ETCD
	if reg != nil {
		caps := make([]protocol.Capability, 0, len(config.Capabilities))
		for _, name := range config.Capabilities {
			caps = append(caps, protocol.Capability{Name: name})
		}
		card := &protocol.AgentCard{
			Name:         config.Name,
			DisplayName:  config.DisplayName,
			Description:  config.Description,
			Version:      config.Version,
			URL:          fmt.Sprintf("http://%s:%d", config.Host, config.Port),
			Port:         config.Port,
			Type:         protocol.AgentTypeOrchestrator,
			Capabilities: caps,
			Status:       protocol.AgentStatusOnline,
		}
		if err := reg.Register(ctx, card); err != nil {
			log.Printf("Warning: Failed to register: %v", err)
		}

		// Load and watch agents
		if err := orc.LoadAgents(ctx); err != nil {
			log.Printf("Warning: Failed to load agents: %v", err)
		}
		go orc.WatchAgents(ctx)
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Orchestrator starting on port %d", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down...")

	if reg != nil {
		reg.Deregister(ctx, config.Name)
		reg.Close()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)
}
