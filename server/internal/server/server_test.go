package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/bjia56/spacenet/server/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to make HTTP claim request with proof of work
func makeHTTPClaimRequest(t *testing.T, baseURL, targetIP, claimant string, difficulty uint8) *http.Response {
	// Solve proof of work
	pow, err := api.SolveProofOfWork(net.ParseIP(targetIP), claimant, difficulty, 1000000)
	require.NoError(t, err, "Should be able to solve proof of work")

	// Create claim request
	claimReq := api.ClaimRequest{
		Nonce: pow.Nonce,
		Name:  claimant,
	}

	// Serialize request
	reqBody, err := json.Marshal(claimReq)
	require.NoError(t, err, "Should be able to marshal claim request")

	// Make HTTP request
	resp, err := http.Post(fmt.Sprintf("%s/api/claim/%s", baseURL, targetIP), "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err, "HTTP claim request should succeed")

	return resp
}

// TestServerBasicFunctionality tests basic server start/stop functionality
func TestServerBasicFunctionality(t *testing.T) {
	// Use ephemeral ports for testing
	server := NewServerWithOptions(ServerOptions{
		HTTPPort: 0, // Use ephemeral port for HTTP
	})

	// Test server creation
	require.NotNil(t, server, "NewServerWithOptions() should not return nil")

	// Test that server can start
	err := server.Start()
	require.NoError(t, err, "Server should start without error")

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
	require.NoError(t, err, "ProcessClaim should not fail")

	// Test retrieving a claim
	claimant, exists := store.GetClaim("2001:db8::1")
	assert.True(t, exists, "Claim should exist")
	assert.Equal(t, "testuser", claimant, "Claimant should match")

	// Test non-existent claim
	_, exists = store.GetClaim("2001:db8::2")
	assert.False(t, exists, "Non-existent claim should not exist")

	// Test getting all claims
	allClaims := store.GetAllClaims()
	assert.Len(t, allClaims, 1, "Should have exactly 1 claim")

	// Test overwriting claim
	err = store.ProcessClaim("2001:db8::1", "newuser")
	require.NoError(t, err, "ProcessClaim should not fail on overwrite")

	claimant, exists = store.GetClaim("2001:db8::1")
	assert.True(t, exists, "Overwritten claim should exist")
	assert.Equal(t, "newuser", claimant, "Claimant should be updated")

	// Test store close
	err = store.Close()
	require.NoError(t, err, "Store close should not fail")
}

// TestHTTPHandlerBasicOperations tests HTTP API endpoints
func TestHTTPHandlerBasicOperations(t *testing.T) {
	// Create a server with ephemeral ports
	server := NewServerWithOptions(ServerOptions{
		HTTPPort: 0,
	})

	// Start the server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	// Wait for HTTP port to be assigned
	httpPort, err := server.WaitForHTTPPort(5 * time.Second)
	require.NoError(t, err, "HTTP port should be assigned within timeout")

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)

	// Test health endpoint
	resp, err := http.Get(fmt.Sprintf("%s/health", baseURL))
	require.NoError(t, err, "Health check request should succeed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health check should return 200")

	// Add a claim directly to the store
	err = server.store.ProcessClaim("2001:db8::1", "testuser")
	require.NoError(t, err, "Adding claim should succeed")

	// Test get claim endpoint
	resp, err = http.Get(fmt.Sprintf("%s/api/ip/2001:db8::1", baseURL))
	require.NoError(t, err, "Get claim request should succeed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get claim should return 200")

	var claimResp api.ClaimResponse
	err = json.NewDecoder(resp.Body).Decode(&claimResp)
	require.NoError(t, err, "Response should decode successfully")

	assert.Equal(t, "testuser", claimResp.Name, "Response claimant should match")

	// Test non-existent claim
	resp, err = http.Get(fmt.Sprintf("%s/api/ip/2001:db8::999", baseURL))
	require.NoError(t, err, "Non-existent claim request should succeed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Non-existent claim should return 404")
}

