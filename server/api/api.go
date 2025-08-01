// Package api defines shared data structures between the SpaceNet server and client
package api

// ClaimResponse represents the JSON response for a claim
type ClaimResponse struct {
	Name       string `json:"name,omitempty"`
	Difficulty uint8  `json:"difficulty,omitempty"`
}

// SubnetResponse represents the JSON response for subnet statistics
type SubnetResponse struct {
	Owner      string  `json:"owner,omitempty"`
	Percentage float64 `json:"percentage,omitempty"`
}

// ClaimRequest represents a request to claim an IPv6 address
type ClaimRequest struct {
	Nonce string `json:"nonce"`
	Name  string `json:"name"`
}
