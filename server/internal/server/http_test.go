package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bjia56/spacenet/server/api"
	"github.com/gorilla/mux"
)

func TestNewHTTPHandler(t *testing.T) {
	mockStore := NewMockStore()
	handler := NewHTTPHandler(mockStore)
	
	if handler == nil {
		t.Fatal("NewHTTPHandler() returned nil")
	}
	
	if handler.store != mockStore {
		t.Error("NewHTTPHandler() did not set store correctly")
	}
}

func TestHTTPHandler_RegisterRoutes(t *testing.T) {
	mockStore := NewMockStore()
	handler := NewHTTPHandler(mockStore)
	router := mux.NewRouter()
	
	handler.RegisterRoutes(router)
	
	// Test that routes are registered by trying to match them
	testRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/health"},
		{"GET", "/api/ip/2001:db8::1"},
		{"GET", "/api/subnet/2001:db8::/64"},
	}
	
	for _, route := range testRoutes {
		req := httptest.NewRequest(route.method, route.path, nil)
		match := &mux.RouteMatch{}
		
		if !router.Match(req, match) {
			t.Errorf("Route %s %s was not registered", route.method, route.path)
		}
	}
}

func TestHTTPHandler_HandleHealth(t *testing.T) {
	mockStore := NewMockStore()
	handler := NewHTTPHandler(mockStore)
	
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	handler.handleHealth(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type application/json, got %s", contentType)
	}
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func TestHTTPHandler_HandleGetClaimByIP_Success(t *testing.T) {
	mockStore := NewMockStore()
	handler := NewHTTPHandler(mockStore)
	
	// Setup mock store to return a claim
	testIP := "2001:db8::1"
	testClaimant := "alice"
	mockStore.AddClaim(testIP, testClaimant)
	
	// Create request with mux variables
	req := httptest.NewRequest("GET", "/api/ip/"+testIP, nil)
	w := httptest.NewRecorder()
	
	// Set up mux variables (simulate what gorilla/mux would do)
	req = mux.SetURLVars(req, map[string]string{"ip": testIP})
	
	handler.handleGetClaimByIP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type application/json, got %s", contentType)
	}
	
	var response api.ClaimResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response.IP != testIP {
		t.Errorf("Expected IP %s, got %s", testIP, response.IP)
	}
	
	if response.Claimant != testClaimant {
		t.Errorf("Expected claimant %s, got %s", testClaimant, response.Claimant)
	}
}

func TestHTTPHandler_HandleGetClaimByIP_NotFound(t *testing.T) {
	mockStore := NewMockStore()
	handler := NewHTTPHandler(mockStore)
	
	testIP := "2001:db8::1"
	
	req := httptest.NewRequest("GET", "/api/ip/"+testIP, nil)
	w := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"ip": testIP})
	
	handler.handleGetClaimByIP(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response["error"] == "" {
		t.Error("Expected error message in response")
	}
}

func TestHTTPHandler_HandleGetClaimByIP_InvalidIP(t *testing.T) {
	mockStore := NewMockStore()
	handler := NewHTTPHandler(mockStore)
	
	testCases := []struct {
		name   string
		ip     string
		status int
	}{
		{
			name:   "empty IP",
			ip:     "",
			status: http.StatusBadRequest,
		},
		{
			name:   "invalid IP format",
			ip:     "invalid-ip",
			status: http.StatusBadRequest,
		},
		{
			name:   "malformed IPv6",
			ip:     "2001:db8:::1",
			status: http.StatusBadRequest,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/ip/"+tc.ip, nil)
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"ip": tc.ip})
			
			handler.handleGetClaimByIP(w, req)
			
			if w.Code != tc.status {
				t.Errorf("Expected status %d, got %d", tc.status, w.Code)
			}
			
			var response map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
			
			if response["error"] == "" {
				t.Error("Expected error message in response")
			}
		})
	}
}

func TestHTTPHandler_HandleGetClaimByIP_MissingIP(t *testing.T) {
	mockStore := NewMockStore()
	handler := NewHTTPHandler(mockStore)
	
	req := httptest.NewRequest("GET", "/api/ip/", nil)
	w := httptest.NewRecorder()
	// Don't set mux variables to simulate missing IP
	
	handler.handleGetClaimByIP(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHTTPHandler_HandleGetStatsBySubnet_Success(t *testing.T) {
	mockStore := NewMockStore()
	handler := NewHTTPHandler(mockStore)
	
	// Setup mock store to return subnet stats
	testSubnet := "2001:db8::/64"
	testStats := &SubnetStats{
		Owner:      "alice",
		Percentage: 75.5,
	}
	mockStore.AddSubnetStats(testSubnet, testStats)
	
	req := httptest.NewRequest("GET", "/api/subnet/2001:db8::/64", nil)
	w := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{
		"address": "2001:db8::",
		"prefix":  "64",
	})
	
	handler.handleGetStatsBySubnet(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type application/json, got %s", contentType)
	}
	
	var response api.SubnetResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response.Owner != testStats.Owner {
		t.Errorf("Expected owner %s, got %s", testStats.Owner, response.Owner)
	}
	
	if response.Percentage != testStats.Percentage {
		t.Errorf("Expected percentage %f, got %f", testStats.Percentage, response.Percentage)
	}
}

func TestHTTPHandler_HandleGetStatsBySubnet_InvalidSubnet(t *testing.T) {
	mockStore := NewMockStore()
	handler := NewHTTPHandler(mockStore)
	
	// Configure mock to return false for invalid subnet
	mockStore.SetGetSubnetStatsFunc(func(subnet string) (*SubnetStats, bool) {
		return nil, false
	})
	
	req := httptest.NewRequest("GET", "/api/subnet/invalid/subnet", nil)
	w := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{
		"address": "invalid",
		"prefix":  "subnet",
	})
	
	handler.handleGetStatsBySubnet(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response["error"] == "" {
		t.Error("Expected error message in response")
	}
}

func TestHTTPHandler_FullIntegration(t *testing.T) {
	mockStore := NewMockStore()
	handler := NewHTTPHandler(mockStore)
	router := mux.NewRouter()
	handler.RegisterRoutes(router)
	
	// Test data
	testIP := "2001:db8::1"
	testClaimant := "alice"
	testSubnet := "2001:db8::/64"
	testStats := &SubnetStats{
		Owner:      "alice",
		Percentage: 80.0,
	}
	
	// Setup mock data
	mockStore.AddClaim(testIP, testClaimant)
	mockStore.AddSubnetStats(testSubnet, testStats)
	
	// Test health endpoint
	t.Run("health", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Health check failed with status %d", w.Code)
		}
	})
	
	// Test IP claim endpoint
	t.Run("ip claim", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/ip/"+testIP, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("IP claim request failed with status %d", w.Code)
		}
		
		var response api.ClaimResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal claim response: %v", err)
		}
		
		if response.Claimant != testClaimant {
			t.Errorf("Expected claimant %s, got %s", testClaimant, response.Claimant)
		}
	})
	
	// Test subnet stats endpoint
	t.Run("subnet stats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/subnet/2001:db8::/64", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Subnet stats request failed with status %d", w.Code)
		}
		
		var response api.SubnetResponse  
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal subnet response: %v", err)
		}
		
		if response.Owner != testStats.Owner {
			t.Errorf("Expected owner %s, got %s", testStats.Owner, response.Owner)
		}
	})
}