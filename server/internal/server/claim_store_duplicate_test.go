package server

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClaimStore_DuplicateClaimHandling tests that duplicate claims by the same user
// don't inflate statistics or cause issues
func TestClaimStore_DuplicateClaimHandling(t *testing.T) {
	store := NewClaimStore()
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Error closing store: %v", err)
		}
	}()

	testIP := "2001:db8::1"
	testUser := "testuser"

	// Make initial claim
	err := store.ProcessClaim(testIP, testUser)
	require.NoError(t, err, "Initial claim should succeed")

	// Verify claim exists
	claimant, exists := store.GetClaim(testIP)
	assert.True(t, exists, "Initial claim should exist")
	assert.Equal(t, testUser, claimant, "Initial claimant should match")

	// Get initial subnet stats
	stats, ok := store.GetSubnetStats("2001:db8::1/128")
	require.True(t, ok, "Should be able to get subnet stats")
	initialPercentage := stats.Percentage

	// Make duplicate claim (same user, same IP)
	err = store.ProcessClaim(testIP, testUser)
	require.NoError(t, err, "Duplicate claim should not error")

	// Verify claim still exists and hasn't changed
	claimant, exists = store.GetClaim(testIP)
	assert.True(t, exists, "Claim should still exist after duplicate")
	assert.Equal(t, testUser, claimant, "Claimant should still be the same")

	// Verify we still have only one claim total
	allClaims := store.GetAllClaims()
	assert.Len(t, allClaims, 1, "Should still have exactly one claim")

	// Most importantly: verify stats haven't inflated
	stats, ok = store.GetSubnetStats("2001:db8::1/128")
	require.True(t, ok, "Should still be able to get subnet stats")
	assert.Equal(t, initialPercentage, stats.Percentage, "Percentage should not change after duplicate claim")
	assert.LessOrEqual(t, stats.Percentage, 100.0, "Percentage should never exceed 100%")
}

// TestClaimStore_MultipleDuplicateClaims tests that multiple duplicate claims
// don't cause exponential inflation of statistics
func TestClaimStore_MultipleDuplicateClaims(t *testing.T) {
	store := NewClaimStore()
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Error closing store: %v", err)
		}
	}()

	testIP := "2001:db8::1"
	testUser := "testuser"

	// Make initial claim
	err := store.ProcessClaim(testIP, testUser)
	require.NoError(t, err, "Initial claim should succeed")

	// Get initial subnet stats
	stats, ok := store.GetSubnetStats("2001:db8::1/128")
	require.True(t, ok, "Should be able to get subnet stats")
	initialPercentage := stats.Percentage

	// Make many duplicate claims
	for i := 0; i < 10; i++ {
		err = store.ProcessClaim(testIP, testUser)
		require.NoError(t, err, "Duplicate claim %d should not error", i+1)
	}

	// Verify we still have only one claim total
	allClaims := store.GetAllClaims()
	assert.Len(t, allClaims, 1, "Should still have exactly one claim after multiple duplicates")

	// Verify stats remain unchanged
	stats, ok = store.GetSubnetStats("2001:db8::1/128")
	require.True(t, ok, "Should still be able to get subnet stats")
	assert.Equal(t, initialPercentage, stats.Percentage, "Percentage should remain unchanged after multiple duplicates")
	assert.LessOrEqual(t, stats.Percentage, 100.0, "Percentage should never exceed 100%")
}

// TestClaimStore_DuplicateVsDifferentUser tests that duplicate claims are handled
// differently from legitimate claim takeovers
func TestClaimStore_DuplicateVsDifferentUser(t *testing.T) {
	store := NewClaimStore()
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Error closing store: %v", err)
		}
	}()

	testIP := "2001:db8::1"
	user1 := "user1"
	user2 := "user2"

	// User1 makes initial claim
	err := store.ProcessClaim(testIP, user1)
	require.NoError(t, err, "Initial claim should succeed")

	claimant, exists := store.GetClaim(testIP)
	assert.True(t, exists, "Initial claim should exist")
	assert.Equal(t, user1, claimant, "Initial claimant should be user1")

	// User1 makes duplicate claim - should be ignored
	err = store.ProcessClaim(testIP, user1)
	require.NoError(t, err, "Duplicate claim should not error")

	claimant, exists = store.GetClaim(testIP)
	assert.True(t, exists, "Claim should still exist")
	assert.Equal(t, user1, claimant, "Claimant should still be user1 after duplicate")

	// User2 makes legitimate takeover claim - should work
	err = store.ProcessClaim(testIP, user2)
	require.NoError(t, err, "Takeover claim should not error")

	claimant, exists = store.GetClaim(testIP)
	assert.True(t, exists, "Claim should exist after takeover")
	assert.Equal(t, user2, claimant, "Claimant should now be user2 after takeover")

	// Verify we still have only one claim total
	allClaims := store.GetAllClaims()
	assert.Len(t, allClaims, 1, "Should have exactly one claim after takeover")
}

