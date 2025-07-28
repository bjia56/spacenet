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

// InjectableServer is an enhanced version of Server that supports dependency injection
// This makes the server more testable by allowing injection of mock dependencies
type InjectableServer struct {
	store       Store
	httpHandler *HTTPHandler
	listener    *net.UDPConn
	httpServer  *http.Server
	port        int
	httpPort    int
	
	// Injectable dependencies
	storeFactory    func() Store
	handlerFactory  func(Store) *HTTPHandler
	listenerFactory func(string, int) (*net.UDPConn, error)
	httpServerFactory func(string, int, http.Handler) *http.Server
}

// InjectableServerOptions holds configuration and dependency injection options
type InjectableServerOptions struct {
	Port      int
	HTTPPort  int
	RedisAddr string
	
	// Optional dependency injection
	Store           Store
	StoreFactory    func() Store
	HandlerFactory  func(Store) *HTTPHandler
	ListenerFactory func(string, int) (*net.UDPConn, error)
	HTTPServerFactory func(string, int, http.Handler) *http.Server
}

// NewInjectableServer creates a new server with dependency injection support
func NewInjectableServer(opts InjectableServerOptions) *InjectableServer {
	server := &InjectableServer{
		port:              opts.Port,
		httpPort:          opts.HTTPPort,
		storeFactory:      opts.StoreFactory,
		handlerFactory:    opts.HandlerFactory,
		listenerFactory:   opts.ListenerFactory,
		httpServerFactory: opts.HTTPServerFactory,
	}
	
	// Set default factories if not provided
	if server.storeFactory == nil {
		server.storeFactory = func() Store {
			if opts.RedisAddr == "" {
				return NewClaimStore()
			} else {
				store, err := NewClaimStoreWithRedis(opts.RedisAddr)
				if err != nil {
					log.Fatalf("Failed to connect to Redis at %s: %v", opts.RedisAddr, err)
				}
				return store
			}
		}
	}
	
	if server.handlerFactory == nil {
		server.handlerFactory = NewHTTPHandler
	}
	
	if server.listenerFactory == nil {
		server.listenerFactory = defaultListenerFactory
	}
	
	if server.httpServerFactory == nil {
		server.httpServerFactory = defaultHTTPServerFactory
	}
	
	// Use provided store or create one
	if opts.Store != nil {
		server.store = opts.Store
	} else {
		server.store = server.storeFactory()
	}
	
	// Create HTTP handler
	server.httpHandler = server.handlerFactory(server.store)
	
	return server
}

// Start starts the injectable server
func (s *InjectableServer) Start() error {
	// Start UDP server
	if err := s.startUDPServer(); err != nil {
		return fmt.Errorf("failed to start UDP server: %w", err)
	}
	
	// Start HTTP server
	if err := s.startHTTPServer(); err != nil {
		s.stopUDPServer()
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
	
	return nil
}

// startUDPServer starts the UDP server using the injected factory
func (s *InjectableServer) startUDPServer() error {
	listener, err := s.listenerFactory("::", s.port)
	if err != nil {
		return err
	}
	
	s.listener = listener
	log.Printf("Injectable SpaceNet UDP server listening on [::]:%d", s.port)
	
	// Start processing packets
	go s.processPackets()
	
	return nil
}

// startHTTPServer starts the HTTP server using the injected factory
func (s *InjectableServer) startHTTPServer() error {
	router := mux.NewRouter()
	s.httpHandler.RegisterRoutes(router)
	
	s.httpServer = s.httpServerFactory("", s.httpPort, router)
	
	// Start the HTTP server in a goroutine
	go func() {
		log.Printf("Injectable SpaceNet HTTP server listening on :%d", s.httpPort)
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()
	
	return nil
}

// Stop stops the injectable server
func (s *InjectableServer) Stop() {
	s.stopHTTPServer()
	s.stopUDPServer()
	
	if s.store != nil {
		s.store.Close()
	}
}

// stopUDPServer stops the UDP server
func (s *InjectableServer) stopUDPServer() {
	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}
}

// stopHTTPServer stops the HTTP server
func (s *InjectableServer) stopHTTPServer() {
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down HTTP server: %v", err)
		}
		
		s.httpServer = nil
	}
}

// processPackets handles incoming UDP packets (same as original)
func (s *InjectableServer) processPackets() {
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
				claimantName := string(job.data)
				s.store.ProcessClaim(clientIP, claimantName)
				log.Printf("Address %s claimed by %s", clientIP, claimantName)
			}
		}(i)
	}
	
	// Main loop to receive UDP packets
	for {
		buffer := make([]byte, 32)
		
		n, clientAddr, err := s.listener.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading UDP packet: %v", err)
			if _, ok := err.(net.Error); ok {
				log.Println("Listener closed, stopping packet processing")
				close(jobs)
				return
			}
			continue
		}
		
		if n == 0 {
			log.Printf("Empty claim from %s, ignoring", clientAddr.IP.String())
			continue
		}
		
		if n > 32 {
			log.Printf("Claim payload too large from %s (%d bytes), ignoring", clientAddr.IP.String(), n)
			continue
		}
		
		job := claimJob{
			addr: clientAddr,
			data: buffer[:n],
		}
		select {
		case jobs <- job:
			// Job queued successfully
		default:
			log.Printf("Worker queue is full, dropping packet from %s", clientAddr.IP.String())
		}
	}
}

// GetStore returns the server's store (useful for testing)
func (s *InjectableServer) GetStore() Store {
	return s.store
}

// GetHTTPHandler returns the server's HTTP handler (useful for testing)
func (s *InjectableServer) GetHTTPHandler() *HTTPHandler {
	return s.httpHandler
}

// Default factory implementations

func defaultListenerFactory(ip string, port int) (*net.UDPConn, error) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(ip),
	}
	return net.ListenUDP("udp", &addr)
}

func defaultHTTPServerFactory(host string, port int, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: handler,
	}
}