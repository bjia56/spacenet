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
	store          Store
	listener       *net.UDPConn
	httpServer     *http.Server
	port           int
	httpPort       int
	httpHandler    *HTTPHandler
	udpPortReady   chan int
	httpPortReady  chan int
}

// ServerOptions holds configuration options for the server
type ServerOptions struct {
	Port      int
	HTTPPort  int
	RedisAddr string // Format: "host:port"
}

// NewServerWithOptions creates a new spacenet server instance with custom options
func NewServerWithOptions(opts ServerOptions) *Server {
	var store Store
	var err error

	if opts.RedisAddr == "" {
		store = NewClaimStore()
	} else {
		// Use ClaimStore with Redis backend
		store, err = NewClaimStoreWithRedis(opts.RedisAddr)
		if err != nil {
			log.Fatalf("Failed to connect to Redis at %s: %v", opts.RedisAddr, err)
		}
	}

	// Create HTTP handler for API endpoints
	httpHandler := NewHTTPHandler(store)

	return &Server{
		store:         store,
		port:          opts.Port,
		httpPort:      opts.HTTPPort,
		httpHandler:   httpHandler,
		udpPortReady:  make(chan int, 1),
		httpPortReady: make(chan int, 1),
	}
}

// Start starts the spacenet server
func (s *Server) Start() error {
	// Start UDP server for claims
	if err := s.startUDPServer(); err != nil {
		return fmt.Errorf("failed to start UDP server: %w", err)
	}

	// Start HTTP server for API endpoints
	if err := s.startHTTPServer(); err != nil {
		s.stopUDPServer() // Clean up UDP server if HTTP fails
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// startUDPServer starts the UDP server for receiving claims
func (s *Server) startUDPServer() error {
	// Listen for UDP packets on the specified port
	addr := net.UDPAddr{
		Port: s.port,
		IP:   net.ParseIP("::"), // Listen on all IPv6 interfaces
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return err
	}
	s.listener = conn

	// Update port with the actual assigned port if using ephemeral port (0)
	if s.port == 0 {
		s.port = conn.LocalAddr().(*net.UDPAddr).Port
	}

	// Notify that UDP port is ready
	select {
	case s.udpPortReady <- s.port:
	default:
		// Channel already has a value, which is fine
	}

	log.Printf("SpaceNet UDP server listening on [%s]:%d", addr.IP, s.port)

	// Start processing packets
	go s.processPackets()

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

// WaitForUDPPort waits for the UDP server to be ready and returns the port
func (s *Server) WaitForUDPPort(timeout time.Duration) (int, error) {
	select {
	case port := <-s.udpPortReady:
		return port, nil
	case <-time.After(timeout):
		return 0, fmt.Errorf("timeout waiting for UDP port")
	}
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
	s.stopUDPServer()

	if s.store != nil {
		s.store.Close()
	}
}

// stopUDPServer stops the UDP server
func (s *Server) stopUDPServer() {
	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
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

// processPackets handles incoming UDP packets
func (s *Server) processPackets() {
	// Use a worker pool to handle a high volume of requests
	const numWorkers = 10
	type claimJob struct {
		addr *net.UDPAddr
		data []byte
	}
	jobs := make(chan claimJob, 100)

	// Start worker goroutines
	for i := range numWorkers {
		go func(workerID int) {
			for job := range jobs {
				// Process the claim
				clientIP := job.addr.IP.String()
				claimantName := string(job.data)
				s.store.ProcessClaim(clientIP, claimantName)
				log.Printf("Address %s claimed by %s", clientIP, claimantName)
			}
		}(i)
	}

	// Main loop to receive UDP packets
	for {
		// Buffer for incoming packets - read more than max to detect oversized packets
		buffer := make([]byte, 64) // Read up to 64 bytes to detect oversized packets

		n, clientAddr, err := s.listener.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading UDP packet: %v", err)
			// If the listener was closed, exit the loop
			if _, ok := err.(net.Error); ok {
				log.Println("Listener closed, stopping packet processing")
				close(jobs)
				return
			}
			continue
		}

		// Ensure the payload is not empty and doesn't exceed max length
		if n == 0 {
			log.Printf("Empty claim from %s, ignoring", clientAddr.IP.String())
			continue
		}

		if n > 32 {
			log.Printf("Claim payload too large from %s (%d bytes), ignoring", clientAddr.IP.String(), n)
			continue
		}

		// Queue the address and data for processing by a worker
		job := claimJob{
			addr: clientAddr,
			data: buffer[:n],
		}
		select {
		case jobs <- job:
			// Job queued successfully
		default:
			// Queue is full, log and continue
			log.Printf("Worker queue is full, dropping packet from %s", clientAddr.IP.String())
		}
	}
}
