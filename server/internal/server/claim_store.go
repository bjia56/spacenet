package server

import (
	"context"
	"sync"
)

// ClaimStore is an in-memory store for IP address claims
// It can optionally use Redis as a backend store
type ClaimStore struct {
	mutex       sync.RWMutex
	claims      map[string]string // map[ipAddress]claimantName
	ipTree      *IPTree           // Hierarchical tree for subnet-based queries
	redisClient RedisClient       // Optional Redis client for persistence
	ctx         context.Context   // Context for Redis operations
}

// Verify ClaimStore implements Store interface
var _ Store = (*ClaimStore)(nil)

// NewClaimStore creates a new in-memory claim store without Redis
func NewClaimStore() *ClaimStore {
	return &ClaimStore{
		claims: make(map[string]string),
		ipTree: NewIPTree(),
	}
}

// NewClaimStoreWithRedis creates a claim store with Redis backend
func NewClaimStoreWithRedis(redisAddr string) (*ClaimStore, error) {
	client, err := NewRedisClient(redisAddr)
	if err != nil {
		return nil, err
	}

	return NewClaimStoreWithRedisClient(client), nil
}

// NewClaimStoreWithRedisClient creates a claim store with an injected Redis client
// This is useful for testing with mocks or custom Redis configurations
func NewClaimStoreWithRedisClient(redisClient RedisClient) *ClaimStore {
	ctx := context.Background()

	store := &ClaimStore{
		claims:      make(map[string]string),
		ipTree:      NewIPTree(),
		redisClient: redisClient,
		ctx:         ctx,
	}

	// Load existing claims from Redis (ignore errors for mocks in tests)
	store.loadFromRedis()

	return store
}

// loadFromRedis loads all claims from Redis into memory
func (cs *ClaimStore) loadFromRedis() error {
	// Use SCAN to iterate over all keys
	var cursor uint64
	for {
		keys, nextCursor, err := cs.redisClient.Scan(cs.ctx, cursor, "*", 10)
		if err != nil {
			return err
		}

		// Get values for all keys
		for _, key := range keys {
			val, exists, err := cs.redisClient.Get(cs.ctx, key)
			if err == nil && exists && val != "" {
				// Store in memory
				cs.claims[key] = val
				// Update the tree
				cs.ipTree.processClaim(key, val, "")
			}
		}

		// Update cursor for next iteration
		cursor = nextCursor
		
		// No more keys
		if cursor == 0 {
			break
		}
	}

	return nil
}

// ProcessClaim processes a claim request and updates the store
// Note: Updated to overwrite existing claims as per new requirements
func (cs *ClaimStore) ProcessClaim(ipAddr string, claimant string) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Get existing claimant if any
	oldClaimant, exists := cs.claims[ipAddr]

	// Store new claim in memory
	cs.claims[ipAddr] = claimant

	// If Redis is enabled, write through to Redis
	if cs.redisClient != nil {
		if err := cs.redisClient.Set(cs.ctx, ipAddr, claimant, 0); err != nil {
			// If Redis fails, revert the in-memory change and propagate error
			if exists {
				cs.claims[ipAddr] = oldClaimant
			} else {
				delete(cs.claims, ipAddr)
			}
			return err
		}
	}

	// Update tree with hierarchical information
	if exists {
		// We're updating an existing claim
		cs.ipTree.processClaim(ipAddr, claimant, oldClaimant)
	} else {
		// New claim
		cs.ipTree.processClaim(ipAddr, claimant, "")
	}

	return nil
}

// GetClaim retrieves the claimant for an IP address
func (cs *ClaimStore) GetClaim(ipAddr string) (string, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	claimant, exists := cs.claims[ipAddr]
	return claimant, exists
}

// GetSubnetStats retrieves statistics for a specific subnet
func (cs *ClaimStore) GetSubnetStats(subnet string) (*SubnetStats, bool) {
	return cs.ipTree.GetSubnetStats(subnet)
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

// Close releases any resources held by the store
func (cs *ClaimStore) Close() error {
	if cs.redisClient != nil {
		return cs.redisClient.Close()
	}
	return nil
}
