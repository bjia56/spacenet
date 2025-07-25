package server

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// RedisStore implements a Redis-backed claim store
type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}

// Verify RedisStore implements Store interface
var _ Store = (*RedisStore)(nil)

// NewRedisStore creates a new Redis-backed store
func NewRedisStore(redisAddr string) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password
		DB:       0,  // use default DB
	})

	ctx := context.Background()

	// Test the connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &RedisStore{
		client: client,
		ctx:    ctx,
	}, nil
}

// ProcessClaim processes a claim request and updates the store
// Note: In the new implementation, claims overwrite existing ones
func (rs *RedisStore) ProcessClaim(ipAddr string, claimant string) error {
	// Set the new claim, overwriting any existing one
	return rs.client.Set(rs.ctx, ipAddr, claimant, 0).Err()
}

// GetClaim retrieves the claimant for an IP address
func (rs *RedisStore) GetClaim(ipAddr string) (string, bool) {
	claimant, err := rs.client.Get(rs.ctx, ipAddr).Result()
	if err == redis.Nil {
		// Key does not exist
		return "", false
	}
	if err != nil {
		// Other error
		return "", false
	}

	return claimant, true
}

// GetAllClaims returns all claims in the store
func (rs *RedisStore) GetAllClaims() map[string]string {
	claims := make(map[string]string)

	// Use SCAN to iterate over all keys
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = rs.client.Scan(rs.ctx, cursor, "*", 10).Result()
		if err != nil {
			break
		}

		// Get values for all keys
		for _, key := range keys {
			val, err := rs.client.Get(rs.ctx, key).Result()
			if err == nil {
				claims[key] = val
			}
		}

		// No more keys
		if cursor == 0 {
			break
		}
	}

	return claims
}

// Close closes the Redis connection
func (rs *RedisStore) Close() error {
	return rs.client.Close()
}
