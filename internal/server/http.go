package server

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
)

// HTTPHandler implements HTTP endpoints for claim management
type HTTPHandler struct {
	store Store
}

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

// NewHTTPHandler creates a new HTTP handler with the given store
func NewHTTPHandler(store Store) *HTTPHandler {
	return &HTTPHandler{
		store: store,
	}
}

// RegisterRoutes registers all HTTP routes on the provided mux
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/ip/", h.handleGetClaimByIP)
	mux.HandleFunc("/api/subnet/", h.handleGetStatsBySubnet)
	mux.HandleFunc("/health", h.handleHealth)
}

// handleHealth handles the health check endpoint
func (h *HTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleGetClaimByIP returns the claim for a specific IP
func (h *HTTPHandler) handleGetClaimByIP(w http.ResponseWriter, r *http.Request) {
	// Extract IP from path: /api/claims/[ip]
	path := strings.TrimPrefix(r.URL.Path, "/api/claims/")
	if path == "" || path == r.URL.Path {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "IP address is required"})
		return
	}

	ipAddr := path

	// Validate the IP address
	if net.ParseIP(ipAddr) == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid IP address format"})
		return
	}

	claimant, exists := h.store.GetClaim(ipAddr)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "No claim found for this IP"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := ClaimResponse{
		IP:       ipAddr,
		Claimant: claimant,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// handleGetStatsSubnet returns statistics for a specified IPv6 subnet
func (h *HTTPHandler) handleGetStatsBySubnet(w http.ResponseWriter, r *http.Request) {
	// Extract subnet from path: /api/subnet/[subnet]
	path := strings.TrimPrefix(r.URL.Path, "/api/subnet/")
	if path == "" || path == r.URL.Path {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Subnet is required"})
		return
	}

	subnetStr := path

	// Get subnet statistics
	stats, ok := h.store.GetSubnetStats(subnetStr)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid subnet format, use CIDR notation (e.g., 2001:db8::/64)"})
		return
	}

	// Convert to response format
	response := SubnetResponse{
		Subnet:             stats.Subnet,
		TotalAddresses:     stats.TotalAddresses,
		ClaimedAddresses:   stats.ClaimedAddresses,
		DominantClaimant:   stats.DominantClaimant,
		DominantPercentage: stats.DominantPercentage,
		AllClaimants:       stats.AllClaimants,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
