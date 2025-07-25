package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bjia56/spacenet/internal/server"
	"github.com/spf13/cobra"
)

var (
	port      int
	redisAddr string
	useRedis  bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "spacenet",
		Short: "An IPv6 territory control game",
		Long:  "A space-themed network control game where players claim IPv6 addresses by sending UDP packets.",
		Run: func(cmd *cobra.Command, args []string) {
			runServer()
		},
	}

	// Define flags
	rootCmd.Flags().IntVarP(&port, "port", "p", 1337, "UDP port to listen on")
	rootCmd.Flags().StringVarP(&redisAddr, "redis", "r", "", "Redis address (host:port), if not specified in-memory store is used")

	// Parse flags and check if Redis address was explicitly set
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		useRedis = redisAddr != ""
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Failed to execute command: %v", err)
	}
}

// runServer starts the SpaceNet server with the configured options
func runServer() {
	log.Printf("Starting SpaceNet server on port %d", port)
	if !useRedis {
		log.Println("Using in-memory store")
	} else {
		log.Printf("Using Redis store at %s", redisAddr)
	}

	// Create a new server with options
	srv := server.NewServerWithOptions(server.ServerOptions{
		Port:        port,
		RedisAddr:   redisAddr,
		UseInMemory: !useRedis,
	})

	// Start the server
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-sigCh

	log.Println("Shutting down server...")
	srv.Stop()
	log.Println("Server stopped")
}
