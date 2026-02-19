package main

// WebUI HTTP server.

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server is the main WebUI server.
type Server struct {
	config    *Config
	discovery *Discovery
	sessions  *SessionManager
	router    *Router
	sse       *SSEManager
	handlers  *Handlers
	server    *http.Server
}

// NewServer creates a new WebUI server.
func NewServer(config *Config) (*Server, error) {
	// Initialize discovery
	discovery, err := NewDiscovery(config.Registry)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery: %w", err)
	}

	// Initialize session manager
	sessions, err := NewSessionManager(config.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session manager: %w", err)
	}

	// Initialize SSE manager
	sse := NewSSEManager()

	// Initialize router
	router := NewRouter(discovery, sessions, config)

	// Initialize handlers
	handlers := NewHandlers(discovery, sessions, router, sse, config)

	return &Server{
		config:    config,
		discovery: discovery,
		sessions:  sessions,
		router:    router,
		sse:       sse,
		handlers:  handlers,
	}, nil
}

// Setup sets up the HTTP routes.
func (s *Server) Setup() *http.ServeMux {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/agents", s.cors(s.handlers.HandleAgents))
	mux.HandleFunc("/api/sessions", s.cors(s.handleSessionsRoute))
	mux.HandleFunc("/api/sessions/", s.cors(s.handlers.HandleSession))
	mux.HandleFunc("/api/chat", s.cors(s.handlers.HandleChat))
	mux.HandleFunc("/api/callback", s.cors(s.handlers.HandleCallback))
	mux.HandleFunc("/api/route", s.cors(s.handlers.HandleRoute))
	mux.HandleFunc("/api/sse", s.cors(s.handlers.HandleSSE))

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","name":"%s"}`, s.config.Name)
	})

	// Static files with no-cache headers
	staticDir := "./static"
	if _, err := os.Stat(staticDir); err == nil {
		fs := http.FileServer(http.Dir(staticDir))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Add no-cache headers for development
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			fs.ServeHTTP(w, r)
		})
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				w.Header().Set("Content-Type", "text/html")
				fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>AI-Nexus WebUI</title></head>
<body>
<h1>AI-Nexus WebUI</h1>
<p>Static files not found. Please add files to ./static directory.</p>
</body>
</html>`)
				return
			}
			http.NotFound(w, r)
		})
	}

	return mux
}

// handleSessionsRoute routes to appropriate session handler.
func (s *Server) handleSessionsRoute(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handlers.HandleSessions(w, r)
	case http.MethodPost:
		s.handlers.HandleCreateSession(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// cors adds CORS headers.
func (s *Server) cors(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		origin := "*"
		if len(s.config.Server.CORSOrigins) > 0 && s.config.Server.CORSOrigins[0] != "*" {
			origin = s.config.Server.CORSOrigins[0]
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, r)
	}
}

// Run starts the server.
func (s *Server) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start discovery
	if err := s.discovery.Start(ctx); err != nil {
		log.Printf("Warning: Failed to start discovery: %v", err)
	}

	// Setup HTTP server
	mux := s.Setup()
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      mux,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("WebUI starting on http://%s:%d", s.config.Host, s.config.Port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	s.Close()
	return s.server.Shutdown(shutdownCtx)
}

// Close closes all resources.
func (s *Server) Close() {
	s.discovery.Close()
	s.sessions.Close()
}
