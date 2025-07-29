package server

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClientImpl is a concrete implementation of RedisClient that wraps redis.Client
type RedisClientImpl struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client with the given address
func NewRedisClient(addr string) (*RedisClientImpl, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password
		DB:       0,  // use default DB
	})

	// Test the connection
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &RedisClientImpl{client: client}, nil
}

// Ping tests connectivity to Redis server
func (r *RedisClientImpl) Ping(ctx context.Context) error {
	_, err := r.client.Ping(ctx).Result()
	return err
}

// Get retrieves a value from Redis by key
func (r *RedisClientImpl) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

// Set stores a key-value pair in Redis with optional expiration
func (r *RedisClientImpl) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Scan iterates over keys matching a pattern
func (r *RedisClientImpl) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return r.client.Scan(ctx, cursor, match, count).Result()
}

// Close closes the Redis connection
func (r *RedisClientImpl) Close() error {
	return r.client.Close()
}