// TestHTTPClaimProcessing tests HTTP claim processing functionality
func TestHTTPClaimProcessing(t *testing.T) {
	// Create server with ephemeral ports
	server := NewServerWithOptions(ServerOptions{
		HTTPPort: 0,
	})

	// Start the server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	// Wait for HTTP port to be assigned
	httpPort, err := server.WaitForHTTPPort(5 * time.Second)
	require.NoError(t, err, "HTTP port should be assigned within timeout")

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)

	// Make HTTP claim request
	targetIP := "2001:db8::1"
	resp := makeHTTPClaimRequest(t, baseURL, targetIP, "testuser", 8)
	defer resp.Body.Close()

	// Verify claim was accepted
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "HTTP claim should return 201 Created")

	// Verify the claim was processed by checking the store
	claimant, exists := server.store.GetClaim(targetIP)
	assert.True(t, exists, "Claim should exist in store")
	assert.Equal(t, "testuser", claimant, "Claimant should match")
}

// TestSubnetStats tests basic subnet statistics functionality
func TestSubnetStats(t *testing.T) {
	store := NewClaimStore()

	// Add some claims in a subnet
	err := store.ProcessClaim("2001:db8::1", "user1")
	require.NoError(t, err, "ProcessClaim should succeed")

	err = store.ProcessClaim("2001:db8::2", "user1")
	require.NoError(t, err, "ProcessClaim should succeed")

	err = store.ProcessClaim("2001:db8::3", "user2")
	require.NoError(t, err, "ProcessClaim should succeed")

	// Test getting subnet stats
	stats, ok := store.GetSubnetStats("2001:db8::/64")
	assert.True(t, ok, "Subnet stats should be available")

	// Should have stats (though specific values depend on the tree implementation)
	assert.NotNil(t, stats, "Subnet stats should not be nil")

	// The stats should be valid (no specific owner expected due to low claim density)
	t.Logf("Subnet stats: Owner=%s, Percentage=%.2f", stats.Owner, stats.Percentage)
}

// TestServerIntegration tests end-to-end functionality
func TestServerIntegration(t *testing.T) {
	// Create server
	server := NewServerWithOptions(ServerOptions{
		HTTPPort: 0,
	})

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	// Wait for HTTP port to be assigned
	httpPort, err := server.WaitForHTTPPort(5 * time.Second)
	require.NoError(t, err, "HTTP port should be assigned within timeout")

	// Add claim via store (simulating UDP processing)
	testIP := "2001:db8::42"
	testUser := "integrationtest"
	err = server.store.ProcessClaim(testIP, testUser)
	require.NoError(t, err, "Claim processing should succeed")

	// Verify via HTTP API
	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)
	resp, err := http.Get(fmt.Sprintf("%s/api/ip/%s", baseURL, testIP))
	require.NoError(t, err, "HTTP request should succeed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "API request should return 200")

	var claimResp api.ClaimResponse
	err = json.NewDecoder(resp.Body).Decode(&claimResp)
	require.NoError(t, err, "Response should decode successfully")

	assert.Equal(t, testUser, claimResp.Name, "Response claimant should match")
}

// TestHTTPHandler_InvalidIPAddress tests error handling for invalid IP addresses
func TestHTTPHandler_InvalidIPAddress(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		HTTPPort: 0,
	})

	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	httpPort, err := server.WaitForHTTPPort(5 * time.Second)
	require.NoError(t, err, "HTTP port should be assigned within timeout")

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)

	// Test invalid IP address format
	resp, err := http.Get(fmt.Sprintf("%s/api/ip/invalid-ip", baseURL))
	require.NoError(t, err, "Request should succeed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Invalid IP should return 400")
}

// TestHTTPHandler_SubnetStats tests subnet statistics endpoint
func TestHTTPHandler_SubnetStats(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		HTTPPort: 0,
	})

	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	httpPort, err := server.WaitForHTTPPort(5 * time.Second)
	require.NoError(t, err, "HTTP port should be assigned within timeout")

	// Add some claims to create subnet statistics
	err = server.store.ProcessClaim("2001:db8::1", "user1")
	require.NoError(t, err, "Adding claim should succeed")
	err = server.store.ProcessClaim("2001:db8::2", "user1")
	require.NoError(t, err, "Adding claim should succeed")

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)

	// Test subnet stats endpoint
	resp, err := http.Get(fmt.Sprintf("%s/api/subnet/2001:db8::/64", baseURL))
	require.NoError(t, err, "Subnet stats request should succeed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Subnet stats should return 200")

	var statsResp api.SubnetResponse
	err = json.NewDecoder(resp.Body).Decode(&statsResp)
	require.NoError(t, err, "Stats response should decode successfully")

	// Stats should contain valid data
	t.Logf("Subnet stats: Owner=%s, Percentage=%.2f", statsResp.Owner, statsResp.Percentage)
}

