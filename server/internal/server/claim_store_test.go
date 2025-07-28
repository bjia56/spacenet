package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
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

// MockRedisClient for testing Redis functionality
type MockRedisClient struct {
	data map[string]string
	pingError error
	getError error
	setError error
	scanKeys []string
	scanError error
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string]string),
	}
}

func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	if m.pingError != nil {
		cmd.SetErr(m.pingError)
	} else {
		cmd.SetVal("PONG")
	}
	return cmd
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx, "get", key)
	if m.getError != nil {
		cmd.SetErr(m.getError)
	} else if val, exists := m.data[key]; exists {
		cmd.SetVal(val)
	} else {
		cmd.SetErr(redis.Nil)
	}
	return cmd
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx, "set", key, value)
	if m.setError != nil {
		cmd.SetErr(m.setError)
	} else {
		m.data[key] = fmt.Sprintf("%v", value)
		cmd.SetVal("OK")
	}
	return cmd
}

func (m *MockRedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	cmd := redis.NewScanCmd(ctx, "scan", cursor, "match", match, "count", count)
	if m.scanError != nil {
		cmd.SetErr(m.scanError)
	} else {
		keys := make([]string, 0)
		for k := range m.data {
			keys = append(keys, k)
		}
		cmd.SetVal(keys, 0) // Return all keys with cursor 0 (no more pages)
	}
	return cmd
}

func (m *MockRedisClient) Close() error {
	return nil
}

// Note: The actual Redis tests would require a more complex setup
// This demonstrates the testing approach but would need proper Redis mocking
// or integration test setup for full Redis functionality testing