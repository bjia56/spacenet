package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Server represents the server for spacenet
type Server struct {
	store         Store
	httpServer    *http.Server
	httpPort      int
	httpHandler   *HTTPHandler
	httpPortReady chan int
}

// ServerOptions holds configuration options for the server
type ServerOptions struct {
	HTTPPort int
	DBPath   string // Path to SQLite database file
}

// NewServerWithOptions creates a new spacenet server instance with custom options
func NewServerWithOptions(opts ServerOptions) *Server {
	var store Store
	var err error

	if opts.DBPath == "" {
		store = NewClaimStore()
	} else {
		// Use ClaimStore with SQLite backend
		store, err = NewClaimStoreWithSQLite(opts.DBPath)
		if err != nil {
			log.Fatalf("Failed to open SQLite database at %s: %v", opts.DBPath, err)
		}
	}

	// Create HTTP handler for API endpoints
	httpHandler := NewHTTPHandler(store)

	return &Server{
		store:         store,
		httpPort:      opts.HTTPPort,
		httpHandler:   httpHandler,
		httpPortReady: make(chan int, 1),
	}
}

// Start starts the spacenet server
func (s *Server) Start() error {
	// Start HTTP server for API endpoints
	if err := s.startHTTPServer(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// startHTTPServer starts the HTTP server for the API
func (s *Server) startHTTPServer() error {
	router := mux.NewRouter()
	s.httpHandler.RegisterRoutes(router)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.httpPort),
		Handler: router,
	}

	// Start the HTTP server in a goroutine
	go func() {
		listener, err := net.Listen("tcp", s.httpServer.Addr)
		if err != nil {
			log.Printf("Failed to create HTTP listener: %v", err)
			return
		}

		// Update httpPort with the actual assigned port if using ephemeral port (0)
		if s.httpPort == 0 {
			s.httpPort = listener.Addr().(*net.TCPAddr).Port
		}

		// Notify that HTTP port is ready
		select {
		case s.httpPortReady <- s.httpPort:
		default:
			// Channel already has a value, which is fine
		}

		log.Printf("SpaceNet HTTP server listening on :%d", s.httpPort)
		if err := s.httpServer.Serve(listener); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// WaitForHTTPPort waits for the HTTP server to be ready and returns the port
func (s *Server) WaitForHTTPPort(timeout time.Duration) (int, error) {
	select {
	case port := <-s.httpPortReady:
		return port, nil
	case <-time.After(timeout):
		return 0, fmt.Errorf("timeout waiting for HTTP port")
	}
}

// Stop stops all server components
func (s *Server) Stop() {
	s.stopHTTPServer()

	if s.store != nil {
		if err := s.store.Close(); err != nil {
			log.Printf("Error closing store during shutdown: %v", err)
		}
	}
}

// stopHTTPServer stops the HTTP server
func (s *Server) stopHTTPServer() {
	if s.httpServer != nil {
		// Create a context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down HTTP server: %v", err)
		}

		s.httpServer = nil
	}
}

