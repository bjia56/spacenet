package server

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)

func TestNewIPTree(t *testing.T) {
	tree := NewIPTree()
	
	if tree == nil {
		t.Fatal("NewIPTree() returned nil")
	}
	
	if tree.root == nil {
		t.Error("NewIPTree() did not initialize root node")
	}
	
	// Verify root node represents entire IPv6 space
	if tree.root.prefixLen != 0 {
		t.Errorf("Expected root prefix length 0, got %d", tree.root.prefixLen)
	}
	
	if tree.root.subnet == nil {
		t.Error("Root node should have subnet")
	}
	
	// Verify root node has correct total addresses (2^128)
	expectedTotal := new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil)
	if tree.root.totalAddresses.Cmp(expectedTotal) != 0 {
		t.Error("Root node should have 2^128 total addresses")
	}
	
	// Verify initial state
	if tree.root.claimedCount.Cmp(big.NewInt(0)) != 0 {
		t.Error("Root node should start with 0 claimed addresses")
	}
	
	if len(tree.root.claimants) != 0 {
		t.Error("Root node should start with no claimants")
	}
	
	if len(tree.root.children) != 0 {
		t.Error("Root node should start with no children")
	}
}

func TestIPTree_ProcessClaim_NewClaim(t *testing.T) {
	tree := NewIPTree()
	
	testIP := "2001:db8::1"
	testClaimant := "alice"
	
	tree.processClaim(testIP, testClaimant, "")
	
	// Verify that subnet nodes were created for standard prefix lengths
	expectedPrefixes := []int{16, 32, 48, 64, 80, 96, 112, 128}
	
	for _, prefixLen := range expectedPrefixes {
		found := false
		for _, child := range tree.root.children {
			if child.prefixLen == prefixLen {
				found = true
				
				// Verify claimant was added
				count, exists := child.claimants[testClaimant]
				if !exists {
					t.Errorf("Claimant not found in subnet with prefix %d", prefixLen)
				}
				
				if count.Cmp(big.NewInt(1)) != 0 {
					t.Errorf("Expected count 1 for prefix %d, got %s", prefixLen, count.String())
				}
				
				// Verify claimed count
				if child.claimedCount.Cmp(big.NewInt(1)) != 0 {
					t.Errorf("Expected claimed count 1 for prefix %d, got %s", prefixLen, child.claimedCount.String())
				}
				
				break
			}
		}
		
		if !found {
			t.Errorf("Subnet with prefix %d was not created", prefixLen)
		}
	}
}

func TestIPTree_ProcessClaim_UpdateExisting(t *testing.T) {
	tree := NewIPTree()
	
	testIP := "2001:db8::1"
	originalClaimant := "alice"
	newClaimant := "bob"
	
	// Process initial claim
	tree.processClaim(testIP, originalClaimant, "")
	
	// Update claim
	tree.processClaim(testIP, newClaimant, originalClaimant)
	
	// Find the /128 subnet (most specific)
	var targetNode *IPNode
	for _, child := range tree.root.children {
		if child.prefixLen == 128 {
			targetNode = child
			break
		}
	}
	
	if targetNode == nil {
		t.Fatal("Could not find /128 subnet node")
	}
	
	// Verify old claimant was removed
	if _, exists := targetNode.claimants[originalClaimant]; exists {
		t.Error("Old claimant should have been removed")
	}
	
	// Verify new claimant was added
	count, exists := targetNode.claimants[newClaimant]
	if !exists {
		t.Error("New claimant was not added")
	}
	
	if count.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("Expected count 1 for new claimant, got %s", count.String())
	}
	
	// Verify total claimed count remains 1
	if targetNode.claimedCount.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("Expected claimed count 1, got %s", targetNode.claimedCount.String())
	}
}

