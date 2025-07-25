package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/bjia56/spacenet/internal/server"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 1337, "UDP port to listen on")
	redisAddr := flag.String("redis", "localhost:6379", "Redis address (host:port)")
	useInMemory := flag.Bool("memory", false, "Use in-memory store instead of Redis")
	flag.Parse()

	// Allow overriding settings via environment variables
	if envPort := os.Getenv("SPACENET_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			*port = p
		}
	}
	if envRedis := os.Getenv("SPACENET_REDIS_ADDR"); envRedis != "" {
		*redisAddr = envRedis
	}
	if envMemory := os.Getenv("SPACENET_USE_MEMORY"); envMemory != "" {
		*useInMemory = envMemory == "1" || envMemory == "true"
	}

	log.Printf("Starting SpaceNet server on port %d", *port)
	if *useInMemory {
		log.Println("Using in-memory store")
	} else {
		log.Printf("Using Redis store at %s", *redisAddr)
	}

	// Create a new server with options
	srv := server.NewServerWithOptions(server.ServerOptions{
		Port:        *port,
		RedisAddr:   *redisAddr,
		UseInMemory: *useInMemory,
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
