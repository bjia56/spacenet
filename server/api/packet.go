// Package api defines shared packet structures between the SpaceNet server and client
package api

import (
	"encoding/binary"
	"fmt"
	"net"
)

// ClaimPacket represents the structure of a claim UDP packet
type ClaimPacket struct {
	Difficulty uint8   // 1 byte: Required difficulty level
	Nonce      uint64  // 8 bytes: Proof of work nonce
	Claimant   string  // Remaining bytes: Claimant name (max 23 bytes)
}

const (
	MaxPacketSize    = 32  // Maximum UDP packet size
	difficultySize   = 1   // Size of difficulty field
	nonceSize        = 8   // Size of nonce field
	MaxClaimantSize  = MaxPacketSize - difficultySize - nonceSize // 23 bytes
)

// ParseClaimPacket parses a UDP packet into a ClaimPacket structure
func ParseClaimPacket(data []byte) (*ClaimPacket, error) {
	if len(data) < difficultySize+nonceSize {
		return nil, fmt.Errorf("packet too small: need at least %d bytes, got %d", difficultySize+nonceSize, len(data))
	}
	
	if len(data) > MaxPacketSize {
		return nil, fmt.Errorf("packet too large: max %d bytes, got %d", MaxPacketSize, len(data))
	}
	
	packet := &ClaimPacket{}
	
	// Parse difficulty (1 byte)
	packet.Difficulty = data[0]
	
	// Parse nonce (8 bytes, big endian)
	packet.Nonce = binary.BigEndian.Uint64(data[1:9])
	
	// Parse claimant name (remaining bytes)
	if len(data) > 9 {
		claimantBytes := data[9:]
		// Remove any null bytes at the end
		for i := len(claimantBytes) - 1; i >= 0; i-- {
			if claimantBytes[i] != 0 {
				claimantBytes = claimantBytes[:i+1]
				break
			}
		}
		packet.Claimant = string(claimantBytes)
	}
	
	// Validate claimant name
	if len(packet.Claimant) == 0 {
		return nil, fmt.Errorf("claimant name cannot be empty")
	}
	
	if len(packet.Claimant) > MaxClaimantSize {
		return nil, fmt.Errorf("claimant name too long: max %d bytes, got %d", MaxClaimantSize, len(packet.Claimant))
	}
	
	return packet, nil
}

// Serialize converts a ClaimPacket into bytes for transmission
func (cp *ClaimPacket) Serialize() ([]byte, error) {
	claimantBytes := []byte(cp.Claimant)
	
	if len(claimantBytes) > MaxClaimantSize {
		return nil, fmt.Errorf("claimant name too long: max %d bytes, got %d", MaxClaimantSize, len(claimantBytes))
	}
	
	data := make([]byte, difficultySize+nonceSize+len(claimantBytes))
	
	// Set difficulty (1 byte)
	data[0] = cp.Difficulty
	
	// Set nonce (8 bytes, big endian)
	binary.BigEndian.PutUint64(data[1:9], cp.Nonce)
	
	// Set claimant name
	copy(data[9:], claimantBytes)
	
	return data, nil
}

// CreateProofOfWork creates a ProofOfWork struct from the packet and target IP
func (cp *ClaimPacket) CreateProofOfWork(targetIP net.IP) *ProofOfWork {
	return &ProofOfWork{
		Target:     targetIP,
		Claimant:   cp.Claimant,
		Nonce:      cp.Nonce,
		Difficulty: cp.Difficulty,
	}
}

// IsLegacyPacket checks if the packet data is in the old format (plain text claimant name)
func IsLegacyPacket(data []byte) bool {
	// Legacy packets are plain text, so they shouldn't start with a high difficulty value
	// New packets will have difficulty in range 8-28 typically
	if len(data) == 0 {
		return false
	}
	
	// If first byte is a reasonable difficulty value and packet is at least 9 bytes,
	// it's likely a new packet format
	if data[0] >= 8 && data[0] <= 28 && len(data) >= 9 {
		return false
	}
	
	// Otherwise, treat as legacy
	return true
}