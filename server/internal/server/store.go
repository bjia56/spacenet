package server

import "github.com/bjia56/spacenet/server/api"

// SubnetStats represents statistics about a subnet
type SubnetStats = api.SubnetResponse

// Store defines the interface for claim storage backends
type Store interface {
	// ProcessClaim processes a claim request and updates the store
	ProcessClaim(ipAddr string, claimant string) error

	// GetClaim retrieves the claimant for an IP address
	GetClaim(ipAddr string) (string, bool)

	// GetAllClaims returns all claims in the store
	GetAllClaims() map[string]string

	// GetSubnetStats retrieves statistics for a specific subnet
	GetSubnetStats(subnet string) (*SubnetStats, bool)

	// Close releases any resources held by the store
	Close() error
}
