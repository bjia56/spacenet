package server

import (
	"log"
	"net"
	"sync"
)

// ClaimStore is an in-memory store for IP address claims
type ClaimStore struct {
	mutex  sync.RWMutex
	claims map[string]string // map[ipAddress]claimantName
}

// NewClaimStore creates a new in-memory claim store
func NewClaimStore() *ClaimStore {
	return &ClaimStore{
		claims: make(map[string]string),
	}
}

// ProcessClaim processes a claim request and updates the store
func (cs *ClaimStore) ProcessClaim(ipAddr string, claimant string) (bool, string) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if existingClaimant, exists := cs.claims[ipAddr]; exists {
		// Address already claimed
		return false, existingClaimant
	}

	// Store new claim
	cs.claims[ipAddr] = claimant
	return true, ""
}

// GetClaim retrieves the claimant for an IP address
func (cs *ClaimStore) GetClaim(ipAddr string) (string, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	claimant, exists := cs.claims[ipAddr]
	return claimant, exists
}

// GetAllClaims returns all claims in the store
func (cs *ClaimStore) GetAllClaims() map[string]string {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	// Create a copy to avoid concurrent access issues
	claims := make(map[string]string)
	for ip, claimant := range cs.claims {
		claims[ip] = claimant
	}

	return claims
}

// Server represents the UDP server for spacenet
type Server struct {
	store    *ClaimStore
	listener *net.UDPConn
	port     int
}

// NewServer creates a new spacenet server instance
func NewServer(port int) *Server {
	return &Server{
		store: NewClaimStore(),
		port:  port,
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

				success, existingClaimant := s.store.ProcessClaim(clientIP, claimantName)

				if success {
					log.Printf("Address %s claimed by %s", clientIP, claimantName)
				} else {
					log.Printf("Address %s already claimed by %s, rejected claim from %s",
						clientIP, existingClaimant, claimantName)
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
