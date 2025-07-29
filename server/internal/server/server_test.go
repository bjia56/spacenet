package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/bjia56/spacenet/server/api"
)

// TestServerBasicFunctionality tests basic server start/stop functionality
func TestServerBasicFunctionality(t *testing.T) {
	// Use ephemeral ports for testing
	server := NewServer(0) // Port 0 will assign an available port
	server.httpPort = 0    // Also use ephemeral port for HTTP

	// Test server creation
	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	// Test that server can start
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test that server can stop
	server.Stop()
}

// TestClaimStoreBasicOperations tests basic ClaimStore functionality
func TestClaimStoreBasicOperations(t *testing.T) {
	store := NewClaimStore()

	// Test processing a claim
	err := store.ProcessClaim("2001:db8::1", "testuser")
	if err != nil {
		t.Fatalf("ProcessClaim failed: %v", err)
	}

	// Test retrieving a claim
	claimant, exists := store.GetClaim("2001:db8::1")
	if !exists {
		t.Fatal("Expected claim to exist")
	}
	if claimant != "testuser" {
		t.Fatalf("Expected claimant 'testuser', got '%s'", claimant)
	}

	// Test non-existent claim
	_, exists = store.GetClaim("2001:db8::2")
	if exists {
		t.Fatal("Expected claim to not exist")
	}

	// Test getting all claims
	allClaims := store.GetAllClaims()
	if len(allClaims) != 1 {
		t.Fatalf("Expected 1 claim, got %d", len(allClaims))
	}

	// Test overwriting claim
	err = store.ProcessClaim("2001:db8::1", "newuser")
	if err != nil {
		t.Fatalf("ProcessClaim failed: %v", err)
	}

	claimant, exists = store.GetClaim("2001:db8::1")
	if !exists {
		t.Fatal("Expected claim to exist")
	}
	if claimant != "newuser" {
		t.Fatalf("Expected claimant 'newuser', got '%s'", claimant)
	}

	// Test store close
	err = store.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

// TestHTTPHandlerBasicOperations tests HTTP API endpoints
func TestHTTPHandlerBasicOperations(t *testing.T) {
	// Create a server with ephemeral ports
	server := NewServer(0)
	server.httpPort = 0

	// Start the server
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Get the actual HTTP port that was assigned
	httpPort := server.httpPort
	if httpPort == 0 {
		// If still 0, try to extract from the HTTP server
		t.Skip("Cannot determine HTTP port for testing")
	}

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)

	// Test health endpoint
	resp, err := http.Get(fmt.Sprintf("%s/health", baseURL))
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Add a claim directly to the store
	err = server.store.ProcessClaim("2001:db8::1", "testuser")
	if err != nil {
		t.Fatalf("Failed to add claim: %v", err)
	}

	// Test get claim endpoint
	resp, err = http.Get(fmt.Sprintf("%s/api/ip/2001:db8::1", baseURL))
	if err != nil {
		t.Fatalf("Get claim failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var claimResp api.ClaimResponse
	err = json.NewDecoder(resp.Body).Decode(&claimResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if claimResp.IP != "2001:db8::1" {
		t.Fatalf("Expected IP '2001:db8::1', got '%s'", claimResp.IP)
	}
	if claimResp.Claimant != "testuser" {
		t.Fatalf("Expected claimant 'testuser', got '%s'", claimResp.Claimant)
	}

	// Test non-existent claim
	resp, err = http.Get(fmt.Sprintf("%s/api/ip/2001:db8::999", baseURL))
	if err != nil {
		t.Fatalf("Get non-existent claim failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", resp.StatusCode)
	}
}

// TestUDPClaimProcessing tests UDP claim processing functionality
func TestUDPClaimProcessing(t *testing.T) {
	// Create server with ephemeral ports
	server := NewServer(0)
	server.httpPort = 0

	// Start the server
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Get the actual UDP port that was assigned
	udpPort := server.port
	if udpPort == 0 {
		t.Skip("Cannot determine UDP port for testing")
	}

	// Create UDP client
	serverAddr := &net.UDPAddr{
		IP:   net.ParseIP("::1"),
		Port: udpPort,
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.Fatalf("Failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	// Send a claim
	claimData := []byte("testuser")
	_, err = conn.Write(claimData)
	if err != nil {
		t.Fatalf("Failed to send UDP claim: %v", err)
	}

	// Give server time to process the claim
	time.Sleep(200 * time.Millisecond)

	// Verify the claim was processed by checking the store
	// Note: We need to get the client's IPv6 address that the server saw
	// For testing purposes, we'll check if any claims were added
	allClaims := server.store.GetAllClaims()
	if len(allClaims) == 0 {
		t.Fatal("Expected at least one claim to be processed")
	}

	// Verify that one of the claims has our claimant name
	found := false
	for _, claimant := range allClaims {
		if claimant == "testuser" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("Expected to find claim from 'testuser'")
	}
}

// TestSubnetStats tests basic subnet statistics functionality
func TestSubnetStats(t *testing.T) {
	store := NewClaimStore()

	// Add some claims in a subnet
	err := store.ProcessClaim("2001:db8::1", "user1")
	if err != nil {
		t.Fatalf("ProcessClaim failed: %v", err)
	}

	err = store.ProcessClaim("2001:db8::2", "user1")
	if err != nil {
		t.Fatalf("ProcessClaim failed: %v", err)
	}

	err = store.ProcessClaim("2001:db8::3", "user2")
	if err != nil {
		t.Fatalf("ProcessClaim failed: %v", err)
	}

	// Test getting subnet stats
	stats, ok := store.GetSubnetStats("2001:db8::/64")
	if !ok {
		t.Fatal("Expected subnet stats to be available")
	}

	// Should have stats (though specific values depend on the tree implementation)
	if stats == nil {
		t.Fatal("Expected non-nil subnet stats")
	}

	// The stats should be valid (no specific owner expected due to low claim density)
	t.Logf("Subnet stats: Owner=%s, Percentage=%.2f", stats.Owner, stats.Percentage)
}

// TestServerIntegration tests end-to-end functionality
func TestServerIntegration(t *testing.T) {
	// Create server
	server := NewServer(0)
	server.httpPort = 0

	// Start server
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Add claim via store (simulating UDP processing)
	testIP := "2001:db8::42"
	testUser := "integrationtest"
	err = server.store.ProcessClaim(testIP, testUser)
	if err != nil {
		t.Fatalf("Failed to process claim: %v", err)
	}

	// Verify via HTTP API
	if server.httpPort == 0 {
		t.Skip("Cannot determine HTTP port for integration test")
	}

	baseURL := fmt.Sprintf("http://localhost:%d", server.httpPort)
	resp, err := http.Get(fmt.Sprintf("%s/api/ip/%s", baseURL, testIP))
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var claimResp api.ClaimResponse
	err = json.NewDecoder(resp.Body).Decode(&claimResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if claimResp.IP != testIP {
		t.Fatalf("Expected IP '%s', got '%s'", testIP, claimResp.IP)
	}
	if claimResp.Claimant != testUser {
		t.Fatalf("Expected claimant '%s', got '%s'", testUser, claimResp.Claimant)
	}
}