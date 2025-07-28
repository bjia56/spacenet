package server

import (
	"github.com/bjia56/spacenet/server/api"
)

// Type aliases for compatibility
type ClaimPacket = api.ClaimPacket
type ProofOfWork = api.ProofOfWork

// Function aliases for compatibility
var (
	ParseClaimPacket = api.ParseClaimPacket
	IsLegacyPacket   = api.IsLegacyPacket
)