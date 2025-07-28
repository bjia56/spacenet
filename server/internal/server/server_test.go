package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	port := 1337
	server := NewServer(port)
	
	if server == nil {
		t.Fatal("NewServer() returned nil")
	}
	
	if server.port != port {
		t.Errorf("Expected port %d, got %d", port, server.port)
	}
	
	if server.httpPort != 8080 {
		t.Errorf("Expected default HTTP port 8080, got %d", server.httpPort)
	}
	
	if server.store == nil {
		t.Error("NewServer() did not initialize store")
	}
	
	if server.httpHandler == nil {
		t.Error("NewServer() did not initialize HTTP handler")
	}
}

func TestNewServerWithOptions(t *testing.T) {
	opts := ServerOptions{
		Port:      1338,
		HTTPPort:  8081,
		RedisAddr: "", // Use in-memory store
	}
	
	server := NewServerWithOptions(opts)
	
	if server == nil {
		t.Fatal("NewServerWithOptions() returned nil")
	}
	
	if server.port != opts.Port {
		t.Errorf("Expected port %d, got %d", opts.Port, server.port)
	}
	
	if server.httpPort != opts.HTTPPort {
		t.Errorf("Expected HTTP port %d, got %d", opts.HTTPPort, server.httpPort)
	}
	
	if server.store == nil {
		t.Error("NewServerWithOptions() did not initialize store")
	}
	
	if server.httpHandler == nil {
		t.Error("NewServerWithOptions() did not initialize HTTP handler")
	}
}

func TestServer_StartAndStop(t *testing.T) {
	// Use ephemeral ports to avoid conflicts
	server := NewServerWithOptions(ServerOptions{
		Port:     0, // Let the OS choose UDP port
		HTTPPort: 0, // Let the OS choose HTTP port
	})
	
	// Start server
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	
	// Verify UDP listener is created
	if server.listener == nil {
		t.Error("UDP listener was not created")
	}
	
	// Verify HTTP server is created
	if server.httpServer == nil {
		t.Error("HTTP server was not created")
	}
	
	// Give the server a moment to fully start
	time.Sleep(100 * time.Millisecond)
	
	// Stop server
	server.Stop()
	
	// Verify cleanup
	if server.listener != nil {
		t.Error("UDP listener was not cleaned up")
	}
	
	if server.httpServer != nil {
		t.Error("HTTP server was not cleaned up")
	}
}

