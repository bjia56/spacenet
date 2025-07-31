package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	_ "embed"
)

//go:embed ipv6names.json
var ipv6NamesData []byte

type IPv6Names struct {
	Adjectives     map[string][]string `json:"adjectives"`
	Nouns          map[string][]string `json:"nouns"`
	CelestialTypes map[string][]string `json:"celestialTypes"`
	SubnetMappings map[string]int      `json:"subnetMappings"`
	LevelNames     []string            `json:"levelNames"`
}

var (
	adjectives     map[int][]string
	nouns          map[int][]string
	celestialTypes map[int][]string
)

func init() {
	var namesData IPv6Names
	if err := json.Unmarshal(ipv6NamesData, &namesData); err != nil {
		panic(fmt.Sprintf("Failed to parse embedded IPv6 names data: %v", err))
	}

	adjectives = make(map[int][]string)
	nouns = make(map[int][]string)
	celestialTypes = make(map[int][]string)

	for strKey, value := range namesData.Adjectives {
		var key int
		fmt.Sscanf(strKey, "%d", &key)
		adjectives[key] = value
	}

	for strKey, value := range namesData.Nouns {
		var key int
		fmt.Sscanf(strKey, "%d", &key)
		nouns[key] = value
	}

	for strKey, value := range namesData.CelestialTypes {
		var key int
		fmt.Sscanf(strKey, "%d", &key)
		celestialTypes[key] = value
	}
}

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
