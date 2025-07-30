import { createHash } from 'crypto';

// Predefined word lists for name generation (ported from Go TUI)
const adjectives: Record<number, string[]> = {
  // Level 16 - Universe-scale adjectives
  16: [
    "Absolute", "Boundless", "Colossal", "Cosmic", "Eternal", "Fundamental",
    "Infinite", "Interstellar", "Limitless", "Paramount", "Primordial", "Universal",
    "Ultimate", "Vast", "Celestial", "Omnipotent", "Singular", "Supreme",
  ],

  // Level 32 - Supercluster adjectives
  32: [
    "Abundant", "Ascending", "Binding", "Dynamic", "Endless", "Fundamental",
    "Grand", "Hyperion", "Imperial", "Majestic", "Sovereign", "Stellar",
    "Sublime", "Unified", "Dominant", "Mammoth", "Massive",
  ],

  // Level 48 - Galaxy Group adjectives
  48: [
    "Abundant", "Clustered", "Collective", "Connected", "Flowing", "Gathering",
    "Harmonious", "Linked", "Networked", "Streaming", "Unified", "United",
    "Woven", "Bound", "Converging", "Joined",
  ],

  // Level 64 - Galaxy adjectives
  64: [
    "Astral", "Blazing", "Brilliant", "Luminous", "Nebulous", "Radiant",
    "Shining", "Spiraling", "Stellar", "Swirling", "Whirling", "Rotating",
    "Galactic", "Cosmic", "Starborn", "Celestial",
  ],

  // Level 80 - Star Group adjectives
  80: [
    "Burning", "Flaring", "Gleaming", "Glowing", "Golden", "Illuminated",
    "Lucent", "Pulsing", "Radiating", "Scintillating", "Twinkling", "Bright",
    "Dazzling", "Effulgent", "Starlit",
  ],

  // Level 96 - Solar System adjectives
  96: [
    "Balanced", "Circular", "Gravitational", "Harmonious", "Orbital", "Planetary",
    "Revolving", "Solar", "Synchronous", "Systematic", "Aligned", "Cyclic",
    "Ecliptic", "Ordered",
  ],

  // Level 112 - Planet adjectives
  112: [
    "Azure", "Crystalline", "Emerald", "Frozen", "Gaseous", "Molten",
    "Obsidian", "Rocky", "Sapphire", "Terrestrial", "Tropical", "Verdant",
    "Violet", "Volcanic", "Windswept", "Crimson",
  ],

  // Level 128 - Settlement adjectives
  128: [
    "Ancient", "Bustling", "Colonial", "Fortified", "Hidden", "Industrial",
    "Metropolitan", "Noble", "Prosperous", "Rising", "Sacred", "Thriving",
    "Urban", "Vibrant", "Wealthy", "Established",
  ],
};

const nouns: Record<number, string[]> = {
  // Level 16 - Universe-scale nouns
  16: [
    "Axis", "Boundary", "Expanse", "Firmament", "Infinity", "Matrix",
    "Membrane", "Nexus", "Singularity", "Terminus", "Vector", "Vertex",
    "Void", "Web", "Fabric", "Lattice", "Framework",
  ],

  // Level 32 - Supercluster nouns
  32: [
    "Amalgam", "Bastion", "Colossus", "Domain", "Empire", "Formation",
    "Legion", "Mandate", "Realm", "Sphere", "Supremacy", "Unity",
    "Dominion", "Coalition", "Assembly",
  ],

  // Level 48 - Galaxy Group nouns
  48: [
    "Assembly", "Chain", "Confluence", "Fellowship", "Network", "Pattern",
    "Sequence", "Stream", "Symphony", "Union", "Weave", "Collection",
    "Gathering", "Association",
  ],

  // Level 64 - Galaxy nouns
  64: [
    "Corona", "Cosmos", "Disk", "Eye", "Helix", "Nebula",
    "Spiral", "Star", "Vortex", "Whirlpool", "Ring", "Cloud",
    "Field", "Sea", "Cluster",
  ],

  // Level 80 - Star Group nouns
  80: [
    "Beacon", "Constellation", "Crucible", "Ember", "Flame", "Forge",
    "Light", "Pyre", "Spark", "Torch", "Aurora", "Flare",
    "Stream", "Garden",
  ],

  // Level 96 - Solar System nouns
  96: [
    "Circuit", "Cycle", "Horizon", "Orbit", "Path", "Procession",
    "Ring", "Sanctuary", "Sphere", "System", "Dance", "Family",
    "Haven", "Domain",
  ],

  // Level 112 - Planet nouns
  112: [
    "Globe", "Haven", "Heart", "Keep", "Oasis", "Paradise",
    "Sanctuary", "Sphere", "Stronghold", "Vale", "World", "Garden",
    "Realm", "Crown", "Jewel",
  ],

  // Level 128 - Settlement nouns
  128: [
    "Arcology", "Citadel", "Enclave", "Fortress", "Haven", "Nexus",
    "Outpost", "Sanctuary", "Spire", "Stronghold", "Tower", "Ward",
    "Capital", "Port", "Station", "Hub",
  ],
};

