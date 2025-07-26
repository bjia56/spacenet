package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

// Predefined word lists for name generation
var (
	adjectives = map[int][]string{
		// Level 16 - Universe-scale adjectives
		16: {
			"Absolute", "Boundless", "Colossal", "Cosmic", "Eternal", "Fundamental",
			"Infinite", "Interstellar", "Limitless", "Paramount", "Primordial", "Universal",
			"Ultimate", "Vast", "Celestial", "Omnipotent", "Singular", "Supreme",
		},

		// Level 32 - Supercluster adjectives
		32: {
			"Abundant", "Ascending", "Binding", "Dynamic", "Endless", "Fundamental",
			"Grand", "Hyperion", "Imperial", "Majestic", "Sovereign", "Stellar",
			"Sublime", "Unified", "Dominant", "Mammoth", "Massive",
		},

		// Level 48 - Galaxy Group adjectives
		48: {
			"Abundant", "Clustered", "Collective", "Connected", "Flowing", "Gathering",
			"Harmonious", "Linked", "Networked", "Streaming", "Unified", "United",
			"Woven", "Bound", "Converging", "Joined",
		},

		// Level 64 - Galaxy adjectives
		64: {
			"Astral", "Blazing", "Brilliant", "Luminous", "Nebulous", "Radiant",
			"Shining", "Spiraling", "Stellar", "Swirling", "Whirling", "Rotating",
			"Galactic", "Cosmic", "Starborn", "Celestial",
		},

		// Level 80 - Star Group adjectives
		80: {
			"Burning", "Flaring", "Gleaming", "Glowing", "Golden", "Illuminated",
			"Lucent", "Pulsing", "Radiating", "Scintillating", "Twinkling", "Bright",
			"Dazzling", "Effulgent", "Starlit",
		},

		// Level 96 - Solar System adjectives
		96: {
			"Balanced", "Circular", "Gravitational", "Harmonious", "Orbital", "Planetary",
			"Revolving", "Solar", "Synchronous", "Systematic", "Aligned", "Cyclic",
			"Ecliptic", "Ordered",
		},

		// Level 112 - Planet adjectives
		112: {
			"Azure", "Crystalline", "Emerald", "Frozen", "Gaseous", "Molten",
			"Obsidian", "Rocky", "Sapphire", "Terrestrial", "Tropical", "Verdant",
			"Violet", "Volcanic", "Windswept", "Crimson",
		},

		// Level 128 - Settlement adjectives
		128: {
			"Ancient", "Bustling", "Colonial", "Fortified", "Hidden", "Industrial",
			"Metropolitan", "Noble", "Prosperous", "Rising", "Sacred", "Thriving",
			"Urban", "Vibrant", "Wealthy", "Established",
		},
	}

	nouns = map[int][]string{
		// Level 16 - Universe-scale nouns
		16: {
			"Axis", "Boundary", "Expanse", "Firmament", "Infinity", "Matrix",
			"Membrane", "Nexus", "Singularity", "Terminus", "Vector", "Vertex",
			"Void", "Web", "Fabric", "Lattice", "Framework",
		},

		// Level 32 - Supercluster nouns
		32: {
			"Amalgam", "Bastion", "Colossus", "Domain", "Empire", "Formation",
			"Legion", "Mandate", "Realm", "Sphere", "Supremacy", "Unity",
			"Dominion", "Coalition", "Assembly",
		},

		// Level 48 - Galaxy Group nouns
		48: {
			"Assembly", "Chain", "Confluence", "Fellowship", "Network", "Pattern",
			"Sequence", "Stream", "Symphony", "Union", "Weave", "Collection",
			"Gathering", "Association",
		},

		// Level 64 - Galaxy nouns
		64: {
			"Corona", "Cosmos", "Disk", "Eye", "Helix", "Nebula",
			"Spiral", "Star", "Vortex", "Whirlpool", "Ring", "Cloud",
			"Field", "Sea", "Cluster",
		},

		// Level 80 - Star Group nouns
		80: {
			"Beacon", "Constellation", "Crucible", "Ember", "Flame", "Forge",
			"Light", "Pyre", "Spark", "Torch", "Aurora", "Flare",
			"Stream", "Garden",
		},

		// Level 96 - Solar System nouns
		96: {
			"Circuit", "Cycle", "Horizon", "Orbit", "Path", "Procession",
			"Ring", "Sanctuary", "Sphere", "System", "Dance", "Family",
			"Haven", "Domain",
		},

		// Level 112 - Planet nouns
		112: {
			"Globe", "Haven", "Heart", "Keep", "Oasis", "Paradise",
			"Sanctuary", "Sphere", "Stronghold", "Vale", "World", "Garden",
			"Realm", "Crown", "Jewel",
		},

		// Level 128 - Settlement nouns
		128: {
			"Arcology", "Citadel", "Enclave", "Fortress", "Haven", "Nexus",
			"Outpost", "Sanctuary", "Spire", "Stronghold", "Tower", "Ward",
			"Capital", "Port", "Station", "Hub",
		},
	}

	celestialTypes = map[int][]string{
		// Largest scale structures - massive cosmic boundaries and constructs
		16: {
			"Superstructure", "Cosmic Wall", "Great Wall", "Filament",
			"Cosmic Web", "Void Wall", "Megastructure", "Cosmic Ring",
			"Barrier", "Cosmic Membrane", "Universal Divide",
		},

		// Major galaxy collection scales
		32: {
			"Supercluster", "Galaxy Shell", "Cosmic Shell", "Massive Cluster",
			"Meta Cluster", "Celestial Complex", "Cosmic Cloud", "Stellar Sea",
		},

		// Medium-large galaxy groupings
		48: {
			"Galaxy Group", "Star Cluster", "Stellar Association",
			"Galactic Cloud", "Cosmic Stream", "Celestial Chain",
			"Stellar Complex", "Galactic Gathering",
		},

		// Individual large stellar collections
		64: {
			"Galaxy", "Nebula", "Star Sea", "Stellar Spiral",
			"Cosmic Disk", "Star Cloud", "Stellar Field", "Galactic Ring",
		},

		// Localized star groupings
		80: {
			"Star Group", "Stellar Cluster", "Star Stream", "Star Field",
			"Cosmic Oasis", "Stellar Neighborhood", "Star Colony", "Stellar Circuit",
		},

		// Individual star systems
		96: {
			"Solar System", "Star System", "Planetary System", "Stellar Domain",
			"Cosmic Haven", "Star Domain", "Stellar Sanctuary", "Orbital Realm",
		},

		// Major celestial bodies
		112: {
			"Planet", "Celestial Body", "World", "Planetoid",
			"Moon", "Satellite", "Giant Moon", "Megamoon",
			"Dwarf Planet", "Minor Planet", "Ice Giant", "Gas Giant",
		},

		// Inhabited locations
		128: {
			"Metropolis", "City", "Megacity", "Habitat",
			"Colony", "Settlement", "Outpost", "Station",
			"Enclave", "Base", "Community", "Urban Center",
			"Village", "Town", "Borough", "District",
		},
	}
)

