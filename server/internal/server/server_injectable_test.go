package server

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"
)

// Mock factories for testing

type mockUDPListener struct {
	addr      *net.UDPAddr
	closed    bool
	readFunc  func([]byte) (int, *net.UDPAddr, error)
	closeFunc func() error
}

func (m *mockUDPListener) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
	if m.readFunc != nil {
		return m.readFunc(b)
	}
	// Default behavior - simulate closed connection
	return 0, nil, &net.OpError{Op: "read", Net: "udp", Err: fmt.Errorf("closed")}
}

func (m *mockUDPListener) Close() error {
	m.closed = true
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *mockUDPListener) LocalAddr() net.Addr {
	if m.addr != nil {
		return m.addr
	}
	return &net.UDPAddr{IP: net.ParseIP("::1"), Port: 1337}
}

// Mock HTTP server
type mockHTTPServer struct {
	addr           string
	handler        http.Handler
	startFunc      func() error
	shutdownFunc   func() error
	listenAndServe func() error
}

func (m *mockHTTPServer) ListenAndServe() error {
	if m.listenAndServe != nil {
		return m.listenAndServe()
	}
	// Simulate successful start
	return nil
}

func (m *mockHTTPServer) Shutdown(ctx interface{}) error {
	if m.shutdownFunc != nil {
		return m.shutdownFunc()
	}
	return nil
}

func TestNewInjectableServer_WithDefaults(t *testing.T) {
	opts := InjectableServerOptions{
		Port:     1337,
		HTTPPort: 8080,
	}
	
	server := NewInjectableServer(opts)
	
	if server == nil {
		t.Fatal("NewInjectableServer returned nil")
	}
	
	if server.port != opts.Port {
		t.Errorf("Expected port %d, got %d", opts.Port, server.port)
	}
	
	if server.httpPort != opts.HTTPPort {
		t.Errorf("Expected HTTP port %d, got %d", opts.HTTPPort, server.httpPort)
	}
	
	if server.store == nil {
		t.Error("Server should have initialized store")
	}
	
	if server.httpHandler == nil {
		t.Error("Server should have initialized HTTP handler")
	}
}

func TestNewInjectableServer_WithInjectedStore(t *testing.T) {
	mockStore := NewMockStore()
	
	opts := InjectableServerOptions{
		Port:     1337,
		HTTPPort: 8080,
		Store:    mockStore,
	}
	
	server := NewInjectableServer(opts)
	
	if server.store != mockStore {
		t.Error("Server should use injected store")
	}
	
	// Verify the store is the same instance
	if server.GetStore() != mockStore {
		t.Error("GetStore() should return the injected store")
	}
}

func TestNewInjectableServer_WithCustomFactories(t *testing.T) {
	mockStore := NewMockStore()
	customStoreFactory := func() Store {
		return mockStore
	}
	
	var handlerCreated bool
	customHandlerFactory := func(store Store) *HTTPHandler {
		handlerCreated = true
		return NewHTTPHandler(store)
	}
	
	opts := InjectableServerOptions{
		Port:           1337,
		HTTPPort:       8080,
		StoreFactory:   customStoreFactory,
		HandlerFactory: customHandlerFactory,
	}
	
	server := NewInjectableServer(opts)
	
	if !handlerCreated {
		t.Error("Custom handler factory was not called")
	}
	
	if server.GetStore() != mockStore {
		t.Error("Custom store factory was not used")
	}
}

func TestInjectableServer_StartWithMocks(t *testing.T) {
	mockStore := NewMockStore()
	mockListener := &mockUDPListener{
		addr: &net.UDPAddr{IP: net.ParseIP("::1"), Port: 1337},
	}
	mockHTTPServer := &mockHTTPServer{
		listenAndServe: func() error {
			// Simulate successful HTTP server start
			return nil
		},
	}
	
	var listenerCreated bool
	var httpServerCreated bool
	
	opts := InjectableServerOptions{
		Port:     1337,
		HTTPPort: 8080,
		Store:    mockStore,
		ListenerFactory: func(ip string, port int) (*net.UDPConn, error) {
			listenerCreated = true
			// Return the mock as a *net.UDPConn (this is a hack for testing)
			// In real testing, you'd use an interface or more sophisticated mocking
			return (*net.UDPConn)(mockListener), nil
		},
		HTTPServerFactory: func(host string, port int, handler http.Handler) *http.Server {
			httpServerCreated = true
			return (*http.Server)(mockHTTPServer)
		},
	}
	
	server := NewInjectableServer(opts)
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	
	if !listenerCreated {
		t.Error("Listener factory was not called")
	}
	
	if !httpServerCreated {
		t.Error("HTTP server factory was not called")
	}
	
	// Clean up
	server.Stop()
	
	if !mockListener.closed {
		t.Error("Mock listener was not closed")
	}
}

func TestInjectableServer_StoreIntegration(t *testing.T) {
	mockStore := NewMockStore()
	
	// Track store interactions
	var processClaimCalled bool
	mockStore.SetProcessClaimFunc(func(ipAddr string, claimant string) error {
		processClaimCalled = true
		return nil
	})
	
	opts := InjectableServerOptions{
		Port:     0, // Use ephemeral port
		HTTPPort: 0, // Use ephemeral port
		Store:    mockStore,
	}
	
	server := NewInjectableServer(opts)
	
	// Test direct store access
	err := server.GetStore().ProcessClaim("2001:db8::1", "test_user")
	if err != nil {
		t.Fatalf("Failed to process claim: %v", err)
	}
	
	if !processClaimCalled {
		t.Error("ProcessClaim was not called on mock store")
	}
	
	// Verify the HTTP handler uses the same store
	handler := server.GetHTTPHandler()
	if handler.store != mockStore {
		t.Error("HTTP handler should use the same injected store")
	}
}

