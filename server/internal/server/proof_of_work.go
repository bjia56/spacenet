package server

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"net"
	"strings"
)

// ProofOfWork represents a proof of work challenge and solution
type ProofOfWork struct {
	Target     net.IP  // IPv6 address being claimed
	Claimant   string  // Name of the claimant
	Nonce      uint64  // Nonce used to solve the challenge
	Difficulty uint8   // Difficulty level (number of leading zero bits required)
}

// Hash computes the SHA-256 hash of the proof of work data
func (pow *ProofOfWork) Hash() [32]byte {
	// Create input data: target_ip + claimant + nonce
	data := make([]byte, 0, 16+len(pow.Claimant)+8)
	
	// Add IPv6 address (16 bytes)
	data = append(data, pow.Target.To16()...)
	
	// Add claimant name
	data = append(data, []byte(pow.Claimant)...)
	
	// Add nonce (8 bytes, big endian)
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, pow.Nonce)
	data = append(data, nonceBytes...)
	
	return sha256.Sum256(data)
}

// IsValid checks if the proof of work satisfies the difficulty requirement
func (pow *ProofOfWork) IsValid() bool {
	hash := pow.Hash()
	
	// Count leading zero bits
	leadingZeros := 0
	for i := 0; i < 32; i++ {
		b := hash[i]
		if b == 0 {
			leadingZeros += 8
		} else {
			// Count leading zeros in this byte
			for j := 7; j >= 0; j-- {
				if (b>>j)&1 == 0 {
					leadingZeros++
				} else {
					break
				}
			}
			break
		}
	}
	
	return leadingZeros >= int(pow.Difficulty)
}

// CalculateDifficulty determines the required difficulty for claiming an address
func (store *ClaimStore) CalculateDifficulty(targetIP string) uint8 {
	const (
		baseDifficulty   = 8  // Base difficulty (8 leading zero bits)
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
		
		difficulty += uint8(contiguous * contiguityBonus)
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
func (store *ClaimStore) ValidateProofOfWork(pow *ProofOfWork) error {
	// Check if the proof of work is valid
	if !pow.IsValid() {
		return fmt.Errorf("proof of work does not satisfy difficulty requirement")
	}
	
	// Check if the difficulty matches what's required
	requiredDifficulty := store.CalculateDifficulty(pow.Target.String())
	if pow.Difficulty < requiredDifficulty {
		return fmt.Errorf("insufficient difficulty: required %d, provided %d", requiredDifficulty, pow.Difficulty)
	}
	
	return nil
}

// SolveProofOfWork attempts to solve a proof of work challenge (for testing/client use)
func SolveProofOfWork(target net.IP, claimant string, difficulty uint8, maxAttempts uint64) (*ProofOfWork, error) {
	pow := &ProofOfWork{
		Target:     target,
		Claimant:   claimant,
		Difficulty: difficulty,
	}
	
	for nonce := uint64(0); nonce < maxAttempts; nonce++ {
		pow.Nonce = nonce
		if pow.IsValid() {
			return pow, nil
		}
	}
	
	return nil, fmt.Errorf("could not solve proof of work within %d attempts", maxAttempts)
}