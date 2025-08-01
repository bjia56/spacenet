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
	router.HandleFunc("/api/claim/{ip}", h.handleSubmitClaim).Methods("POST")
	router.HandleFunc("/health", h.handleHealth).Methods("GET")
}

// handleHealth handles the health check endpoint
func (h *HTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// handleGetClaimByIP returns the claim for a specific IP
func (h *HTTPHandler) handleGetClaimByIP(w http.ResponseWriter, r *http.Request) {
	// Extract IP from URL variables
	vars := mux.Vars(r)
	ipAddr, ok := vars["ip"]
	if !ok || ipAddr == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate the IP address
	if net.ParseIP(ipAddr) == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claimant, exists := h.store.GetClaim(ipAddr)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	difficulty := h.store.CalculateDifficulty(ipAddr)

	w.Header().Set("Content-Type", "application/json")
	response := api.ClaimResponse{
		Name:       claimant,
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
		return
	}

	// Convert to response format
	response := stats

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// handleSubmitClaim handles claim submission via HTTP POST
func (h *HTTPHandler) handleSubmitClaim(w http.ResponseWriter, r *http.Request) {
	// Extract IP from URL path
	vars := mux.Vars(r)
	ipAddr, ok := vars["ip"]
	if !ok || ipAddr == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate IP address
	targetIP := net.ParseIP(ipAddr)
	if targetIP == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse JSON request body
	var claimReq api.ClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&claimReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate claimant name
	if len(claimReq.Name) == 0 || len(claimReq.Name) > 24 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create proof of work object
	pow := &api.ProofOfWork{
		Target: targetIP,
		Name:   claimReq.Name,
		Nonce:  claimReq.Nonce,
	}

	// Validate proof of work
	if err := h.store.ValidateProofOfWork(pow); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Process the claim
	err := h.store.ProcessClaim(ipAddr, claimReq.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return success with no content
	w.WriteHeader(http.StatusCreated)
}
