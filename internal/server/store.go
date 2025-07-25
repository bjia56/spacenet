package server

import (
	"github.com/bjia56/spacenet/api"
)

// Store defines the interface for claim storage backends
type Store interface {
	// ProcessClaim processes a claim request and updates the store
	ProcessClaim(ipAddr string, claimant string) error

	// GetClaim retrieves the claimant for an IP address
	GetClaim(ipAddr string) (string, bool)

	// GetAllClaims returns all claims in the store
	GetAllClaims() map[string]string

	// GetSubnetStats retrieves statistics for a specific subnet
	GetSubnetStats(subnet string) (*api.SubnetStats, bool)

	// GetAllSubnets retrieves statistics for all tracked subnets with the given prefix length
	GetAllSubnets(prefixLen int) []api.SubnetStats

	// Close releases any resources held by the store
	Close() error
}
