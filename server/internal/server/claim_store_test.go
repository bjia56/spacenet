package server

import (
	"fmt"
	"testing"
	"time"
)

func TestNewClaimStore(t *testing.T) {
	store := NewClaimStore()
	
	if store == nil {
		t.Fatal("NewClaimStore() returned nil")
	}
	
	if store.claims == nil {
		t.Error("NewClaimStore() did not initialize claims map")
	}
	
	if store.ipTree == nil {
		t.Error("NewClaimStore() did not initialize ipTree")
	}
	
	if store.redisClient != nil {
		t.Error("NewClaimStore() should not have Redis client for in-memory store")
	}
}

func TestClaimStore_ProcessClaim_InMemory(t *testing.T) {
	store := NewClaimStore()
	
	testCases := []struct {
		name      string
		ipAddr    string
		claimant  string
		expectErr bool
	}{
		{
			name:      "valid IPv6 claim",
			ipAddr:    "2001:db8::1",
			claimant:  "alice",
			expectErr: false,
		},
		{
			name:      "valid IPv4 claim", 
			ipAddr:    "192.168.1.1",
			claimant:  "bob", 
			expectErr: false,
		},
		{
			name:      "overwrite existing claim",
			ipAddr:    "2001:db8::1", 
			claimant:  "charlie",
			expectErr: false,
		},
		{
			name:      "empty claimant name",
			ipAddr:    "2001:db8::2",
			claimant:  "",
			expectErr: false, // Should be allowed
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := store.ProcessClaim(tc.ipAddr, tc.claimant)
			
			if tc.expectErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			
			if !tc.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if !tc.expectErr {
				// Verify claim was stored
				claimant, exists := store.GetClaim(tc.ipAddr)
				if !exists {
					t.Errorf("Claim was not stored for IP %s", tc.ipAddr)
				}
				if claimant != tc.claimant {
					t.Errorf("Expected claimant %s, got %s", tc.claimant, claimant)
				}
			}
		})
	}
}

func TestClaimStore_GetClaim(t *testing.T) {
	store := NewClaimStore()
	
	// Test non-existent claim
	claimant, exists := store.GetClaim("2001:db8::1")
	if exists {
		t.Error("Expected no claim for non-existent IP")
	}
	if claimant != "" {
		t.Errorf("Expected empty claimant, got %s", claimant)
	}
	
	// Add a claim and test retrieval
	testIP := "2001:db8::1"
	testClaimant := "alice"
	
	err := store.ProcessClaim(testIP, testClaimant)
	if err != nil {
		t.Fatalf("Failed to process claim: %v", err)
	}
	
	claimant, exists = store.GetClaim(testIP)
	if !exists {
		t.Error("Expected claim to exist")
	}
	if claimant != testClaimant {
		t.Errorf("Expected claimant %s, got %s", testClaimant, claimant)
	}
}

func TestClaimStore_GetAllClaims(t *testing.T) {
	store := NewClaimStore()
	
	// Test empty store
	claims := store.GetAllClaims()
	if len(claims) != 0 {
		t.Errorf("Expected empty claims map, got %d entries", len(claims))
	}
	
	// Add multiple claims
	testClaims := map[string]string{
		"2001:db8::1": "alice",
		"2001:db8::2": "bob", 
		"192.168.1.1": "charlie",
	}
	
	for ip, claimant := range testClaims {
		err := store.ProcessClaim(ip, claimant)
		if err != nil {
			t.Fatalf("Failed to process claim for %s: %v", ip, err)
		}
	}
	
	// Retrieve all claims
	claims = store.GetAllClaims()
	
	if len(claims) != len(testClaims) {
		t.Errorf("Expected %d claims, got %d", len(testClaims), len(claims))
	}
	
	for ip, expectedClaimant := range testClaims {
		claimant, exists := claims[ip]
		if !exists {
			t.Errorf("Missing claim for IP %s", ip)
		}
		if claimant != expectedClaimant {
			t.Errorf("Expected claimant %s for IP %s, got %s", expectedClaimant, ip, claimant)
		}
	}
	
	// Verify returned map is a copy (mutations shouldn't affect store)
	claims["test"] = "test"
	originalClaims := store.GetAllClaims()
	if _, exists := originalClaims["test"]; exists {
		t.Error("Mutations to returned claims map affected the original store")
	}
}

func TestClaimStore_GetSubnetStats(t *testing.T) {
	store := NewClaimStore()
	
	// Test non-existent subnet
	stats, exists := store.GetSubnetStats("2001:db8::/64")
	if !exists {
		t.Error("GetSubnetStats should always return true for valid subnet")
	}
	if stats == nil {
		t.Error("GetSubnetStats should return non-nil stats")
	}
	if stats.Owner != "" || stats.Percentage != 0 {
		t.Error("Empty subnet should have empty owner and 0 percentage")
	}
}

func TestClaimStore_Close(t *testing.T) {
	store := NewClaimStore()
	
	err := store.Close()
	if err != nil {
		t.Errorf("Close() returned unexpected error: %v", err)
	}
}

