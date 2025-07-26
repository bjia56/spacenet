// Package api defines shared data structures between the SpaceNet server and client
package api

// ClaimResponse represents the JSON response for a claim
type ClaimResponse struct {
	IP       string `json:"ip"`
	Claimant string `json:"claimant"`
}

// SubnetResponse represents the JSON response for subnet statistics
type SubnetResponse struct {
	Owner      string  `json:"owner,omitempty"`
	Percentage float64 `json:"percentage,omitempty"`
}
