package server

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIPTree_PercentageCalculationAccuracy tests that percentage calculations
// are accurate and never exceed 100% under various scenarios
func TestIPTree_PercentageCalculationAccuracy(t *testing.T) {

	testCases := []struct {
		name        string
		claims      []struct{ ip, user string }
		subnet      string
		expectOwner string
		maxPercent  float64
	}{
		{
			name: "Single claim in /128",
			claims: []struct{ ip, user string }{
				{"2001:db8::1", "user1"},
			},
			subnet:      "2001:db8::1/128",
			expectOwner: "user1",
			maxPercent:  100.0,
		},
		{
			name: "Multiple claims different users in /112",
			claims: []struct{ ip, user string }{
				{"2001:db8::1", "user1"},
				{"2001:db8::2", "user2"},
				{"2001:db8::3", "user1"},
				{"2001:db8::4", "user1"},
			},
			subnet:     "2001:db8::/112",
			maxPercent: 100.0,
		},
		{
			name: "Same user multiple IPs",
			claims: []struct{ ip, user string }{
				{"2001:db8::1", "user1"},
				{"2001:db8::2", "user1"},
				{"2001:db8::3", "user1"},
			},
			subnet:      "2001:db8::/112",
			expectOwner: "", // 3 claims in 2^16 subnet = tiny percentage, no owner
			maxPercent:  100.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Fresh tree for each test
			testTree := NewIPTree()

			// Process all claims
			for _, claim := range tc.claims {
				testTree.processClaim(claim.ip, claim.user, "")
			}

			// Get stats
			stats, ok := testTree.GetSubnetStats(tc.subnet)
			require.True(t, ok, "Should be able to get stats for %s", tc.subnet)

			// Check percentage bounds
			assert.GreaterOrEqual(t, stats.Percentage, 0.0, "Percentage should be non-negative")
			assert.LessOrEqual(t, stats.Percentage, tc.maxPercent, "Percentage should not exceed %f", tc.maxPercent)

			// Check expected owner if specified
			if tc.expectOwner != "" {
				assert.Equal(t, tc.expectOwner, stats.Owner, "Owner should match expected")
			} else if tc.expectOwner == "" {
				// If we expect empty owner, the percentage should be <= 50%
				assert.LessOrEqual(t, stats.Percentage, 50.0, "Empty owner should mean percentage <= 50%")
			}

			t.Logf("Subnet %s: Owner=%s, Percentage=%.6f%%", tc.subnet, stats.Owner, stats.Percentage)
		})
	}
}

// TestIPTree_PercentageWithDuplicates specifically tests percentage calculations
// when duplicate claims are involved
func TestIPTree_PercentageWithDuplicates(t *testing.T) {
	tree := NewIPTree()

	testIP := "2001:db8::1"
	testUser := "user1"

	// Make initial claim
	tree.processClaim(testIP, testUser, "")

	// Get initial percentage
	stats, ok := tree.GetSubnetStats("2001:db8::1/128")
	require.True(t, ok, "Should get stats after initial claim")
	initialPercentage := stats.Percentage

	// Make multiple duplicate claims
	for i := 0; i < 10; i++ {
		tree.processClaim(testIP, testUser, testUser) // Duplicate claim

		stats, ok = tree.GetSubnetStats("2001:db8::1/128")
		require.True(t, ok, "Should get stats after duplicate %d", i+1)

		// Percentage should remain exactly the same
		assert.Equal(t, initialPercentage, stats.Percentage,
			"Percentage should not change after duplicate %d", i+1)
		assert.LessOrEqual(t, stats.Percentage, 100.0,
			"Percentage should not exceed 100% after duplicate %d", i+1)
	}
}

// TestIPTree_PercentageConsistency tests that percentage calculations are
// consistent across different subnet levels
func TestIPTree_PercentageConsistency(t *testing.T) {
	tree := NewIPTree()

	// Create claims in a specific pattern
	basePrefix := "2001:db8:1234:5678:9abc:def0::"
	testUser := "testuser"

	// Claim several addresses in the same /112
	for i := 0; i < 8; i++ {
		ip := fmt.Sprintf("%s%x", basePrefix, i)
		tree.processClaim(ip, testUser, "")
	}

	// Check stats at different subnet levels
	subnets := []string{
		"2001:db8:1234:5678:9abc:def0::/112",
		"2001:db8:1234:5678:9abc::/96",
		"2001:db8:1234:5678::/80",
		"2001:db8:1234::/64",
		"2001:db8::/32",
	}

	for _, subnet := range subnets {
		stats, ok := tree.GetSubnetStats(subnet)
		require.True(t, ok, "Should get stats for %s", subnet)

		// All percentages should be reasonable
		assert.GreaterOrEqual(t, stats.Percentage, 0.0, "Percentage should be non-negative for %s", subnet)
		assert.LessOrEqual(t, stats.Percentage, 100.0, "Percentage should not exceed 100% for %s", subnet)

		// If there's a dominant owner, it should be our test user
		if stats.Percentage > 50.0 {
			assert.Equal(t, testUser, stats.Owner, "Dominant owner should be testuser for %s", subnet)
		}

		t.Logf("Subnet %s: Owner=%s, Percentage=%.8f%%", subnet, stats.Owner, stats.Percentage)
	}
}