// TestHTTPHandler_InvalidSubnet tests error handling for invalid subnet format
func TestHTTPHandler_InvalidSubnet(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		HTTPPort: 0,
	})

	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	httpPort, err := server.WaitForHTTPPort(5 * time.Second)
	require.NoError(t, err, "HTTP port should be assigned within timeout")

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)

	// Test invalid subnet format
	resp, err := http.Get(fmt.Sprintf("%s/api/subnet/invalid-subnet/prefix", baseURL))
	require.NoError(t, err, "Request should succeed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Invalid subnet should return 400")
}

// TestHTTPServer_InvalidPayload tests handling of invalid HTTP payloads
func TestHTTPServer_InvalidPayload(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		HTTPPort: 0,
	})

	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	httpPort, err := server.WaitForHTTPPort(5 * time.Second)
	require.NoError(t, err, "HTTP port should be assigned within timeout")

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)

	// Test empty payload
	resp, err := http.Post(fmt.Sprintf("%s/api/claim/2001:db8::1", baseURL), "application/json", bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "Empty HTTP payload should be sent")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Empty payload should return 400")

	// Test invalid JSON
	resp, err = http.Post(fmt.Sprintf("%s/api/claim/2001:db8::1", baseURL), "application/json", bytes.NewBuffer([]byte("invalid json")))
	require.NoError(t, err, "Invalid JSON should be sent")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Invalid JSON should return 400")

	// Verify no claims were added (invalid payloads should be rejected)
	allClaims := server.store.GetAllClaims()
	assert.Empty(t, allClaims, "Invalid payloads should not create claims")
}

// TestHTTPServer_PayloadValidation tests HTTP payload validation
func TestHTTPServer_PayloadValidation(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		HTTPPort: 0,
	})

	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	httpPort, err := server.WaitForHTTPPort(5 * time.Second)
	require.NoError(t, err, "HTTP port should be assigned within timeout")

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)

	// Test oversized claimant name (> 24 characters)
	longClaimant := make([]byte, 30)
	for i := range longClaimant {
		longClaimant[i] = 'A'
	}

	claimReq := api.ClaimRequest{
		Nonce: "12345",
		Name:  string(longClaimant),
	}

	reqBody, err := json.Marshal(claimReq)
	require.NoError(t, err, "Should be able to marshal request")

	resp, err := http.Post(fmt.Sprintf("%s/api/claim/2001:db8::1", baseURL), "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err, "Request should be sent")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Oversized claimant should return 400")

	// Test invalid IP address in path
	claimReq = api.ClaimRequest{
		Nonce: "12345",
		Name:  "testuser",
	}

	reqBody, err = json.Marshal(claimReq)
	require.NoError(t, err, "Should be able to marshal request")

	resp, err = http.Post(fmt.Sprintf("%s/api/claim/invalid-ip", baseURL), "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err, "Request should be sent")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Invalid IP should return 400")

	// Verify no claims were added
	allClaims := server.store.GetAllClaims()
	assert.Empty(t, allClaims, "Invalid payloads should not create claims")
}

// TestServerPortTimeout tests timeout behavior for port assignment
func TestServerPortTimeout(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		HTTPPort: 0,
	})

	// Test timeout without starting server
	_, err := server.WaitForHTTPPort(100 * time.Millisecond)
	assert.Error(t, err, "Should timeout when server not started")
	assert.Contains(t, err.Error(), "timeout waiting for HTTP port")
}

// TestServerStop_Graceful tests graceful server shutdown
func TestServerStop_Graceful(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		HTTPPort: 0,
	})

	err := server.Start()
	require.NoError(t, err, "Server should start successfully")

	// Verify HTTP port is assigned
	httpPort, err := server.WaitForHTTPPort(2 * time.Second)
	require.NoError(t, err, "HTTP port should be assigned")
	assert.Greater(t, httpPort, 0, "HTTP port should be positive")

	// Stop should not hang or error
	server.Stop()

	// Give server time to fully shut down
	time.Sleep(200 * time.Millisecond)

	// HTTP server should be stopped
	client := &http.Client{Timeout: 100 * time.Millisecond}
	_, err = client.Get(fmt.Sprintf("http://localhost:%d/health", httpPort))
	assert.Error(t, err, "HTTP server should be stopped")

	// The main verification is that Stop() completed gracefully without hanging
}

