package server

import (
	"math/big"
	"net"
	"slices"
	"sync"
)

// IPTree represents a hierarchical structure for managing IPv6 address claims
// It organizes claims by subnet hierarchy for efficient lookups
type IPTree struct {
	mu   sync.RWMutex
	root *IPNode
	// No longer stores its own claims map - uses external map
}

// IPNode represents a node in the IP tree
type IPNode struct {
	// Subnet represented by this node (in CIDR notation)
	subnet *net.IPNet

	// Prefix length of this subnet
	prefixLen int

	// Total number of addresses claimed in this subnet
	claimedCount *big.Int

	// Total possible addresses in this subnet
	totalAddresses *big.Int

	// Map of claimants to their claimed address count in this subnet
	claimants map[string]*big.Int

	// Dominant claimant in this subnet (with highest percentage)
	dominantClaimant string

	// Percentage of subnet owned by dominant claimant (0-100)
	dominantPercentage float64

	// Child nodes (more specific subnets)
	children map[string]*IPNode
}

// Import the shared SubnetStats type from the api package

// NewIPTree creates a new IP tree
func NewIPTree() *IPTree {
	// Create root node for the entire IPv6 space
	_, rootNet, _ := net.ParseCIDR("::/0")

	root := &IPNode{
		subnet:         rootNet,
		prefixLen:      0,
		claimedCount:   big.NewInt(0),
		totalAddresses: new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil), // 2^128
		claimants:      make(map[string]*big.Int),
		children:       make(map[string]*IPNode),
	}

	return &IPTree{
		root: root,
	}
}

// processClaim updates the tree with a new claim
func (t *IPTree) processClaim(ipAddr string, claimant string, oldClaimant string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Update the tree structure
	ip := net.ParseIP(ipAddr)
	if ip == nil || ip.To16() == nil {
		return // Invalid IP
	}

	// If this is replacing an existing claim, first remove the old one
	if oldClaimant != "" && oldClaimant != claimant {
		t.removeClaimLocked(ipAddr, oldClaimant)
	}

	// Update tree for standard subnet sizes
	t.updateSubnet(ip, 16, claimant)
	t.updateSubnet(ip, 32, claimant)
	t.updateSubnet(ip, 48, claimant)
	t.updateSubnet(ip, 64, claimant)
	t.updateSubnet(ip, 80, claimant)
	t.updateSubnet(ip, 96, claimant)
	t.updateSubnet(ip, 112, claimant)
	t.updateSubnet(ip, 128, claimant)
}

// updateSubnet updates a specific subnet node for an IP claim
func (t *IPTree) updateSubnet(ip net.IP, prefixLen int, claimant string) {
	// Create subnet mask for the given prefix length
	mask := net.CIDRMask(prefixLen, 128)
	subnet := &net.IPNet{
		IP:   ip.Mask(mask),
		Mask: mask,
	}

	// Find or create the node for this subnet
	node := t.findOrCreateNode(subnet, prefixLen)

	// Update node statistics
	claimantCount, exists := node.claimants[claimant]
	if !exists {
		claimantCount = big.NewInt(0)
		node.claimants[claimant] = claimantCount
	}

	// Increment the claimed count for this claimant
	claimantCount.Add(claimantCount, big.NewInt(1))

	// Increment total claimed count for this subnet
	node.claimedCount.Add(node.claimedCount, big.NewInt(1))

	// Recalculate dominant claimant
	t.recalculateDominant(node)
}

// findOrCreateNode finds or creates a node for the given subnet
func (t *IPTree) findOrCreateNode(subnet *net.IPNet, prefixLen int) *IPNode {
	// Start at the root
	node := t.root

	subnetStr := subnet.String()

	// Check if we already have a node for this subnet
	if child, exists := node.children[subnetStr]; exists {
		return child
	}

	// Calculate total addresses in this subnet
	totalAddrs := new(big.Int).Exp(big.NewInt(2), big.NewInt(128-int64(prefixLen)), nil)

	// Create a new node
	newNode := &IPNode{
		subnet:         subnet,
		prefixLen:      prefixLen,
		claimedCount:   big.NewInt(0),
		totalAddresses: totalAddrs,
		claimants:      make(map[string]*big.Int),
		children:       make(map[string]*IPNode),
	}

	// Add to children
	node.children[subnetStr] = newNode

	return newNode
}

// recalculateDominant recalculates the dominant claimant for a node
func (t *IPTree) recalculateDominant(node *IPNode) {
	var maxCount *big.Int
	var dominantClaimant string

	maxCount = big.NewInt(0)

	// Find claimant with highest count
	for claimant, count := range node.claimants {
		if count.Cmp(maxCount) > 0 {
			maxCount = count
			dominantClaimant = claimant
		} else if count.Cmp(maxCount) == 0 {
			// If there's a tie, prefer the lexicographically smaller claimant
			if dominantClaimant == "" || claimant < dominantClaimant {
				dominantClaimant = claimant
			}
		}
	}

	// Calculate percentage if we have claims
	var percentage float64 = 0
	if node.claimedCount.Cmp(big.NewInt(0)) > 0 {
		// Convert to float for percentage calculation
		countFloat := new(big.Float).SetInt(maxCount)
		totalFloat := new(big.Float).SetInt(node.totalAddresses)

		ratio, _ := new(big.Float).Quo(countFloat, totalFloat).Float64()
		percentage = ratio * 100.0
	}

	node.dominantClaimant = dominantClaimant
	node.dominantPercentage = percentage
}