func TestInjectableServer_HTTPHandlerCustomization(t *testing.T) {
	mockStore := NewMockStore()
	
	// Custom handler factory that adds tracking
	var customHandlerCreated bool
	customHandlerFactory := func(store Store) *HTTPHandler {
		customHandlerCreated = true
		// Could customize the handler here
		return NewHTTPHandler(store)
	}
	
	opts := InjectableServerOptions{
		Port:           1337,
		HTTPPort:       8080,
		Store:          mockStore,
		HandlerFactory: customHandlerFactory,
	}
	
	server := NewInjectableServer(opts)
	
	if !customHandlerCreated {
		t.Error("Custom handler factory was not called")
	}
	
	handler := server.GetHTTPHandler()
	if handler == nil {
		t.Error("HTTP handler was not created")
	}
	
	if handler.store != mockStore {
		t.Error("HTTP handler should use the injected store")
	}
}

func TestInjectableServer_ErrorHandling(t *testing.T) {
	mockStore := NewMockStore()
	
	// Factory that returns an error
	failingListenerFactory := func(ip string, port int) (*net.UDPConn, error) {
		return nil, fmt.Errorf("simulated listener creation failure")
	}
	
	opts := InjectableServerOptions{
		Port:            1337,
		HTTPPort:        8080,
		Store:           mockStore,
		ListenerFactory: failingListenerFactory,
	}
	
	server := NewInjectableServer(opts)
	
	err := server.Start()
	if err == nil {
		t.Error("Expected error when listener factory fails")
		server.Stop() // Clean up if it somehow started
	}
	
	if err != nil && !fmt.Sprintf("%v", err).Contains("failed to start UDP server") {
		t.Errorf("Expected UDP server error, got: %v", err)
	}
}

func TestInjectableServer_StopMultipleTimes(t *testing.T) {
	mockStore := NewMockStore()
	
	opts := InjectableServerOptions{
		Port:     0,
		HTTPPort: 0,
		Store:    mockStore,
	}
	
	server := NewInjectableServer(opts)
	
	// Start server
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	
	// Stop multiple times (should be safe)
	server.Stop()
	server.Stop() // Second call should not panic
	server.Stop() // Third call should not panic
	
	// Verify store was closed (mock store tracks this)
	callLog := mockStore.GetCallLog()
	closeCallFound := false
	for _, call := range callLog {
		if call == "Close" {
			closeCallFound = true
			break
		}
	}
	
	if !closeCallFound {
		t.Error("Store.Close() was not called")
	}
}

func TestInjectableServer_ConcurrentAccessToInjectedStore(t *testing.T) {
	mockStore := NewMockStore()
	
	opts := InjectableServerOptions{
		Port:     0,
		HTTPPort: 0,
		Store:    mockStore,
	}
	
	server := NewInjectableServer(opts)
	
	// Test concurrent store access through server
	done := make(chan bool)
	numGoroutines := 10
	claimsPerGoroutine := 20
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			for j := 0; j < claimsPerGoroutine; j++ {
				ip := fmt.Sprintf("2001:db8::%d:%d", id, j)
				claimant := fmt.Sprintf("user_%d_%d", id, j)
				
				err := server.GetStore().ProcessClaim(ip, claimant)
				if err != nil {
					t.Errorf("ProcessClaim failed: %v", err)
				}
			}
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}
	
	// Verify all claims were processed
	allClaims := server.GetStore().GetAllClaims()
	expectedCount := numGoroutines * claimsPerGoroutine
	
	if len(allClaims) != expectedCount {
		t.Errorf("Expected %d claims, got %d", expectedCount, len(allClaims))
	}
}

// TestInjectableServer_ComparisonWithOriginal demonstrates the testing advantages
func TestInjectableServer_TestabilityComparison(t *testing.T) {
	// With injectable server, we can easily mock all dependencies
	mockStore := NewMockStore()
	mockStore.AddClaim("2001:db8::1", "alice")
	
	opts := InjectableServerOptions{
		Port:     1337,
		HTTPPort: 8080,
		Store:    mockStore,
	}
	
	server := NewInjectableServer(opts)
	
	// We can verify exact interactions with the store
	claimant, exists := server.GetStore().GetClaim("2001:db8::1")
	if !exists {
		t.Error("Expected claim to exist")
	}
	
	if claimant != "alice" {
		t.Errorf("Expected claimant 'alice', got '%s'", claimant)
	}
	
	// We can verify that the HTTP handler uses the same store instance
	if server.GetHTTPHandler().store != mockStore {
		t.Error("HTTP handler should use the injected mock store")
	}
	
	// We can inspect call logs from the mock
	mockStore.GetClaim("test") // Make a call
	callLog := mockStore.GetCallLog()
	
	found := false
	for _, call := range callLog {
		if call == "GetClaim:test" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected GetClaim call to be logged")
	}
	
	t.Log("Injectable server provides much better testability than the original")
}