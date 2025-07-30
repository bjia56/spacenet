package server

import (
	"net"
	"testing"

	"github.com/bjia56/spacenet/server/api"
)

func TestProofOfWork_IsValid(t *testing.T) {
	// Test case: known valid proof of work
	pow := &api.ProofOfWork{
		Target:   net.ParseIP("2001:db8::1"),
		Claimant: "alice",
		Nonce:    12345,
	}
	difficulty := uint8(8) // Example difficulty

	// Find a valid nonce for this difficulty
	validPow, err := api.SolveProofOfWork(pow.Target, pow.Claimant, difficulty, 1000000)
	if err != nil {
		t.Fatalf("Failed to solve proof of work: %v", err)
	}

	if !validPow.IsValid(difficulty) {
		t.Error("Solved proof of work should be valid")
	}

	// Test invalid proof of work (wrong nonce)
	invalidPow := &api.ProofOfWork{
		Target:   validPow.Target,
		Claimant: validPow.Claimant,
		Nonce:    validPow.Nonce + 1, // Wrong nonce
	}

	if invalidPow.IsValid(difficulty) {
		t.Error("Invalid proof of work should not be valid")
	}
}

func TestCalculateDifficulty(t *testing.T) {
	store := NewClaimStore()

	// Test unclaimed address
	difficulty := store.CalculateDifficulty("2001:db8::1")
	expected := uint8(8) // Base difficulty
	if difficulty != expected {
		t.Errorf("Expected difficulty %d for unclaimed address, got %d", expected, difficulty)
	}

	// Claim the address
	store.ProcessClaim("2001:db8::1", "alice")

	// Test claimed address
	difficulty = store.CalculateDifficulty("2001:db8::1")
	expected = uint8(12) // Base (8) + claim bonus (4)
	if difficulty != expected {
		t.Errorf("Expected difficulty %d for claimed address, got %d", expected, difficulty)
	}

	// Claim contiguous addresses in the same /124 block
	// 2001:db8::1 is in the /124 block 2001:db8::/124
	// Let's claim a few more addresses in this block
	store.ProcessClaim("2001:db8::2", "alice")
	store.ProcessClaim("2001:db8::3", "alice")

	// Check difficulty for another address in the same block
	difficulty = store.CalculateDifficulty("2001:db8::4")
	expected = uint8(8) // Base difficulty (not claimed yet)
	if difficulty != expected {
		t.Errorf("Expected difficulty %d for unclaimed address in block with contiguous claims, got %d", expected, difficulty)
	}

	// Check difficulty for claiming an address owned by someone with contiguous claims
	difficulty = store.CalculateDifficulty("2001:db8::1")
	// Base (8) + claim bonus (4) + contiguous bonus (2 * 2 contiguous addresses)
	expected = uint8(16)
	if difficulty != expected {
		t.Errorf("Expected difficulty %d for address with contiguous claims, got %d", expected, difficulty)
	}
}

func TestClaimPacket_SerializeAndParse(t *testing.T) {
	original := &api.ClaimPacket{
		Nonce:    123456789,
		Claimant: "alice",
	}

	// Serialize
	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize packet: %v", err)
	}

	// Parse back
	parsed, err := api.ParseClaimPacket(data)
	if err != nil {
		t.Fatalf("Failed to parse packet: %v", err)
	}

	// Verify fields
	if parsed.Nonce != original.Nonce {
		t.Errorf("Nonce mismatch: expected %d, got %d", original.Nonce, parsed.Nonce)
	}

	if parsed.Claimant != original.Claimant {
		t.Errorf("Claimant mismatch: expected %s, got %s", original.Claimant, parsed.Claimant)
	}
}

func TestValidateProofOfWork(t *testing.T) {
	store := NewClaimStore()
	target := net.ParseIP("2001:db8::1")

	// Create a valid proof of work
	requiredDifficulty := store.CalculateDifficulty(target.String())
	validPow, err := api.SolveProofOfWork(target, "alice", requiredDifficulty, 1000000)
	if err != nil {
		t.Fatalf("Failed to solve proof of work: %v", err)
	}

	// Should be valid
	if err := store.ValidateProofOfWork(validPow); err != nil {
		t.Errorf("Valid proof of work should pass validation: %v", err)
	}

	// Test with insufficient difficulty
	invalidPow, err := api.SolveProofOfWork(target, "alice", requiredDifficulty-1, 1000000)
	if err != nil {
		t.Fatalf("Failed to solve proof of work: %v", err)
	}

	if err := store.ValidateProofOfWork(invalidPow); err == nil {
		t.Error("Proof of work with insufficient difficulty should fail validation")
	}
}