// TestHTTPServer_DuplicateClaimHandling tests HTTP endpoint behavior with duplicate claims
func TestHTTPServer_DuplicateClaimHandling(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		HTTPPort: 0,
	})

	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	httpPort, err := server.WaitForHTTPPort(5 * time.Second)
	require.NoError(t, err, "HTTP port should be assigned within timeout")

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)
	targetIP := "2001:db8::1"
	testUser := "testuser"

	// Send initial claim
	resp := makeHTTPClaimRequest(t, baseURL, targetIP, testUser, 8)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Initial claim should be accepted")

	// Verify claim exists
	claimant, exists := server.store.GetClaim(targetIP)
	assert.True(t, exists, "Initial claim should exist")
	assert.Equal(t, testUser, claimant, "Initial claimant should match")

	// Get initial subnet stats
	stats, ok := server.store.GetSubnetStats("2001:db8::1/128")
	require.True(t, ok, "Should get initial subnet stats")
	initialPercentage := stats.Percentage

	// Send duplicate claim (same user, same IP) - need higher difficulty due to claim bonus
	resp = makeHTTPClaimRequest(t, baseURL, targetIP, testUser, 12)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Duplicate claim should still be accepted by HTTP")

	// Verify claim still exists and hasn't changed
	claimant, exists = server.store.GetClaim(targetIP)
	assert.True(t, exists, "Claim should still exist after duplicate")
	assert.Equal(t, testUser, claimant, "Claimant should still be the same")

	// Most importantly: verify stats haven't inflated
	stats, ok = server.store.GetSubnetStats("2001:db8::1/128")
	require.True(t, ok, "Should still get subnet stats")
	assert.Equal(t, initialPercentage, stats.Percentage, "Percentage should not change after duplicate claim")
	assert.LessOrEqual(t, stats.Percentage, 100.0, "Percentage should never exceed 100%")

	// Verify we still have only one claim total
	allClaims := server.store.GetAllClaims()
	assert.Len(t, allClaims, 1, "Should still have exactly one claim")
}

// TestHTTPServer_MultipleDuplicateClaimsPercentage tests that multiple duplicate claims
// via HTTP don't cause percentage inflation
func TestHTTPServer_MultipleDuplicateClaimsPercentage(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		HTTPPort: 0,
	})

	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	httpPort, err := server.WaitForHTTPPort(5 * time.Second)
	require.NoError(t, err, "HTTP port should be assigned within timeout")

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)
	targetIP := "2001:db8::1"
	testUser := "testuser"

	// Send initial claim
	resp := makeHTTPClaimRequest(t, baseURL, targetIP, testUser, 8)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Initial claim should be accepted")

	// Get initial stats for comparison
	stats, ok := server.store.GetSubnetStats("2001:db8::1/128")
	require.True(t, ok, "Should get initial subnet stats")
	initialPercentage := stats.Percentage

	// Send multiple duplicate claims - need higher difficulty due to claim bonus
	for i := 0; i < 5; i++ {
		resp := makeHTTPClaimRequest(t, baseURL, targetIP, testUser, 12)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Duplicate claim %d should be accepted", i+1)
	}

	// Verify stats remain unchanged
	stats, ok = server.store.GetSubnetStats("2001:db8::1/128")
	require.True(t, ok, "Should still get subnet stats after multiple duplicates")
	assert.Equal(t, initialPercentage, stats.Percentage, "Percentage should remain unchanged after multiple duplicates")
	assert.LessOrEqual(t, stats.Percentage, 100.0, "Percentage should never exceed 100%")

	// Verify we still have only one claim
	allClaims := server.store.GetAllClaims()
	assert.Len(t, allClaims, 1, "Should still have exactly one claim after multiple duplicates")
}