// TestIPTree_LargeScalePercentageAccuracy tests percentage accuracy with many claims
func TestIPTree_LargeScalePercentageAccuracy(t *testing.T) {
	tree := NewIPTree()

	// Create a large number of claims across different users
	const totalClaims = 1000
	users := []string{"alice", "bob", "charlie", "david", "eve"}

	for i := 0; i < totalClaims; i++ {
		ip := fmt.Sprintf("2001:db8::%04x", i)
		user := users[i%len(users)]
		tree.processClaim(ip, user, "")
	}

	// Check various subnet levels
	subnets := []string{
		"2001:db8::/32",
		"2001:db8::/48",
		"2001:db8::/64",
		"2001:db8::/80",
	}

	for _, subnet := range subnets {
		stats, ok := tree.GetSubnetStats(subnet)
		require.True(t, ok, "Should get stats for %s", subnet)

		// Percentage should never exceed 100%
		assert.LessOrEqual(t, stats.Percentage, 100.0, "Percentage should not exceed 100% for %s", subnet)
		assert.GreaterOrEqual(t, stats.Percentage, 0.0, "Percentage should be non-negative for %s", subnet)

		t.Logf("Large scale - Subnet %s: Owner=%s, Percentage=%.6f%%", subnet, stats.Owner, stats.Percentage)
	}
}

// TestIPTree_EdgeCasePercentages tests edge cases that might cause percentage issues
func TestIPTree_EdgeCasePercentages(t *testing.T) {
	testCases := []struct {
		name   string
		claims func(*IPTree)
		subnet string
	}{
		{
			name: "Single claim then many duplicates",
			claims: func(tree *IPTree) {
				ip := "2001:db8::1"
				user := "user1"
				tree.processClaim(ip, user, "")
				// Many duplicates
				for i := 0; i < 100; i++ {
					tree.processClaim(ip, user, user)
				}
			},
			subnet: "2001:db8::1/128",
		},
		{
			name: "Claim then takeover then duplicates",
			claims: func(tree *IPTree) {
				ip := "2001:db8::1"
				user1 := "user1"
				user2 := "user2"

				tree.processClaim(ip, user1, "")
				tree.processClaim(ip, user2, user1) // Takeover
				// Duplicates by new owner
				for i := 0; i < 50; i++ {
					tree.processClaim(ip, user2, user2)
				}
			},
			subnet: "2001:db8::1/128",
		},
		{
			name: "Multiple takeovers with duplicates",
			claims: func(tree *IPTree) {
				ip := "2001:db8::1"
				users := []string{"alice", "bob", "charlie"}

				// Initial claim
				tree.processClaim(ip, users[0], "")

				// Series of takeovers with duplicates
				for i := 1; i < len(users); i++ {
					tree.processClaim(ip, users[i], users[i-1]) // Takeover
					// Make some duplicates
					for j := 0; j < 10; j++ {
						tree.processClaim(ip, users[i], users[i])
					}
				}
			},
			subnet: "2001:db8::1/128",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := NewIPTree()

			// Execute the claim pattern
			tc.claims(tree)

			// Check resulting stats
			stats, ok := tree.GetSubnetStats(tc.subnet)
			require.True(t, ok, "Should get stats for %s", tc.subnet)

			// Critical checks
			assert.LessOrEqual(t, stats.Percentage, 100.0, "Percentage should not exceed 100%")
			assert.GreaterOrEqual(t, stats.Percentage, 0.0, "Percentage should be non-negative")
			assert.False(t, math.IsNaN(stats.Percentage), "Percentage should not be NaN")
			assert.False(t, math.IsInf(stats.Percentage, 0), "Percentage should not be infinite")

			t.Logf("Edge case %s - Subnet %s: Owner=%s, Percentage=%.6f%%",
				tc.name, tc.subnet, stats.Owner, stats.Percentage)
		})
	}
}

// TestClaimStore_PercentageAccuracyWithSQLite tests that percentage calculations
// remain accurate when using SQLite persistence
func TestClaimStore_PercentageAccuracyWithSQLite(t *testing.T) {
	// Test with temporary SQLite database
	store, err := NewClaimStoreWithSQLite(":memory:")
	require.NoError(t, err, "Should create SQLite store")
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Error closing store: %v", err)
		}
	}()

	testUser := "testuser"

	// Make several claims
	claims := []string{
		"2001:db8::1",
		"2001:db8::2",
		"2001:db8::3",
	}

	for _, ip := range claims {
		err := store.ProcessClaim(ip, testUser)
		require.NoError(t, err, "Claim should succeed")

		// Make duplicate of each claim to test the fix
		err = store.ProcessClaim(ip, testUser)
		require.NoError(t, err, "Duplicate should not error")
	}

	// Check subnet stats
	stats, ok := store.GetSubnetStats("2001:db8::/112")
	require.True(t, ok, "Should get subnet stats")

	assert.LessOrEqual(t, stats.Percentage, 100.0, "Percentage should not exceed 100% with SQLite")
	assert.GreaterOrEqual(t, stats.Percentage, 0.0, "Percentage should be non-negative with SQLite")

	// Verify total claims count is still correct (not inflated by duplicates)
	allClaims := store.GetAllClaims()
	assert.Len(t, allClaims, len(claims), "Should have correct number of unique claims")

	t.Logf("SQLite store - Subnet stats: Owner=%s, Percentage=%.6f%%", stats.Owner, stats.Percentage)
}