func TestClaimStore_ConcurrentAccess(t *testing.T) {
	store := NewClaimStore()
	
	// Test concurrent writes and reads
	done := make(chan bool)
	numGoroutines := 10
	claimsPerGoroutine := 100
	
	// Start multiple writers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			for j := 0; j < claimsPerGoroutine; j++ {
				ip := fmt.Sprintf("2001:db8::%d:%d", id, j)
				claimant := fmt.Sprintf("user_%d_%d", id, j)
				
				err := store.ProcessClaim(ip, claimant)
				if err != nil {
					t.Errorf("ProcessClaim failed: %v", err)
				}
			}
		}(i)
	}
	
	// Start multiple readers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			for j := 0; j < claimsPerGoroutine; j++ {
				ip := fmt.Sprintf("2001:db8::%d:%d", id, j)
				
				// May or may not exist depending on timing
				store.GetClaim(ip)
				store.GetAllClaims()
			}
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines*2; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out waiting for goroutines")
		}
	}
	
	// Verify final state
	claims := store.GetAllClaims()
	expectedCount := numGoroutines * claimsPerGoroutine
	
	if len(claims) != expectedCount {
		t.Errorf("Expected %d claims after concurrent operations, got %d", expectedCount, len(claims))
	}
}

func TestClaimStore_WithRedisClient_BasicOperations(t *testing.T) {
	mock := NewMockRedisClient()
	store := NewClaimStoreWithRedisClient(mock)

	// Test basic claim processing
	err := store.ProcessClaim("192.168.1.1", "alice")
	if err != nil {
		t.Fatalf("ProcessClaim failed: %v", err)
	}

	// Verify claim was stored in memory
	claimant, exists := store.GetClaim("192.168.1.1")
	if !exists {
		t.Error("Claim should exist in memory")
	}
	if claimant != "alice" {
		t.Errorf("Expected claimant 'alice', got '%s'", claimant)
	}

	// Verify Redis operations were called
	if !mock.HasCall("Set:192.168.1.1:alice") {
		t.Error("Expected Set call to Redis")
	}

	// Verify data was stored in mock Redis
	data := mock.GetData()
	if data["192.168.1.1"] != "alice" {
		t.Error("Data should be stored in mock Redis")
	}
}

func TestClaimStore_WithRedisClient_ErrorHandling(t *testing.T) {
	mock := NewMockRedisClient()
	
	// Set an error for Set operations
	mock.SetSetError(fmt.Errorf("Redis connection failed"))
	
	store := NewClaimStoreWithRedisClient(mock)

	// Try to process a claim - should fail due to Redis error
	err := store.ProcessClaim("192.168.1.1", "alice")
	if err == nil {
		t.Error("Expected error due to Redis failure")
	}

	// Verify claim was NOT stored in memory due to Redis failure
	_, exists := store.GetClaim("192.168.1.1")
	if exists {
		t.Error("Claim should not exist in memory when Redis fails")
	}
}

func TestClaimStore_WithRedisClient_LoadFromRedis(t *testing.T) {
	mock := NewMockRedisClient()
	
	// Pre-populate mock Redis with data
	mock.Set(nil, "192.168.1.1", "alice", 0)
	mock.Set(nil, "192.168.1.2", "bob", 0)
	
	// Create store - should load data from Redis
	store := NewClaimStoreWithRedisClient(mock)

	// Verify data was loaded
	claimant, exists := store.GetClaim("192.168.1.1")
	if !exists || claimant != "alice" {
		t.Error("Expected alice claim to be loaded from Redis")
	}

	claimant, exists = store.GetClaim("192.168.1.2")
	if !exists || claimant != "bob" {
		t.Error("Expected bob claim to be loaded from Redis")
	}

	// Verify Scan was called during loading
	if !mock.HasCall("Scan") {
		t.Error("Expected Scan call during Redis loading")
	}
}

func TestClaimStore_WithRedisClient_OverwriteClaim(t *testing.T) {
	mock := NewMockRedisClient()
	store := NewClaimStoreWithRedisClient(mock)

	// Set initial claim
	err := store.ProcessClaim("192.168.1.1", "alice")
	if err != nil {
		t.Fatalf("ProcessClaim failed: %v", err)
	}

	// Overwrite with new claimant
	err = store.ProcessClaim("192.168.1.1", "bob")
	if err != nil {
		t.Fatalf("ProcessClaim failed: %v", err)
	}

	// Verify new claimant
	claimant, exists := store.GetClaim("192.168.1.1")
	if !exists || claimant != "bob" {
		t.Error("Expected claim to be overwritten with 'bob'")
	}

	// Verify both Redis calls were made
	if !mock.HasCall("Set:192.168.1.1:alice") {
		t.Error("Expected first Set call to Redis")
	}
	if !mock.HasCall("Set:192.168.1.1:bob") {
		t.Error("Expected second Set call to Redis")
	}
}