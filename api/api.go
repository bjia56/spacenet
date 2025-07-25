// Package api defines shared data structures between the SpaceNet server and client
package api

// ClaimResponse represents the JSON response for a claim
type ClaimResponse struct {
	IP       string `json:"ip"`
	Claimant string `json:"claimant"`
}

// SubnetResponse represents the JSON response for subnet statistics
type SubnetResponse struct {
	Subnet             string             `json:"subnet"`
	TotalAddresses     string             `json:"totalAddresses"`
	ClaimedAddresses   string             `json:"claimedAddresses"`
	DominantClaimant   string             `json:"dominantClaimant"`
	DominantPercentage float64            `json:"dominantPercentage"`
	AllClaimants       map[string]float64 `json:"allClaimants"`
}

// SubnetStats represents statistics about a subnet
type SubnetStats struct {
	Subnet             string             `json:"subnet"`
	TotalAddresses     string             `json:"totalAddresses"`
	ClaimedAddresses   string             `json:"claimedAddresses"`
	DominantClaimant   string             `json:"dominantClaimant"`
	DominantPercentage float64            `json:"dominantPercentage"`
	AllClaimants       map[string]float64 `json:"allClaimants"` // Map of claimant to percentage
}
