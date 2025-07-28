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
	store       Store
	listener    *net.UDPConn
	httpServer  *http.Server
	port        int
	httpPort    int
	httpHandler *HTTPHandler
}

// ServerOptions holds configuration options for the server
type ServerOptions struct {
	Port      int
	HTTPPort  int
	RedisAddr string // Format: "host:port"
}

// NewServer creates a new spacenet server instance with default options
func NewServer(port int) *Server {
	return NewServerWithOptions(ServerOptions{
		Port:     port,
		HTTPPort: 8080,
	})
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
		store:       store,
		port:        opts.Port,
		httpPort:    opts.HTTPPort,
		httpHandler: httpHandler,
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
		log.Printf("SpaceNet HTTP server listening on :%d", s.httpPort)
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
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
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			for job := range jobs {
				clientIP := job.addr.IP.String()

				// Parse proof-of-work packet
				packet, err := ParseClaimPacket(job.data)
				if err != nil {
					log.Printf("Failed to parse claim packet from %s: %v", clientIP, err)
					continue
				}

				// Create proof of work object
				targetIP := net.ParseIP(clientIP)
				if targetIP == nil {
					log.Printf("Invalid IP address from client: %s", clientIP)
					continue
				}

				pow := packet.CreateProofOfWork(targetIP)

				// Validate proof of work
				if err := s.store.ValidateProofOfWork(pow); err != nil {
					log.Printf("Invalid proof of work from %s (%s): %v", clientIP, packet.Claimant, err)
					continue
				}

				// Process the claim
				err = s.store.ProcessClaim(clientIP, packet.Claimant)
				if err != nil {
					log.Printf("Failed to process claim from %s: %v", clientIP, err)
				} else {
					log.Printf("Address %s claimed by %s (nonce: %d)",
						clientIP, packet.Claimant, packet.Nonce)
				}
			}
		}(i)
	}

	// Main loop to receive UDP packets
	for {
		// Buffer for incoming packets
		buffer := make([]byte, 32) // Max 32-byte payload as per requirements

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