// removeClaimLocked removes a claim from the tree (assumes lock is held)
func (t *IPTree) removeClaimLocked(ipAddr string, claimant string) {
	ip := net.ParseIP(ipAddr)
	if ip == nil || ip.To16() == nil {
		return // Invalid IP
	}

	// Update tree for standard subnet sizes
	t.removeFromSubnet(ip, 16, claimant)
	t.removeFromSubnet(ip, 32, claimant)
	t.removeFromSubnet(ip, 48, claimant)
	t.removeFromSubnet(ip, 64, claimant)
	t.removeFromSubnet(ip, 80, claimant)
	t.removeFromSubnet(ip, 96, claimant)
	t.removeFromSubnet(ip, 112, claimant)
	t.removeFromSubnet(ip, 128, claimant)
}

// removeFromSubnet removes a claim from a specific subnet
func (t *IPTree) removeFromSubnet(ip net.IP, prefixLen int, claimant string) {
	mask := net.CIDRMask(prefixLen, 128)
	subnet := &net.IPNet{
		IP:   ip.Mask(mask),
		Mask: mask,
	}

	subnetStr := subnet.String()

	// Find node
	node := t.root
	child, exists := node.children[subnetStr]
	if !exists {
		return // Node doesn't exist
	}

	// Update statistics
	claimantCount, exists := child.claimants[claimant]
	if exists {
		// Decrement count
		claimantCount.Sub(claimantCount, big.NewInt(1))

		// If count is zero, remove the claimant
		if claimantCount.Cmp(big.NewInt(0)) <= 0 {
			delete(child.claimants, claimant)
		}

		// Decrement total claimed count
		child.claimedCount.Sub(child.claimedCount, big.NewInt(1))

		// Recalculate dominant claimant
		t.recalculateDominant(child)
	}
}

// GetSubnetStats gets statistics for a subnet
func (t *IPTree) GetSubnetStats(subnetStr string) (*SubnetStats, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Parse subnet
	_, subnet, err := net.ParseCIDR(subnetStr)
	if err != nil {
		return nil, false
	}

	// Get prefix length
	prefixLen, _ := subnet.Mask.Size()

	// Round to nearest standard prefix
	stdPrefixes := []int{16, 32, 48, 64, 80, 96, 112, 128}
	exactMatch := false

	for _, stdPrefix := range stdPrefixes {
		if prefixLen == stdPrefix {
			exactMatch = true
			break
		}
	}

	if !exactMatch {
		// Find nearest standard prefix (round up)
		for _, stdPrefix := range stdPrefixes {
			if stdPrefix > prefixLen {
				prefixLen = stdPrefix
				break
			}
		}

		// If we couldn't find a larger prefix, use the largest
		if !exactMatch {
			prefixLen = 128
		}

		// Create new subnet with standard prefix
		subnet = &net.IPNet{
			IP:   subnet.IP.Mask(net.CIDRMask(prefixLen, 128)),
			Mask: net.CIDRMask(prefixLen, 128),
		}
	}

	subnetStr = subnet.String()

	// Find node
	node := t.root
	child, exists := node.children[subnetStr]
	if !exists {
		// No data for this subnet
		return &SubnetStats{
			Subnet:     subnetStr,
			Owner:      "",
			Percentage: 0,
		}, true
	}

	if child.dominantPercentage <= 50.0 {
		// If no dominant claimant, return empty stats
		return &SubnetStats{
			Subnet:     subnetStr,
			Owner:      "",
			Percentage: 0,
		}, true
	}

	return &SubnetStats{
		Subnet:     subnetStr,
		Owner:      child.dominantClaimant,
		Percentage: child.dominantPercentage,
	}, true
}

// GetAllSubnets gets statistics for all tracked subnets with the given prefix length
func (t *IPTree) GetAllSubnets(prefixLen int) []SubnetStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Validate prefix length
	validPrefix := false
	stdPrefixes := []int{16, 32, 48, 64, 80, 96, 112, 128}
	if slices.Contains(stdPrefixes, prefixLen) {
		validPrefix = true
	}

	if !validPrefix {
		return []SubnetStats{}
	}

	results := []SubnetStats{}

	// Find all subnets with this prefix
	for subnetStr, node := range t.root.children {
		if node.prefixLen == prefixLen {
			stats := SubnetStats{
				Subnet:     subnetStr,
				Owner:      node.dominantClaimant,
				Percentage: node.dominantPercentage,
			}
			results = append(results, stats)
		}
	}

	return results
}
