package server

import (
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

// TestServerBasicFunctionality tests basic server start/stop functionality
func TestServerBasicFunctionality(t *testing.T) {
	// Use ephemeral ports for testing
	server := NewServerWithOptions(ServerOptions{
		Port:     0, // Port 0 will assign an available port
		HTTPPort: 0, // Also use ephemeral port for HTTP
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
		Port:     0,
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

	assert.Equal(t, "testuser", claimResp.Claimant, "Response claimant should match")

	// Test non-existent claim
	resp, err = http.Get(fmt.Sprintf("%s/api/ip/2001:db8::999", baseURL))
	require.NoError(t, err, "Non-existent claim request should succeed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Non-existent claim should return 404")
}

// TestUDPClaimProcessing tests UDP claim processing functionality
func TestUDPClaimProcessing(t *testing.T) {
	// Create server with ephemeral ports
	server := NewServerWithOptions(ServerOptions{
		Port:     0,
		HTTPPort: 0,
	})

	// Start the server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	// Wait for UDP port to be assigned
	udpPort, err := server.WaitForUDPPort(5 * time.Second)
	require.NoError(t, err, "UDP port should be assigned within timeout")

	// Create UDP client
	serverAddr := &net.UDPAddr{
		IP:   net.ParseIP("::1"),
		Port: udpPort,
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	require.NoError(t, err, "UDP connection should be established")
	defer conn.Close()

	// Send a claim
	claimData := []byte("testuser")
	_, err = conn.Write(claimData)
	require.NoError(t, err, "UDP claim should be sent successfully")

	// Give server time to process the claim
	time.Sleep(200 * time.Millisecond)

	// Verify the claim was processed by checking the store
	// Note: We need to get the client's IPv6 address that the server saw
	// For testing purposes, we'll check if any claims were added
	allClaims := server.store.GetAllClaims()
	assert.NotEmpty(t, allClaims, "At least one claim should be processed")

	// Verify that one of the claims has our claimant name
	found := false
	for _, claimant := range allClaims {
		if claimant == "testuser" {
			found = true
			break
		}
	}

	assert.True(t, found, "Should find claim from 'testuser'")
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
		Port:     0,
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

	assert.Equal(t, testUser, claimResp.Claimant, "Response claimant should match")
}

// TestHTTPHandler_InvalidIPAddress tests error handling for invalid IP addresses
func TestHTTPHandler_InvalidIPAddress(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		Port:     0,
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

	var errorResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(t, err, "Error response should decode")
	assert.Contains(t, errorResp["error"], "Invalid IP address format")
}

// TestHTTPHandler_SubnetStats tests subnet statistics endpoint
func TestHTTPHandler_SubnetStats(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		Port:     0,
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
		Port:     0,
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

	var errorResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(t, err, "Error response should decode")
	assert.Contains(t, errorResp["error"], "Invalid subnet format")
}

// TestUDPServer_EmptyPayload tests handling of empty UDP payloads
func TestUDPServer_EmptyPayload(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		Port:     0,
		HTTPPort: 0,
	})

	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	udpPort, err := server.WaitForUDPPort(5 * time.Second)
	require.NoError(t, err, "UDP port should be assigned within timeout")

	// Create UDP client
	serverAddr := &net.UDPAddr{
		IP:   net.ParseIP("::1"),
		Port: udpPort,
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	require.NoError(t, err, "UDP connection should be established")
	defer conn.Close()

	// Send empty payload
	_, err = conn.Write([]byte{})
	require.NoError(t, err, "Empty UDP payload should be sent")

	// Give server time to process
	time.Sleep(200 * time.Millisecond)

	// Verify no claims were added (empty payloads should be ignored)
	allClaims := server.store.GetAllClaims()
	assert.Empty(t, allClaims, "Empty payload should not create claims")
}

// TestUDPServer_OversizedPayload tests handling of oversized UDP payloads
func TestUDPServer_OversizedPayload(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		Port:     0,
		HTTPPort: 0,
	})

	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	udpPort, err := server.WaitForUDPPort(5 * time.Second)
	require.NoError(t, err, "UDP port should be assigned within timeout")

	// Create UDP client
	serverAddr := &net.UDPAddr{
		IP:   net.ParseIP("::1"),
		Port: udpPort,
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	require.NoError(t, err, "UDP connection should be established")
	defer conn.Close()

	// Send oversized payload (> 32 bytes)
	oversizedPayload := make([]byte, 64)
	for i := range oversizedPayload {
		oversizedPayload[i] = 'A'
	}

	_, err = conn.Write(oversizedPayload)
	require.NoError(t, err, "Oversized UDP payload should be sent")

	// Give server time to process
	time.Sleep(200 * time.Millisecond)

	// Verify no claims were added (oversized payloads should be ignored)
	allClaims := server.store.GetAllClaims()
	assert.Empty(t, allClaims, "Oversized payload should not create claims")
}

// TestServerPortTimeout tests timeout behavior for port assignment
func TestServerPortTimeout(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		Port:     0,
		HTTPPort: 0,
	})

	// Test timeout without starting server
	_, err := server.WaitForUDPPort(100 * time.Millisecond)
	assert.Error(t, err, "Should timeout when server not started")
	assert.Contains(t, err.Error(), "timeout waiting for UDP port")

	_, err = server.WaitForHTTPPort(100 * time.Millisecond)
	assert.Error(t, err, "Should timeout when server not started")
	assert.Contains(t, err.Error(), "timeout waiting for HTTP port")
}

