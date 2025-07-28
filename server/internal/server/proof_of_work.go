package server

import (
	"fmt"
	"net"

	"github.com/bjia56/spacenet/server/api"
)

// CalculateDifficulty determines the required difficulty for claiming an address
func (store *ClaimStore) CalculateDifficulty(targetIP string) uint8 {
	const (
		baseDifficulty  = 8  // Base difficulty (8 leading zero bits)
		claimBonus      = 4  // Additional difficulty if address is already claimed
		maxContiguity   = 16 // Maximum contiguous addresses to consider
		contiguityBonus = 2  // Additional difficulty per contiguous address
	)

	difficulty := baseDifficulty

	// Check if address is already claimed
	store.mutex.RLock()
	currentClaimant, exists := store.claims[targetIP]
	store.mutex.RUnlock()

	if exists {
		difficulty += claimBonus

		// Calculate contiguous addresses owned by current claimant
		contiguous := store.countContiguousAddresses(targetIP, currentClaimant)
		if contiguous > maxContiguity {
			contiguous = maxContiguity
		}

		difficulty += contiguous * contiguityBonus
	}

	// Cap difficulty at reasonable maximum (28 bits = ~268 million hashes expected)
	if difficulty > 28 {
		difficulty = 28
	}

	return uint8(difficulty)
}

// countContiguousAddresses counts how many addresses contiguous to the target
// are owned by the specified claimant within a /124 block
func (store *ClaimStore) countContiguousAddresses(targetIP string, claimant string) int {
	ip := net.ParseIP(targetIP)
	if ip == nil {
		return 0
	}

	// Convert to IPv6
	ipv6 := ip.To16()
	if ipv6 == nil {
		return 0
	}

	// Create /124 network (4 bits for host addresses = 16 possible addresses)
	mask := net.CIDRMask(124, 128)
	network := &net.IPNet{
		IP:   ipv6.Mask(mask),
		Mask: mask,
	}

	count := 0

	// Check all 16 addresses in the /124 block
	for i := 0; i < 16; i++ {
		// Create IP by modifying the last 4 bits
		testIP := make(net.IP, 16)
		copy(testIP, network.IP)

		// Set the last 4 bits to i
		testIP[15] = network.IP[15] | byte(i)

		// Skip the target IP itself
		if testIP.Equal(ipv6) {
			continue
		}

		// Check if this address is owned by the claimant
		if owner, exists := store.claims[testIP.String()]; exists && owner == claimant {
			count++
		}
	}

	return count
}

// ValidateProofOfWork validates a proof of work submission
func (store *ClaimStore) ValidateProofOfWork(pow *api.ProofOfWork) error {
	// Get current difficulty for the target address
	requiredDifficulty := store.CalculateDifficulty(pow.Target.String())
	if !pow.IsValid(requiredDifficulty) {
		return fmt.Errorf("invalid proof of work: insufficient difficulty")
	}

	return nil
}