// GenerateName creates a unique name for an IPv6 address at a given subnet size
func GenerateName(saddr string, subnetSize int) (string, error) {
	addr := net.ParseIP(saddr)
	if addr == nil {
		return "", fmt.Errorf("invalid IPv6 address")
	}
	if addr.To16() == nil {
		return "", fmt.Errorf("invalid IPv6 address")
	}
	if _, ok := celestialTypes[subnetSize]; !ok {
		return "", fmt.Errorf("invalid subnet size: %d", subnetSize)
	}

	// Create a hash of the truncated address based on subnet size
	truncatedAddr := truncateIPv6(addr, subnetSize)
	subnetSizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(subnetSizeBytes, uint32(subnetSize))
	hash := sha256.Sum256(append(truncatedAddr, subnetSizeBytes...))

	// Use different parts of the hash for different components of the name
	adjIndex := int(binary.BigEndian.Uint32(hash[0:4])) % len(adjectives[subnetSize])
	nounIndex := int(binary.BigEndian.Uint32(hash[4:8])) % len(nouns[subnetSize])
	celestialIndex := int(binary.BigEndian.Uint32(hash[8:12])) % len(celestialTypes[subnetSize])

	// Generate a numeric suffix using another part of the hash
	suffix := binary.BigEndian.Uint16(hash[12:14]) % 1000

	return fmt.Sprintf("%s %s %s-%d",
		adjectives[subnetSize][adjIndex],
		nouns[subnetSize][nounIndex],
		celestialTypes[subnetSize][celestialIndex],
		suffix,
	), nil
}

// truncateIPv6 masks an IPv6 address to the specified subnet size
func truncateIPv6(addr net.IP, bits int) []byte {
	mask := net.CIDRMask(bits, 128)
	return addr.Mask(mask)
}

// GetHierarchy generates names for all parent categories of an IPv6 address
func GetHierarchy(addr net.IP) ([]string, error) {
	sizes := []int{16, 32, 48, 64, 80, 96, 112, 128}
	var names []string

	for _, size := range sizes {
		name, err := GenerateName(addr.String(), size)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	return names, nil
}

// FormatHierarchy returns a formatted string of the complete hierarchy
func FormatHierarchy(addr net.IP) (string, error) {
	names, err := GetHierarchy(addr)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	for i, name := range names {
		builder.WriteString(strings.Repeat("  ", i))
		builder.WriteString(name)
		builder.WriteString("\n")
	}

	return builder.String(), nil
}
