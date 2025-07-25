package server

import (
	"context"
	"sync"

	"github.com/bjia56/spacenet/api"
	"github.com/redis/go-redis/v9"
)

// ClaimStore is an in-memory store for IP address claims
// It can optionally use Redis as a backend store
type ClaimStore struct {
	mutex       sync.RWMutex
	claims      map[string]string // map[ipAddress]claimantName
	ipTree      *IPTree           // Hierarchical tree for subnet-based queries
	redisClient *redis.Client     // Optional Redis client for persistence
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

	store := &ClaimStore{
		claims:      make(map[string]string),
		ipTree:      NewIPTree(),
		redisClient: client,
		ctx:         ctx,
	}

	// Load existing claims from Redis
	err = store.loadFromRedis()
	if err != nil {
		return nil, err
	}

	return store, nil
}

// loadFromRedis loads all claims from Redis into memory
func (cs *ClaimStore) loadFromRedis() error {
	// Use SCAN to iterate over all keys
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = cs.redisClient.Scan(cs.ctx, cursor, "*", 10).Result()
		if err != nil {
			return err
		}

		// Get values for all keys
		for _, key := range keys {
			val, err := cs.redisClient.Get(cs.ctx, key).Result()
			if err == nil && val != "" {
				// Store in memory
				cs.claims[key] = val
				// Update the tree
				cs.ipTree.processClaim(key, val, "")
			}
		}

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
		if err := cs.redisClient.Set(cs.ctx, ipAddr, claimant, 0).Err(); err != nil {
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
func (cs *ClaimStore) GetSubnetStats(subnet string) (*api.SubnetStats, bool) {
	return cs.ipTree.GetSubnetStats(subnet)
}

// GetAllSubnets retrieves statistics for all tracked subnets with the given prefix length
func (cs *ClaimStore) GetAllSubnets(prefixLen int) []api.SubnetStats {
	return cs.ipTree.GetAllSubnets(prefixLen)
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