func TestIPTree_ProcessClaim_MultipleClaims(t *testing.T) {
	tree := NewIPTree()
	
	// Process claims for different IPs in the same /64 subnet
	testIPs := []string{
		"2001:db8::1",
		"2001:db8::2", 
		"2001:db8::3",
	}
	testClaimant := "alice"
	
	for _, ip := range testIPs {
		tree.processClaim(ip, testClaimant, "")
	}
	
	// Find the /64 subnet
	var targetNode *IPNode
	for _, child := range tree.root.children {
		if child.prefixLen == 64 {
			targetNode = child
			break
		}
	}
	
	if targetNode == nil {
		t.Fatal("Could not find /64 subnet node")
	}
	
	// Verify claimant count
	count, exists := targetNode.claimants[testClaimant]
	if !exists {
		t.Error("Claimant not found in /64 subnet")
	}
	
	expectedCount := int64(len(testIPs))
	if count.Cmp(big.NewInt(expectedCount)) != 0 {
		t.Errorf("Expected count %d, got %s", expectedCount, count.String())
	}
	
	// Verify total claimed count
	if targetNode.claimedCount.Cmp(big.NewInt(expectedCount)) != 0 {
		t.Errorf("Expected claimed count %d, got %s", expectedCount, targetNode.claimedCount.String())
	}
}

func TestIPTree_RecalculateDominant(t *testing.T) {
	tree := NewIPTree()
	
	// Create a test scenario with multiple claimants in same subnet
	baseIP := "2001:db8::"
	
	// Alice claims 3 addresses
	for i := 1; i <= 3; i++ {
		tree.processClaim(baseIP+fmt.Sprintf("%d", i), "alice", "")
	}
	
	// Bob claims 2 addresses  
	for i := 4; i <= 5; i++ {
		tree.processClaim(baseIP+fmt.Sprintf("%d", i), "bob", "")
	}
	
	// Find the /64 subnet
	var targetNode *IPNode
	for _, child := range tree.root.children {
		if child.prefixLen == 64 {
			targetNode = child
			break
		}
	}
	
	if targetNode == nil {
		t.Fatal("Could not find /64 subnet node")
	}
	
	// Alice should be dominant with 3 out of 5 claims
	if targetNode.dominantClaimant != "alice" {
		t.Errorf("Expected dominant claimant 'alice', got '%s'", targetNode.dominantClaimant)
	}
	
	// Check percentage calculation
	if targetNode.dominantPercentage <= 0 {
		t.Error("Dominant percentage should be greater than 0")
	}
	
	// The percentage should be very small since /64 has 2^64 total addresses
	// but should be non-zero since Alice has 3 addresses
	if targetNode.dominantPercentage >= 1.0 {
		t.Errorf("Dominant percentage seems too high: %f", targetNode.dominantPercentage)
	}
}

func TestIPTree_RecalculateDominant_Tie(t *testing.T) {
	tree := NewIPTree()
	
	// Create a tie scenario
	tree.processClaim("2001:db8::1", "bob", "")   // Lexicographically first
	tree.processClaim("2001:db8::2", "alice", "") // Lexicographically second
	
	// Find the /64 subnet
	var targetNode *IPNode
	for _, child := range tree.root.children {
		if child.prefixLen == 64 {
			targetNode = child
			break
		}
	}
	
	if targetNode == nil {
		t.Fatal("Could not find /64 subnet node")
	}
	
	// In case of tie, lexicographically smaller name should win
	if targetNode.dominantClaimant != "alice" {
		t.Errorf("Expected dominant claimant 'alice' (lexicographically smaller), got '%s'", targetNode.dominantClaimant)
	}
}

func TestIPTree_GetSubnetStats_ExistingSubnet(t *testing.T) {
	tree := NewIPTree() 
	
	// Create claims to generate statistics
	testSubnet := "2001:db8::/64"
	testClaimant := "alice"
	
	// Need to claim enough addresses to exceed 50% threshold
	// Since /64 has 2^64 addresses, we'll use the /128 test which is more manageable
	testIP := "2001:db8::1"
	tree.processClaim(testIP, testClaimant, "")
	
	// Test with /128 subnet (exact match)
	stats, ok := tree.GetSubnetStats("2001:db8::1/128")
	if !ok {
		t.Error("GetSubnetStats should return true for valid subnet")
	}
	
	if stats == nil {
		t.Error("GetSubnetStats should return non-nil stats")
	}
	
	// For /128, the single address should give 100% control
	if stats.Owner != testClaimant {
		t.Errorf("Expected owner '%s', got '%s'", testClaimant, stats.Owner)
	}
	
	if stats.Percentage != 100.0 {
		t.Errorf("Expected 100%% control of /128, got %f", stats.Percentage)
	}
}