const celestialTypes: Record<number, string[]> = {
  // Largest scale structures - massive cosmic boundaries and constructs
  16: [
    "Superstructure", "Cosmic Wall", "Great Wall", "Filament",
    "Cosmic Web", "Void Wall", "Megastructure", "Cosmic Ring",
    "Barrier", "Cosmic Membrane", "Universal Divide",
  ],

  // Major galaxy collection scales
  32: [
    "Supercluster", "Galaxy Shell", "Cosmic Shell", "Massive Cluster",
    "Meta Cluster", "Celestial Complex", "Cosmic Cloud", "Stellar Sea",
  ],

  // Medium-large galaxy groupings
  48: [
    "Galaxy Group", "Star Cluster", "Stellar Association",
    "Galactic Cloud", "Cosmic Stream", "Celestial Chain",
    "Stellar Complex", "Galactic Gathering",
  ],

  // Individual large stellar collections
  64: [
    "Galaxy", "Nebula", "Star Sea", "Stellar Spiral",
    "Cosmic Disk", "Star Cloud", "Stellar Field", "Galactic Ring",
  ],

  // Localized star groupings
  80: [
    "Star Group", "Stellar Cluster", "Star Stream", "Star Field",
    "Cosmic Oasis", "Stellar Neighborhood", "Star Colony", "Stellar Circuit",
  ],

  // Individual star systems
  96: [
    "Solar System", "Star System", "Planetary System", "Stellar Domain",
    "Cosmic Haven", "Star Domain", "Stellar Sanctuary", "Orbital Realm",
  ],

  // Major celestial bodies
  112: [
    "Planet", "Celestial Body", "World", "Planetoid",
    "Moon", "Satellite", "Giant Moon", "Megamoon",
    "Dwarf Planet", "Minor Planet", "Ice Giant", "Gas Giant",
  ],

  // Inhabited locations
  128: [
    "Metropolis", "City", "Megacity", "Habitat",
    "Colony", "Settlement", "Outpost", "Station",
    "Enclave", "Base", "Community", "Urban Center",
    "Village", "Town", "Borough", "District",
  ],
};

// Subnet size mappings
export const subnetMappings: Record<number, number> = {
  0: 16,
  1: 32,
  2: 48,
  3: 64,
  4: 80,
  5: 96,
  6: 112,
  7: 128,
};

export const levelNames = [
  "Great Wall",
  "Supercluster", 
  "Galaxy Group",
  "Galaxy",
  "Star Cluster",
  "Solar System",
  "Planet",
  "City"
];

// Helper function to parse IPv6 address and create truncated version
function truncateIPv6(addr: string, bits: number): Buffer {
  // Parse IPv6 address to bytes
  const parts = addr.split(':');
  const bytes = Buffer.alloc(16);
  
  // Handle compressed notation
  let expandedParts: string[] = [];
  const compressedIndex = parts.indexOf('');
  
  if (compressedIndex !== -1) {
    const beforeCompressed = parts.slice(0, compressedIndex);
    const afterCompressed = parts.slice(compressedIndex + 1).filter(p => p !== '');
    const missingParts = 8 - beforeCompressed.length - afterCompressed.length;
    
    expandedParts = [
      ...beforeCompressed,
      ...Array(missingParts).fill('0000'),
      ...afterCompressed
    ];
  } else {
    expandedParts = parts;
  }
  
  // Fill in any missing parts with zeros
  while (expandedParts.length < 8) {
    expandedParts.push('0000');
  }
  
  // Convert to bytes
  for (let i = 0; i < 8; i++) {
    const part = parseInt(expandedParts[i] || '0000', 16);
    bytes.writeUInt16BE(part, i * 2);
  }
  
  // Apply subnet mask
  const maskBytes = Math.floor(bits / 8);
  const maskBits = bits % 8;
  
  // Zero out bytes beyond the mask
  for (let i = maskBytes + (maskBits > 0 ? 1 : 0); i < 16; i++) {
    bytes[i] = 0;
  }
  
  // Apply bit mask to the last partial byte
  if (maskBits > 0) {
    const mask = (0xFF << (8 - maskBits)) & 0xFF;
    bytes[maskBytes] &= mask;
  }
  
  return bytes;
}

// Generate a unique name for an IPv6 address at a given subnet size
export function generateName(addr: string, subnetSize: number): string {
  if (!celestialTypes[subnetSize]) {
    throw new Error(`Invalid subnet size: ${subnetSize}`);
  }
  
  // Create a hash of the truncated address based on subnet size
  const truncatedAddr = truncateIPv6(addr, subnetSize);
  const subnetSizeBuffer = Buffer.alloc(4);
  subnetSizeBuffer.writeUInt32BE(subnetSize, 0);
  
  const hash = createHash('sha256')
    .update(truncatedAddr)
    .update(subnetSizeBuffer)
    .digest();
  
  // Use different parts of the hash for different components of the name
  const adjIndex = hash.readUInt32BE(0) % adjectives[subnetSize].length;
  const nounIndex = hash.readUInt32BE(4) % nouns[subnetSize].length;
  const celestialIndex = hash.readUInt32BE(8) % celestialTypes[subnetSize].length;
  
  // Generate a numeric suffix using another part of the hash
  const suffix = hash.readUInt16BE(12) % 1000;
  
  return `${adjectives[subnetSize][adjIndex]} ${nouns[subnetSize][nounIndex]} ${celestialTypes[subnetSize][celestialIndex]}-${suffix}`;
}

// Generate the full hierarchy for an IPv6 address
export function getHierarchy(addr: string): string[] {
  const sizes = [16, 32, 48, 64, 80, 96, 112, 128];
  return sizes.map(size => generateName(addr, size));
}

// Create a full IPv6 address for a given index and level
export function makeIPv6Full(index: number, prefix: string, level: number): { addr: string; subnet: number } {
  const hex = index.toString(16).padStart(4, '0');
  const numSubBlocks = 8 - (level + 1);
  const zeroBlocks = ':0000'.repeat(numSubBlocks);
  
  const full = prefix ? `${prefix}${hex}${zeroBlocks}` : `${hex}${zeroBlocks}`;
  return {
    addr: full,
    subnet: subnetMappings[level]
  };
}

export type SubnetLevel = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7;

export interface SubnetInfo {
  name: string;
  addr: string;
  subnet: number;
  level: SubnetLevel;
  levelName: string;
}