func TestServer_HTTPEndpoints(t *testing.T) {
	// Create server with ephemeral HTTP port
	server := NewServerWithOptions(ServerOptions{
		Port:     0,
		HTTPPort: 0,
	})
	
	// Start server
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()
	
	// Get the actual HTTP port
	httpAddr := server.httpServer.Addr
	if httpAddr == "" {
		t.Fatal("HTTP server address not set")
	}
	
	// Extract port from address like ":8080"
	port := strings.TrimPrefix(httpAddr, ":")
	baseURL := fmt.Sprintf("http://localhost:%s", port)
	
	// Give server time to start
	time.Sleep(200 * time.Millisecond)
	
	// Test health endpoint
	t.Run("health endpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			t.Fatalf("Failed to call health endpoint: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		
		contentType := resp.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected JSON content type, got %s", contentType)
		}
	})
	
	// Test IP endpoint with no claim
	t.Run("ip endpoint - no claim", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/ip/2001:db8::1")
		if err != nil {
			t.Fatalf("Failed to call IP endpoint: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})
	
	// Test subnet endpoint
	t.Run("subnet endpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/subnet/2001:db8::/64")
		if err != nil {
			t.Fatalf("Failed to call subnet endpoint: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestServer_UDPPacketProcessing(t *testing.T) {
	// Create server with ephemeral ports
	server := NewServerWithOptions(ServerOptions{
		Port:     0,
		HTTPPort: 0,
	})
	
	// Start server
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()
	
	// Get the actual UDP port
	if server.listener == nil {
		t.Fatal("UDP listener not created")
	}
	
	udpAddr := server.listener.LocalAddr().(*net.UDPAddr)
	
	// Give server time to start processing packets
	time.Sleep(100 * time.Millisecond)
	
	// Test sending a UDP packet
	t.Run("valid claim packet", func(t *testing.T) {
		// Create UDP client
		clientConn, err := net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			t.Fatalf("Failed to create UDP client: %v", err)
		}
		defer clientConn.Close()
		
		// Send claim packet
		claimant := "test_user"
		_, err = clientConn.Write([]byte(claimant))
		if err != nil {
			t.Fatalf("Failed to send UDP packet: %v", err)
		}
		
		// Give server time to process the packet
		time.Sleep(200 * time.Millisecond)
		
		// Check if claim was processed
		clientIP := clientConn.LocalAddr().(*net.UDPAddr).IP.String()
		storedClaimant, exists := server.store.GetClaim(clientIP)
		
		if !exists {
			t.Errorf("Claim was not processed for IP %s", clientIP)
		}
		
		if storedClaimant != claimant {
			t.Errorf("Expected claimant %s, got %s", claimant, storedClaimant)
		}
	})
	
	// Test empty packet (should be ignored)
	t.Run("empty packet", func(t *testing.T) {
		clientConn, err := net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			t.Fatalf("Failed to create UDP client: %v", err)
		}
		defer clientConn.Close()
		
		// Send empty packet
		_, err = clientConn.Write([]byte{})
		if err != nil {
			t.Fatalf("Failed to send UDP packet: %v", err)
		}
		
		// Give server time to process
		time.Sleep(100 * time.Millisecond)
		
		// Empty packets should be ignored, so no claim should exist
		clientIP := clientConn.LocalAddr().(*net.UDPAddr).IP.String()
		_, exists := server.store.GetClaim(clientIP)
		
		if exists {
			t.Error("Empty packet should not create a claim")
		}
	})
	
	// Test oversized packet (should be ignored)
	t.Run("oversized packet", func(t *testing.T) {
		clientConn, err := net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			t.Fatalf("Failed to create UDP client: %v", err)
		}
		defer clientConn.Close()
		
		// Send packet larger than 32 bytes
		largePayload := make([]byte, 64)
		for i := range largePayload {
			largePayload[i] = 'A'
		}
		
		_, err = clientConn.Write(largePayload)
		if err != nil {
			t.Fatalf("Failed to send UDP packet: %v", err)
		}
		
		// Give server time to process
		time.Sleep(100 * time.Millisecond)
		
		// Oversized packets should be ignored
		clientIP := clientConn.LocalAddr().(*net.UDPAddr).IP.String()
		_, exists := server.store.GetClaim(clientIP)
		
		if exists {
			t.Error("Oversized packet should not create a claim")
		}
	})
}

func TestServer_ConcurrentUDPRequests(t *testing.T) {
	// Create server with ephemeral ports
	server := NewServerWithOptions(ServerOptions{
		Port:     0,
		HTTPPort: 0,
	})
	
	// Start server
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()
	
	udpAddr := server.listener.LocalAddr().(*net.UDPAddr)
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Send multiple concurrent requests
	numClients := 50
	done := make(chan error, numClients)
	
	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			clientConn, err := net.DialUDP("udp", nil, udpAddr)
			if err != nil {
				done <- fmt.Errorf("client %d: failed to create UDP connection: %v", clientID, err)
				return
			}
			defer clientConn.Close()
			
			claimant := fmt.Sprintf("user_%d", clientID)
			_, err = clientConn.Write([]byte(claimant))
			if err != nil {
				done <- fmt.Errorf("client %d: failed to send packet: %v", clientID, err)
				return
			}
			
			done <- nil
		}(i)
	}
	
	// Wait for all clients to complete
	for i := 0; i < numClients; i++ {
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("Concurrent client error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out waiting for concurrent clients")
		}
	}
	
	// Give server time to process all packets
	time.Sleep(500 * time.Millisecond)
	
	// Verify that claims were processed
	// Note: Due to client port reuse, we might not get exactly numClients unique claims,
	// but we should get some claims processed successfully
	allClaims := server.store.GetAllClaims()
	if len(allClaims) == 0 {
		t.Error("No claims were processed from concurrent requests")
	}
	
	t.Logf("Processed %d claims from %d concurrent clients", len(allClaims), numClients)
}

func TestServer_StopGracefully(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{
		Port:     0,
		HTTPPort: 0,
	})
	
	// Start server
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Stop server multiple times (should be safe)
	server.Stop()
	server.Stop() // Second call should not panic
	
	// Verify resources are cleaned up
	if server.listener != nil {
		t.Error("UDP listener not cleaned up after stop")
	}
	
	if server.httpServer != nil {
		t.Error("HTTP server not cleaned up after stop")
	}
}

func TestServer_StoreIntegration(t *testing.T) {
	// Test with mock store to verify integration
	mockStore := NewMockStore()
	
	// Create server with custom options structure
	// Note: This test highlights a limitation - we can't easily inject a custom store
	// This suggests the server could benefit from dependency injection refactoring
	server := NewServerWithOptions(ServerOptions{
		Port:     0,
		HTTPPort: 0,
		RedisAddr: "", // Force in-memory store
	})
	
	// Start server
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()
	
	// Test that the server uses its store correctly
	testIP := "2001:db8::1"
	testClaimant := "integration_test"
	
	// Process a claim directly through the store
	err = server.store.ProcessClaim(testIP, testClaimant)
	if err != nil {
		t.Fatalf("Failed to process claim: %v", err)
	}
	
	// Verify claim can be retrieved
	claimant, exists := server.store.GetClaim(testIP)
	if !exists {
		t.Error("Claim not found in store")
	}
	
	if claimant != testClaimant {
		t.Errorf("Expected claimant %s, got %s", testClaimant, claimant)
	}
	
	// Test HTTP endpoint returns the claim
	if server.httpServer != nil {
		httpAddr := server.httpServer.Addr
		port := strings.TrimPrefix(httpAddr, ":")
		url := fmt.Sprintf("http://localhost:%s/api/ip/%s", port, testIP)
		
		// Give server time to start
		time.Sleep(200 * time.Millisecond)
		
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("Failed to call HTTP endpoint: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected HTTP 200, got %d", resp.StatusCode)
		}
	}
}