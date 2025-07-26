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
	adjectives = []string{
		"Absolute", "Abundant", "Adamant", "Advancing", "Aegis", "Aeonic", "Ageless", "Ancient",
		"Arcane", "Ascending", "Astral", "Azure", "Binding", "Blazing", "Boundless", "Bright",
		"Brilliant", "Celestial", "Cerulean", "Chaotic", "Chromatic", "Colossal", "Cosmic",
		"Crimson", "Crystal", "Dark", "Desolate", "Divine", "Dormant", "Drifting", "Dynamic",
		"Ebony", "Echoing", "Emerald", "Endless", "Eternal", "Ethereal", "Everborn", "Fading",
		"Fierce", "Flaring", "Flowing", "Frozen", "Fundamental", "Furious", "Fusion", "Gleaming",
		"Glowing", "Golden", "Grand", "Gravity", "Hidden", "Howling", "Hyperion", "Imperial",
		"Infinite", "Interstellar", "Iridescent", "Jovian", "Kinetic", "Lasting", "Limitless",
		"Lunar", "Luminous", "Majestic", "Mysterious", "Mystic", "Nebulous", "Neural", "Noble",
		"Northern", "Nova", "Obsidian", "Orbital", "Paramount", "Perpetual", "Phantom", "Plasma",
		"Polar", "Prismatic", "Pulsing", "Quantum", "Quartz", "Radiant", "Raging", "Remote",
		"Rising", "Royal", "Ruby", "Sacred", "Sapphire", "Savage", "Scarlet", "Shattered",
		"Shining", "Silent", "Silver", "Singular", "Solar", "Sovereign", "Spatial", "Spectral",
		"Stellar", "Storm", "Sublime", "Temporal", "Titan", "Twilight", "Ultimate", "Umbral",
		"Unified", "Universal", "Untamed", "Vanguard", "Vast", "Veiled", "Velvet", "Venerable",
		"Verdant", "Vibrant", "Violet", "Volatile", "Wandering", "Wayward", "Western", "Whispering",
		"Windswept", "Xenial", "Yielding", "Zenith",
	}

	nouns = []string{
		"Aegis", "Altar", "Apex", "Archive", "Arsenal", "Artifact", "Ascent", "Atlas",
		"Aurora", "Avalon", "Axiom", "Bastion", "Beacon", "Breach", "Bulwark", "Cascade",
		"Catalyst", "Celestial", "Citadel", "Codex", "Colossus", "Core", "Corona", "Cosmos",
		"Crown", "Crucible", "Crystal", "Cyclix", "Dawn", "Decree", "Dominion", "Drift",
		"Echo", "Edge", "Elysium", "Ember", "Enigma", "Epoch", "Equinox", "Eternity",
		"Exodus", "Expanse", "Eye", "Fissure", "Flame", "Flow", "Flux", "Forge",
		"Formation", "Fortress", "Frontier", "Gate", "Gateway", "Genesis", "Grove", "Guard",
		"Haven", "Heart", "Helix", "Horizon", "Hymn", "Infinity", "Iris", "Junction",
		"Keep", "Keystone", "Labyrinth", "Legacy", "Lens", "Light", "Locus", "Mandate",
		"Matrix", "Meridian", "Monolith", "Monument", "Nebula", "Nexus", "Nova", "Oasis",
		"Odyssey", "Oracle", "Orbit", "Paragon", "Path", "Pillar", "Portal", "Prism",
		"Prophecy", "Pulse", "Pyre", "Quasar", "Rampart", "Realm", "Relic", "Rift",
		"Sanctum", "Sanctuary", "Sentinel", "Shard", "Sigil", "Singularity", "Solace", "Spire",
		"Star", "Summit", "Synergy", "Terminus", "Throne", "Tower", "Trinity", "Unity",
		"Vale", "Vault", "Vector", "Vertex", "Vigil", "Void", "Vortex", "Ward",
		"Watch", "Web", "Wellspring", "Witness", "Wonder", "Yard", "Zenith", "Zone",
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
			"Haven", "Stronghold", "Citadel", "Nexus",
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
	adjIndex := int(binary.BigEndian.Uint32(hash[0:4])) % len(adjectives)
	nounIndex := int(binary.BigEndian.Uint32(hash[4:8])) % len(nouns)
	celestialIndex := int(binary.BigEndian.Uint32(hash[8:12])) % len(celestialTypes[subnetSize])

	// Generate a numeric suffix using another part of the hash
	suffix := binary.BigEndian.Uint16(hash[12:14]) % 1000

	return fmt.Sprintf("%s %s %s-%d",
		adjectives[adjIndex],
		nouns[nounIndex],
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
