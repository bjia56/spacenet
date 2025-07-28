package server

import "sync"

// MockStore implements the Store interface for testing
type MockStore struct {
	mu            sync.RWMutex
	claims        map[string]string
	subnets       map[string]*SubnetStats
	processClaimFunc func(ipAddr string, claimant string) error
	getClaimFunc     func(ipAddr string) (string, bool)
	getAllClaimsFunc func() map[string]string
	getSubnetStatsFunc func(subnet string) (*SubnetStats, bool)
	closeFunc        func() error
	callLog          []string // Track method calls for verification
}

// NewMockStore creates a new mock store for testing
func NewMockStore() *MockStore {
	return &MockStore{
		claims:  make(map[string]string),
		subnets: make(map[string]*SubnetStats),
		callLog: make([]string, 0),
	}
}

// ProcessClaim implements Store interface
func (m *MockStore) ProcessClaim(ipAddr string, claimant string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.callLog = append(m.callLog, "ProcessClaim:"+ipAddr+":"+claimant)
	
	if m.processClaimFunc != nil {
		return m.processClaimFunc(ipAddr, claimant)
	}
	
	m.claims[ipAddr] = claimant
	return nil
}

// GetClaim implements Store interface
func (m *MockStore) GetClaim(ipAddr string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.callLog = append(m.callLog, "GetClaim:"+ipAddr)
	
	if m.getClaimFunc != nil {
		return m.getClaimFunc(ipAddr)
	}
	
	claimant, exists := m.claims[ipAddr]
	return claimant, exists
}

// GetAllClaims implements Store interface
func (m *MockStore) GetAllClaims() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.callLog = append(m.callLog, "GetAllClaims")
	
	if m.getAllClaimsFunc != nil {
		return m.getAllClaimsFunc()
	}
	
	result := make(map[string]string)
	for k, v := range m.claims {
		result[k] = v
	}
	return result
}

// GetSubnetStats implements Store interface
func (m *MockStore) GetSubnetStats(subnet string) (*SubnetStats, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.callLog = append(m.callLog, "GetSubnetStats:"+subnet)
	
	if m.getSubnetStatsFunc != nil {
		return m.getSubnetStatsFunc(subnet)
	}
	
	stats, exists := m.subnets[subnet]
	return stats, exists
}

// Close implements Store interface
func (m *MockStore) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.callLog = append(m.callLog, "Close")
	
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	
	return nil
}

// Test helper methods

// SetProcessClaimFunc allows customizing ProcessClaim behavior
func (m *MockStore) SetProcessClaimFunc(f func(ipAddr string, claimant string) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processClaimFunc = f
}

// SetGetClaimFunc allows customizing GetClaim behavior
func (m *MockStore) SetGetClaimFunc(f func(ipAddr string) (string, bool)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getClaimFunc = f
}

// SetGetAllClaimsFunc allows customizing GetAllClaims behavior
func (m *MockStore) SetGetAllClaimsFunc(f func() map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getAllClaimsFunc = f
}

// SetGetSubnetStatsFunc allows customizing GetSubnetStats behavior
func (m *MockStore) SetGetSubnetStatsFunc(f func(subnet string) (*SubnetStats, bool)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getSubnetStatsFunc = f
}

// SetCloseFunc allows customizing Close behavior
func (m *MockStore) SetCloseFunc(f func() error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeFunc = f
}

// GetCallLog returns the log of method calls for verification
func (m *MockStore) GetCallLog() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make([]string, len(m.callLog))
	copy(result, m.callLog)
	return result
}

// ClearCallLog clears the call log
func (m *MockStore) ClearCallLog() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callLog = m.callLog[:0]
}

// AddClaim directly adds a claim to the mock store (for test setup)
func (m *MockStore) AddClaim(ipAddr, claimant string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.claims[ipAddr] = claimant
}

// AddSubnetStats directly adds subnet stats to the mock store (for test setup)
func (m *MockStore) AddSubnetStats(subnet string, stats *SubnetStats) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subnets[subnet] = stats
}