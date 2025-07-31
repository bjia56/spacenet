import { createHash } from 'crypto';
import ipv6NamesData from '../../../tui/ipv6names.json';

interface IPv6NamesData {
  adjectives: Record<string, string[]>;
  nouns: Record<string, string[]>;
  celestialTypes: Record<string, string[]>;
  subnetMappings: Record<string, number>;
  levelNames: string[];
}

const namesData = ipv6NamesData as IPv6NamesData;

const adjectives: Record<number, string[]> = {};
const nouns: Record<number, string[]> = {};
const celestialTypes: Record<number, string[]> = {};

for (const [strKey, value] of Object.entries(namesData.adjectives)) {
  adjectives[parseInt(strKey)] = value;
}

for (const [strKey, value] of Object.entries(namesData.nouns)) {
  nouns[parseInt(strKey)] = value;
}

for (const [strKey, value] of Object.entries(namesData.celestialTypes)) {
  celestialTypes[parseInt(strKey)] = value;
}

// Subnet size mappings from JSON data
export const subnetMappings: Record<number, number> = {};
for (const [strKey, value] of Object.entries(namesData.subnetMappings)) {
  subnetMappings[parseInt(strKey)] = value;
}

export const levelNames = namesData.levelNames;

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