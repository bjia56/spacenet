package server

import (
	"encoding/json"
	"log"
	"net"
	"net/http"

	"github.com/bjia56/spacenet/server/api"
	"github.com/gorilla/mux"
)

// HTTPHandler implements HTTP endpoints for claim management
type HTTPHandler struct {
	store Store
}

// NewHTTPHandler creates a new HTTP handler with the given store
func NewHTTPHandler(store Store) *HTTPHandler {
	return &HTTPHandler{
		store: store,
	}
}

// RegisterRoutes registers all HTTP routes on the provided router
func (h *HTTPHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/ip/{ip}", h.handleGetClaimByIP).Methods("GET")
	router.HandleFunc("/api/subnet/{address}/{prefix}", h.handleGetStatsBySubnet).Methods("GET")
	router.HandleFunc("/health", h.handleHealth).Methods("GET")
}

// handleHealth handles the health check endpoint
func (h *HTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleGetClaimByIP returns the claim for a specific IP
func (h *HTTPHandler) handleGetClaimByIP(w http.ResponseWriter, r *http.Request) {
	// Extract IP from URL variables
	vars := mux.Vars(r)
	ipAddr, ok := vars["ip"]
	if !ok || ipAddr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "IP address is required"})
		return
	}

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
	difficulty := h.store.CalculateDifficulty(ipAddr)

	w.Header().Set("Content-Type", "application/json")
	response := api.ClaimResponse{
		Claimant:   claimant,
		Difficulty: difficulty,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// handleGetStatsSubnet returns statistics for a specified IPv6 subnet
func (h *HTTPHandler) handleGetStatsBySubnet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var subnetStr string

	address := vars["address"]
	prefix := vars["prefix"]
	subnetStr = address + "/" + prefix

	// Get subnet statistics
	stats, ok := h.store.GetSubnetStats(subnetStr)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid subnet format, use CIDR notation (e.g., 2001:db8::/64)"})
		return
	}

	// Convert to response format
	var response *api.SubnetResponse = stats

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