// TestClaimStore_SubnetPercentageAccuracy tests that subnet percentages are
// calculated correctly and never exceed 100%
func TestClaimStore_SubnetPercentageAccuracy(t *testing.T) {
	store := NewClaimStore()
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Error closing store: %v", err)
		}
	}()

	testUser := "testuser"

	// Fill an entire /124 subnet (16 addresses) with claims
	for i := 0; i < 16; i++ {
		ip := fmt.Sprintf("2001:db8::ff%x", i)
		err := store.ProcessClaim(ip, testUser)
		require.NoError(t, err, "Claim %d should succeed", i)
	}

	// Make duplicate claims for some addresses
	for i := 0; i < 5; i++ {
		ip := fmt.Sprintf("2001:db8::ff%x", i)
		err := store.ProcessClaim(ip, testUser)
		require.NoError(t, err, "Duplicate claim %d should not error", i)
	}

	// Check subnet stats at various levels
	subnetsToCheck := []string{
		"2001:db8::/32",
		"2001:db8::/48",
		"2001:db8::/64",
		"2001:db8::/80",
		"2001:db8::/96",
		"2001:db8::/112",
	}

	for _, subnet := range subnetsToCheck {
		stats, ok := store.GetSubnetStats(subnet)
		require.True(t, ok, "Should be able to get stats for %s", subnet)
		assert.LessOrEqual(t, stats.Percentage, 100.0, "Percentage for %s should not exceed 100%", subnet)

		t.Logf("Subnet %s: Owner=%s, Percentage=%.6f%%", subnet, stats.Owner, stats.Percentage)
	}

	// Verify total claim count is still 16 (not inflated by duplicates)
	allClaims := store.GetAllClaims()
	assert.Len(t, allClaims, 16, "Should have exactly 16 claims, not more due to duplicates")
}

// TestIPTree_ProcessClaimDuplicateHandling tests the IPTree processClaim method directly
// to ensure duplicate handling works at the tree level
func TestIPTree_ProcessClaimDuplicateHandling(t *testing.T) {
	tree := NewIPTree()

	testIP := "2001:db8::1"
	testUser := "testuser"

	// Initial claim
	tree.processClaim(testIP, testUser, "")

	// Get stats after initial claim
	stats, ok := tree.GetSubnetStats("2001:db8::1/128")
	require.True(t, ok, "Should get stats after initial claim")
	initialPercentage := stats.Percentage

	// Duplicate claim (same user as old claimant)
	tree.processClaim(testIP, testUser, testUser)

	// Verify stats haven't changed
	stats, ok = tree.GetSubnetStats("2001:db8::1/128")
	require.True(t, ok, "Should still get stats after duplicate")
	assert.Equal(t, initialPercentage, stats.Percentage, "Percentage should not change after duplicate claim")
	// Note: Owner may be empty if percentage <= 50% due to IPTree logic
	if initialPercentage > 50.0 {
		assert.Equal(t, testUser, stats.Owner, "Owner should remain the same")
	}
}

// TestIPTree_ProcessClaimTakeoverHandling tests that legitimate takeovers work correctly
func TestIPTree_ProcessClaimTakeoverHandling(t *testing.T) {
	tree := NewIPTree()

	testIP := "2001:db8::1"
	user1 := "user1"
	user2 := "user2"

	// Initial claim by user1
	tree.processClaim(testIP, user1, "")

	stats, ok := tree.GetSubnetStats("2001:db8::1/128")
	require.True(t, ok, "Should get stats after initial claim")
	// Single claim in /128 should be 100%, so owner should be set
	assert.Equal(t, user1, stats.Owner, "Initial owner should be user1")
	assert.Equal(t, 100.0, stats.Percentage, "Single claim in /128 should be 100%")

	// Takeover by user2
	tree.processClaim(testIP, user2, user1)

	stats, ok = tree.GetSubnetStats("2001:db8::1/128")
	require.True(t, ok, "Should get stats after takeover")
	assert.Equal(t, user2, stats.Owner, "Owner should change to user2 after takeover")
	assert.Equal(t, 100.0, stats.Percentage, "Single claim in /128 should still be 100%")

	// Percentage should remain reasonable (not exceed 100%)
	assert.LessOrEqual(t, stats.Percentage, 100.0, "Percentage should not exceed 100% after takeover")
}

// TestClaimStore_ConcurrentDuplicateClaims tests that concurrent duplicate claims
// don't cause race conditions or stat inflation
func TestClaimStore_ConcurrentDuplicateClaims(t *testing.T) {
	store := NewClaimStore()
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Error closing store: %v", err)
		}
	}()

	testIP := "2001:db8::1"
	testUser := "testuser"

	// Make initial claim
	err := store.ProcessClaim(testIP, testUser)
	require.NoError(t, err, "Initial claim should succeed")

	// Launch multiple goroutines making duplicate claims
	const numGoroutines = 10
	const claimsPerGoroutine = 5

	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < claimsPerGoroutine; j++ {
				err := store.ProcessClaim(testIP, testUser)
				assert.NoError(t, err, "Concurrent duplicate claim should not error")
			}
		}()
	}

	// Wait for all goroutines to complete
	for range numGoroutines {
		<-done
	}

	// Verify we still have only one claim
	allClaims := store.GetAllClaims()
	assert.Len(t, allClaims, 1, "Should still have exactly one claim after concurrent duplicates")

	// Verify claim is still correct
	claimant, exists := store.GetClaim(testIP)
	assert.True(t, exists, "Claim should still exist")
	assert.Equal(t, testUser, claimant, "Claimant should still be correct")

	// Verify stats are reasonable
	stats, ok := store.GetSubnetStats("2001:db8::1/128")
	require.True(t, ok, "Should get stats after concurrent duplicates")
	assert.LessOrEqual(t, stats.Percentage, 100.0, "Percentage should not exceed 100% after concurrent duplicates")
}
