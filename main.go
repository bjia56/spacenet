package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bjia56/spacenet/internal/server"
)

func main() {
	// Create a new server on port 1337
	srv := server.NewServer(1337)

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
