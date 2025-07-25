package server

import "sync"

// ClaimStore is an in-memory store for IP address claims
type ClaimStore struct {
	mutex  sync.RWMutex
	claims map[string]string // map[ipAddress]claimantName
}

// Verify ClaimStore implements Store interface
var _ Store = (*ClaimStore)(nil)

// NewClaimStore creates a new in-memory claim store
func NewClaimStore() *ClaimStore {
	return &ClaimStore{
		claims: make(map[string]string),
	}
}

// ProcessClaim processes a claim request and updates the store
// Note: Updated to overwrite existing claims as per new requirements
func (cs *ClaimStore) ProcessClaim(ipAddr string, claimant string) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	// Store new claim (overwriting existing one if any)
	cs.claims[ipAddr] = claimant
	return nil
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

// Close releases any resources held by the store
// For in-memory store, this is a no-op
func (cs *ClaimStore) Close() error {
	return nil
}
