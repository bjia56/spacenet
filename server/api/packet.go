// Package api defines shared packet structures between the SpaceNet server and client
package api

import (
	"encoding/binary"
	"fmt"
	"net"
)

// ClaimPacket represents the structure of a claim UDP packet
type ClaimPacket struct {
	Nonce    uint64 // 8 bytes: Proof of work nonce
	Claimant string // Remaining bytes: Claimant name (max 24 bytes)
}

const (
	MaxPacketSize   = 32                        // Maximum UDP packet size
	nonceSize       = 8                         // Size of nonce field
	MaxClaimantSize = MaxPacketSize - nonceSize // 24 bytes
)

type ErrPacketTooSmall string
type ErrPacketTooLarge string
type ErrInvalidClaimant string

const (
	errPacketTooSmall  ErrPacketTooSmall  = "packet too small"
	errPacketTooLarge  ErrPacketTooLarge  = "packet too large"
	errInvalidClaimant ErrInvalidClaimant = "invalid claimant"
)

func (e ErrPacketTooSmall) Error() string {
	return string(e)
}

func (e ErrPacketTooLarge) Error() string {
	return string(e)
}

func (e ErrInvalidClaimant) Error() string {
	return string(e)
}

// ParseClaimPacket parses a UDP packet into a ClaimPacket structure
func ParseClaimPacket(data []byte) (*ClaimPacket, error) {
	if len(data) < nonceSize {
		return nil, errPacketTooSmall
	}

	if len(data) > MaxPacketSize {
		return nil, errPacketTooLarge
	}

	packet := &ClaimPacket{}

	// Parse nonce (8 bytes, big endian)
	packet.Nonce = binary.BigEndian.Uint64(data[0:8])

	// Parse claimant name (remaining bytes)
	if len(data) > 8 {
		claimantBytes := data[8:]
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
	if len(packet.Claimant) == 0 || len(packet.Claimant) > MaxClaimantSize {
		return nil, errInvalidClaimant
	}

	return packet, nil
}

// Serialize converts a ClaimPacket into bytes for transmission
func (cp *ClaimPacket) Serialize() ([]byte, error) {
	claimantBytes := []byte(cp.Claimant)

	if len(claimantBytes) > MaxClaimantSize {
		return nil, fmt.Errorf("claimant name too long: max %d bytes, got %d", MaxClaimantSize, len(claimantBytes))
	}

	data := make([]byte, nonceSize+len(claimantBytes))

	// Set nonce (8 bytes, big endian)
	binary.BigEndian.PutUint64(data[0:8], cp.Nonce)

	// Set claimant name
	copy(data[8:], claimantBytes)

	return data, nil
}

// CreateProofOfWork creates a ProofOfWork struct from the packet and target IP
func (cp *ClaimPacket) CreateProofOfWork(targetIP net.IP) *ProofOfWork {
	return &ProofOfWork{
		Target:   targetIP,
		Claimant: cp.Claimant,
		Nonce:    cp.Nonce,
	}
}