func TestIPTree_GetSubnetStats_NonExistentSubnet(t *testing.T) {
	tree := NewIPTree()
	
	// Test subnet with no claims
	stats, ok := tree.GetSubnetStats("2001:db8::/64")
	if !ok {
		t.Error("GetSubnetStats should return true even for subnets with no claims")
	}
	
	if stats == nil {
		t.Error("GetSubnetStats should return non-nil stats")
	}
	
	if stats.Owner != "" {
		t.Errorf("Expected empty owner for unclaimed subnet, got '%s'", stats.Owner)
	}
	
	if stats.Percentage != 0 {
		t.Errorf("Expected 0%% for unclaimed subnet, got %f", stats.Percentage)
	}
}

func TestIPTree_GetSubnetStats_InvalidSubnet(t *testing.T) {
	tree := NewIPTree()
	
	invalidSubnets := []string{
		"invalid-subnet",
		"2001:db8::/",
		"not-a-cidr",
		"2001:db8::",  // Missing prefix
	}
	
	for _, subnet := range invalidSubnets {
		stats, ok := tree.GetSubnetStats(subnet)
		if ok {
			t.Errorf("GetSubnetStats should return false for invalid subnet '%s'", subnet)
		}
		
		if stats != nil {
			t.Errorf("GetSubnetStats should return nil stats for invalid subnet '%s'", subnet)
		}
	}
}

func TestIPTree_GetSubnetStats_NonStandardPrefix(t *testing.T) {
	tree := NewIPTree()
	
	// Claim an address
	testIP := "2001:db8::1"
	testClaimant := "alice"
	tree.processClaim(testIP, testClaimant, "")
	
	// Test with non-standard prefix (should round up to nearest standard)
	// /65 should round up to /80
	stats, ok := tree.GetSubnetStats("2001:db8::/65")
	if !ok {
		t.Error("GetSubnetStats should handle non-standard prefixes")
	}
	
	if stats == nil {
		t.Error("GetSubnetStats should return non-nil stats")
	}
	
	// The result should be based on the /80 subnet that contains our IP
	// Since we only have one address in a /80, percentage should be very small
	if stats.Percentage > 50.0 {
		t.Errorf("Expected low percentage for large subnet, got %f", stats.Percentage) 
	}
}

func TestIPTree_GetSubnetStats_DominanceThreshold(t *testing.T) {
	tree := NewIPTree()
	
	// Create a scenario where dominant claimant has <= 50%
	// Use /128 subnets for precise control
	ip1 := "2001:db8::1"
	ip2 := "2001:db8::2"
	
	tree.processClaim(ip1, "alice", "")
	tree.processClaim(ip2, "bob", "")
	
	// Test /112 subnet which should contain both addresses
	// Each claimant has 50%, so no one should be dominant
	stats, ok := tree.GetSubnetStats("2001:db8::/112")
	if !ok {
		t.Error("GetSubnetStats should return true")
	}
	
	// With 50/50 split, no one should be dominant (need >50%)
	if stats.Owner != "" {
		t.Errorf("Expected no dominant owner with 50/50 split, got '%s'", stats.Owner)
	}
	
	if stats.Percentage != 0 {
		t.Errorf("Expected 0%% with no dominant owner, got %f", stats.Percentage)
	}
}

func TestIPTree_ConcurrentAccess(t *testing.T) {
	tree := NewIPTree()
	
	// Test concurrent processing of claims
	done := make(chan bool)
	numGoroutines := 10
	claimsPerGoroutine := 50
	
	// Start multiple goroutines processing claims
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			for j := 0; j < claimsPerGoroutine; j++ {
				ip := fmt.Sprintf("2001:db8::%d:%d", id, j)
				claimant := fmt.Sprintf("user_%d", id)
				
				tree.processClaim(ip, claimant, "")
			}
		}(i)
	}
	
	// Start goroutines reading statistics
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			for j := 0; j < claimsPerGoroutine; j++ {
				subnet := fmt.Sprintf("2001:db8::%d::/128", id)
				tree.GetSubnetStats(subnet)
			}
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines*2; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out waiting for concurrent operations")
		}
	}
	
	// Verify tree integrity - should have created many subnet nodes
	if len(tree.root.children) == 0 {
		t.Error("Concurrent operations should have created subnet nodes")
	}
}