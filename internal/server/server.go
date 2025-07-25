package server

import (
	"log"
	"net"
)

// Server represents the UDP server for spacenet
type Server struct {
	store    Store
	listener *net.UDPConn
	port     int
}

// ServerOptions holds configuration options for the server
type ServerOptions struct {
	Port        int
	RedisAddr   string // Format: "host:port"
	UseInMemory bool   // If true, use in-memory store instead of Redis
}

// NewServer creates a new spacenet server instance with default options
func NewServer(port int) *Server {
	return NewServerWithOptions(ServerOptions{
		Port:        port,
		UseInMemory: true,
	})
}

// NewServerWithOptions creates a new spacenet server instance with custom options
func NewServerWithOptions(opts ServerOptions) *Server {
	var store Store

	if opts.UseInMemory {
		store = NewClaimStore()
	} else {
		// Use Redis
		redisStore, err := NewRedisStore(opts.RedisAddr)
		if err != nil {
			log.Fatalf("Failed to connect to Redis at %s: %v", opts.RedisAddr, err)
		}
		store = redisStore
	}

	return &Server{
		store: store,
		port:  opts.Port,
	}
}

// Start starts the spacenet server
func (s *Server) Start() error {
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

	log.Printf("SpaceNet server listening on [%s]:%d", addr.IP, s.port)

	// Start processing packets
	go s.processPackets()

	return nil
}

// Stop stops the spacenet server
func (s *Server) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}

	if s.store != nil {
		s.store.Close()
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