// TestServerStop_Graceful tests graceful server shutdown
func TestServerStop_Graceful(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		Port:     0,
		HTTPPort: 0,
	})

	err := server.Start()
	require.NoError(t, err, "Server should start successfully")

	// Verify both ports are assigned
	udpPort, err := server.WaitForUDPPort(2 * time.Second)
	require.NoError(t, err, "UDP port should be assigned")
	assert.Greater(t, udpPort, 0, "UDP port should be positive")

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
	// UDP writes would always succeed due to connectionless nature, so HTTP check is sufficient
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
	assert.NotNil(t, server.udpPortReady, "Server should have UDP port channel")
	assert.NotNil(t, server.httpPortReady, "Server should have HTTP port channel")
}

// TestUDPServer_ClaimOverwrite tests that UDP server properly handles claim overwrites
func TestUDPServer_ClaimOverwrite(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		Port:     0,
		HTTPPort: 0,
	})

	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()

	udpPort, err := server.WaitForUDPPort(5 * time.Second)
	require.NoError(t, err, "UDP port should be assigned within timeout")

	// Create UDP client
	serverAddr := &net.UDPAddr{
		IP:   net.ParseIP("::1"),
		Port: udpPort,
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	require.NoError(t, err, "UDP connection should be established")
	defer conn.Close()

	// Send first claim
	firstClaim := []byte("firstuser")
	_, err = conn.Write(firstClaim)
	require.NoError(t, err, "First claim should be sent successfully")

	// Give server time to process
	time.Sleep(200 * time.Millisecond)

	// Verify first claim was processed
	claimant, exists := server.store.GetClaim("::1")
	assert.True(t, exists, "First claim should exist")
	assert.Equal(t, "firstuser", claimant, "First claimant should match")

	// Send second claim to overwrite the first
	secondClaim := []byte("seconduser")
	_, err = conn.Write(secondClaim)
	require.NoError(t, err, "Second claim should be sent successfully")

	// Give server time to process
	time.Sleep(200 * time.Millisecond)

	// Verify claim was overwritten
	claimant, exists = server.store.GetClaim("::1")
	assert.True(t, exists, "Overwritten claim should exist")
	assert.Equal(t, "seconduser", claimant, "Claimant should be updated to second user")

	// Verify only one claim exists for this IP
	allClaims := server.store.GetAllClaims()
	assert.Len(t, allClaims, 1, "Should have exactly one claim")
}
