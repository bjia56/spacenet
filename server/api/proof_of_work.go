// Package api defines shared data structures between the SpaceNet server and client
package api

import (
	"crypto/sha256"
	"fmt"
	"net"
)

// ProofOfWork represents a proof of work challenge and solution
type ProofOfWork struct {
	Target net.IP // IPv6 address being claimed
	Name   string // Name of the claimant
	Nonce  string // Nonce used to solve the challenge
}

// Hash computes the SHA-256 hash of the proof of work data
func (pow *ProofOfWork) Hash() [32]byte {
	// Create input data: target_ip + claimant + nonce
	data := make([]byte, 0, 16+len(pow.Name)+8)

	// Add IPv6 address (16 bytes)
	data = append(data, pow.Target.To16()...)

	// Add claimant name
	data = append(data, []byte(pow.Name)...)

	// Add nonce as string bytes
	data = append(data, []byte(pow.Nonce)...)

	return sha256.Sum256(data)
}

// IsValid checks if the proof of work satisfies the difficulty requirement
func (pow *ProofOfWork) IsValid(difficulty uint8) bool {
	hash := pow.Hash()

	// Count leading zero bits
	leadingZeros := 0
	for i := range 32 {
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
		if leadingZeros >= int(difficulty) {
			return true
		}
	}

	return leadingZeros >= int(difficulty)
}

// SolveProofOfWork attempts to solve a proof of work challenge (for client use)
func SolveProofOfWork(target net.IP, claimant string, difficulty uint8, maxAttempts uint64) (*ProofOfWork, error) {
	pow := &ProofOfWork{
		Target: target,
		Name:   claimant,
	}

	for nonce := range maxAttempts {
		pow.Nonce = fmt.Sprintf("%d", nonce)
		if pow.IsValid(difficulty) {
			return pow, nil
		}
	}

	return nil, fmt.Errorf("could not solve proof of work within %d attempts", maxAttempts)
}
