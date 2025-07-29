package server

import (
	"context"
	"time"
)

// RedisClient defines the interface for Redis operations used by ClaimStore
// This abstraction allows for easy mocking and testing
type RedisClient interface {
	// Ping tests connectivity to Redis server
	Ping(ctx context.Context) error

	// Get retrieves a value from Redis by key
	// Returns the value and true if key exists, empty string and false if not
	Get(ctx context.Context, key string) (string, bool, error)

	// Set stores a key-value pair in Redis with optional expiration
	Set(ctx context.Context, key string, value string, expiration time.Duration) error

	// Scan iterates over keys matching a pattern
	// Returns keys, next cursor, and any error
	Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error)

	// Close closes the Redis connection
	Close() error
}