// TestHTTPServer_SubnetStatsWithDuplicates tests that subnet stats API returns
// correct percentages even when duplicate claims were processed
func TestHTTPServer_SubnetStatsWithDuplicates(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		HTTPPort: 0,
	})

	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	httpPort, err := server.WaitForHTTPPort(5 * time.Second)
	require.NoError(t, err, "HTTP port should be assigned within timeout")

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)
	testUser := "testuser"

	// Make claims in a subnet and duplicates
	ips := []string{
		"2001:db8::1",
		"2001:db8::2",
		"2001:db8::3",
	}

	difficulty := uint8(8)
	for i, ip := range ips {
		// Initial claim
		resp := makeHTTPClaimRequest(t, baseURL, ip, testUser, difficulty+uint8(i)*4)
		resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Claim for %s should be accepted", ip)

		// Duplicate claim
		resp = makeHTTPClaimRequest(t, baseURL, ip, testUser, difficulty+uint8(i)*4+4)
		resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Duplicate claim for %s should be accepted", ip)
	}

	// Test subnet stats via HTTP API
	resp, err := http.Get(fmt.Sprintf("%s/api/subnet/2001:db8::/112", baseURL))
	require.NoError(t, err, "Subnet stats request should succeed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Subnet stats should return 200")

	var statsResp api.SubnetResponse
	err = json.NewDecoder(resp.Body).Decode(&statsResp)
	require.NoError(t, err, "Stats response should decode successfully")

	// Verify stats are reasonable (never exceed 100%)
	assert.LessOrEqual(t, statsResp.Percentage, 100.0, "HTTP API should return percentage <= 100%")
	assert.GreaterOrEqual(t, statsResp.Percentage, 0.0, "HTTP API should return non-negative percentage")

	t.Logf("HTTP API Subnet stats: Owner=%s, Percentage=%.6f%%", statsResp.Owner, statsResp.Percentage)

	// Verify we have exactly the expected number of claims (no duplicates counted)
	allClaims := server.store.GetAllClaims()
	assert.Len(t, allClaims, len(ips), "Should have exactly %d unique claims", len(ips))
}

// TestClaimStore_ConcurrentAccess tests concurrent access to claim store
func TestClaimStore_ConcurrentAccess(t *testing.T) {
	store := NewClaimStore()
	defer store.Close()

	const numGoroutines = 10
	const claimsPerGoroutine = 10

	// Launch multiple goroutines making concurrent claims
	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()
			for j := 0; j < claimsPerGoroutine; j++ {
				ip := fmt.Sprintf("2001:db8::%d:%d", goroutineID, j)
				user := fmt.Sprintf("user%d", goroutineID)
				err := store.ProcessClaim(ip, user)
				assert.NoError(t, err, "Concurrent claim should succeed")
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for range numGoroutines {
		<-done
	}

	// Verify all claims were processed
	allClaims := store.GetAllClaims()
	expectedCount := numGoroutines * claimsPerGoroutine
	assert.Len(t, allClaims, expectedCount, "All claims should be processed")
}

// TestNewServerWithOptions_DefaultValues tests default server creation
func TestNewServerWithOptions_DefaultValues(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{})
	require.NotNil(t, server, "Server should be created with default options")

	// Verify server has required components
	assert.NotNil(t, server.store, "Server should have a store")
	assert.NotNil(t, server.httpHandler, "Server should have HTTP handler")
	// UDP functionality removed - no UDP port channel expected
	assert.NotNil(t, server.httpPortReady, "Server should have HTTP port channel")
}

// TestHTTPServer_ClaimOverwrite tests that HTTP server properly handles claim overwrites
func TestHTTPServer_ClaimOverwrite(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		HTTPPort: 0,
	})

	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	httpPort, err := server.WaitForHTTPPort(5 * time.Second)
	require.NoError(t, err, "HTTP port should be assigned within timeout")

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)
	targetIP := "2001:db8::1"

	// Send first claim with base difficulty (8)
	resp := makeHTTPClaimRequest(t, baseURL, targetIP, "firstuser", 8)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "First claim should be accepted")

	// Verify first claim was processed
	claimant, exists := server.store.GetClaim(targetIP)
	assert.True(t, exists, "First claim should exist")
	assert.Equal(t, "firstuser", claimant, "First claimant should match")

	// Send second claim to overwrite the first
	// Need higher difficulty (12) since address is already claimed (8 base + 4 claim bonus)
	resp = makeHTTPClaimRequest(t, baseURL, targetIP, "seconduser", 12)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Second claim should be accepted")

	// Verify claim was overwritten
	claimant, exists = server.store.GetClaim(targetIP)
	assert.True(t, exists, "Overwritten claim should exist")
	assert.Equal(t, "seconduser", claimant, "Claimant should be updated to second user")

	// Verify only one claim exists for this IP
	allClaims := server.store.GetAllClaims()
	assert.Len(t, allClaims, 1, "Should have exactly one claim")
}
