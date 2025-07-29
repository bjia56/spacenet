package server

import (
	"context"
	"sync"
	"time"
)

// MockRedisClient is a mock implementation of RedisClient for testing
type MockRedisClient struct {
	mu sync.RWMutex
	
	// Data storage
	data map[string]string
	
	// Error injection for testing error scenarios
	pingError error
	getError  error
	setError  error
	scanError error
	
	// Call tracking for test verification
	callLog []string
	
	// Scan behavior customization
	scanKeys   []string
	scanCursor uint64
}

// NewMockRedisClient creates a new mock Redis client
func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data:    make(map[string]string),
		callLog: make([]string, 0),
	}
}

// SetPingError sets an error to be returned by Ping
func (m *MockRedisClient) SetPingError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pingError = err
}

// SetGetError sets an error to be returned by Get
func (m *MockRedisClient) SetGetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getError = err
}

// SetSetError sets an error to be returned by Set
func (m *MockRedisClient) SetSetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setError = err
}

// SetScanError sets an error to be returned by Scan
func (m *MockRedisClient) SetScanError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scanError = err
}

// SetScanKeys sets the keys to be returned by Scan
func (m *MockRedisClient) SetScanKeys(keys []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scanKeys = keys
}

// GetCallLog returns all logged method calls for test verification
func (m *MockRedisClient) GetCallLog() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Return a copy
	log := make([]string, len(m.callLog))
	copy(log, m.callLog)
	return log
}

// ClearCallLog clears the call log
func (m *MockRedisClient) ClearCallLog() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callLog = make([]string, 0)
}

// HasCall checks if a specific call was made
func (m *MockRedisClient) HasCall(call string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, logged := range m.callLog {
		if logged == call {
			return true
		}
	}
	return false
}

// GetData returns a copy of the internal data for test verification
func (m *MockRedisClient) GetData() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	data := make(map[string]string)
	for k, v := range m.data {
		data[k] = v
	}
	return data
}

// Ping tests connectivity to Redis server
func (m *MockRedisClient) Ping(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.callLog = append(m.callLog, "Ping")
	return m.pingError
}

// Get retrieves a value from Redis by key
func (m *MockRedisClient) Get(ctx context.Context, key string) (string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.callLog = append(m.callLog, "Get:"+key)
	
	if m.getError != nil {
		return "", false, m.getError
	}
	
	val, exists := m.data[key]
	return val, exists, nil
}

// Set stores a key-value pair in Redis with optional expiration
func (m *MockRedisClient) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.callLog = append(m.callLog, "Set:"+key+":"+value)
	
	if m.setError != nil {
		return m.setError
	}
	
	m.data[key] = value
	return nil
}

// Scan iterates over keys matching a pattern
func (m *MockRedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.callLog = append(m.callLog, "Scan")
	
	if m.scanError != nil {
		return nil, 0, m.scanError
	}
	
	// If custom scan keys are set, return them
	if m.scanKeys != nil {
		return m.scanKeys, 0, nil
	}
	
	// Otherwise, return all keys in data
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	
	return keys, 0, nil
}

// Close closes the Redis connection
func (m *MockRedisClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.callLog = append(m.callLog, "Close")
	return nil
}