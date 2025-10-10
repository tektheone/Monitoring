package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"fleet-management-server/internal/store"
)

func setupTestServer(t *testing.T) *Server {
	// Create a temporary CSV file with test devices
	csvContent := `device_id
60-6b-44-84-dc-64
b4-45-52-a2-f1-3c
26-9a-66-01-33-83
18-b8-87-e7-1f-06
38-4e-73-e0-33-59
`
	
	tmpFile, err := os.CreateTemp("", "devices_test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()
	
	// Create store and load devices
	s := store.New()
	err = s.LoadFromCSV(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load CSV: %v", err)
	}
	
	return New(s)
}

func TestHeartbeatEndpoint(t *testing.T) {
	server := setupTestServer(t)
	router := server.Router()
	
	// Test successful heartbeat
	heartbeatReq := HeartbeatRequest{
		SentAt: "2025-10-10T12:15:00Z",
	}
	
	reqBody, _ := json.Marshal(heartbeatReq)
	req := httptest.NewRequest("POST", "/api/v1/devices/60-6b-44-84-dc-64/heartbeat", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should return 204 No Content
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHeartbeatEndpointUnknownDevice(t *testing.T) {
	server := setupTestServer(t)
	router := server.Router()
	
	heartbeatReq := HeartbeatRequest{
		SentAt: "2025-10-10T12:15:00Z",
	}
	
	reqBody, _ := json.Marshal(heartbeatReq)
	req := httptest.NewRequest("POST", "/api/v1/devices/unknown-device/heartbeat", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should return 404 + error message
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
	
	var errorResp ErrorResponse
	json.NewDecoder(w.Body).Decode(&errorResp)
	if errorResp.Message != "Device not found" {
		t.Errorf("Expected 'Device not found', got '%s'", errorResp.Message)
	}
}

func TestHeartbeatEndpointInvalidJSON(t *testing.T) {
	server := setupTestServer(t)
	router := server.Router()
	
	req := httptest.NewRequest("POST", "/api/v1/devices/60-6b-44-84-dc-64/heartbeat", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should return 400 + error message
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
	
	var errorResp ErrorResponse
	json.NewDecoder(w.Body).Decode(&errorResp)
	if errorResp.Message != "Invalid request payload" {
		t.Errorf("Expected 'Invalid request payload', got '%s'", errorResp.Message)
	}
}

func TestStatsPostEndpoint(t *testing.T) {
	server := setupTestServer(t)
	router := server.Router()
	
	// Test successful stats post
	statsReq := StatsRequest{
		SentAt:     "2025-10-10T12:15:00Z",
		UploadTime: 250000000, // 250ms in nanoseconds
	}
	
	reqBody, _ := json.Marshal(statsReq)
	req := httptest.NewRequest("POST", "/api/v1/devices/60-6b-44-84-dc-64/stats", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should return 204 No Content
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestStatsGetEndpointNoData(t *testing.T) {
	server := setupTestServer(t)
	router := server.Router()
	
	// Test GET stats with no heartbeats AND no uploads
	req := httptest.NewRequest("GET", "/api/v1/devices/60-6b-44-84-dc-64/stats", nil)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should return 204 No Content
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestStatsGetEndpointWithData(t *testing.T) {
	server := setupTestServer(t)
	router := server.Router()
	
	deviceID := "60-6b-44-84-dc-64"
	
	// Add some heartbeats and upload stats
	baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	server.store.AddHeartbeat(deviceID, baseTime)
	server.store.AddHeartbeat(deviceID, baseTime.Add(1*time.Minute))
	server.store.AddHeartbeat(deviceID, baseTime.Add(2*time.Minute))
	
	server.store.AddUploadTime(deviceID, 250*time.Millisecond)
	server.store.AddUploadTime(deviceID, 300*time.Millisecond)
	
	// Test GET stats with data
	req := httptest.NewRequest("GET", "/api/v1/devices/60-6b-44-84-dc-64/stats", nil)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should return 200 + JSON response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var statsResp StatsResponse
	err := json.NewDecoder(w.Body).Decode(&statsResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	// Verify response format
	if statsResp.AvgUploadTime == "" {
		t.Error("Expected non-empty avg_upload_time")
	}
	
	if statsResp.Uptime <= 0 {
		t.Errorf("Expected positive uptime, got %f", statsResp.Uptime)
	}
	
	// Verify uptime is 100% (3 consecutive minutes)
	if statsResp.Uptime != 100.0 {
		t.Errorf("Expected 100%% uptime, got %f", statsResp.Uptime)
	}
}

func TestInvalidEndpoints(t *testing.T) {
	server := setupTestServer(t)
	router := server.Router()
	
	// Test invalid endpoint
	req := httptest.NewRequest("GET", "/api/v1/devices/60-6b-44-84-dc-64/invalid", nil)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should return 404
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	server := setupTestServer(t)
	router := server.Router()
	
	// Test wrong method for heartbeat endpoint
	req := httptest.NewRequest("GET", "/api/v1/devices/60-6b-44-84-dc-64/heartbeat", nil)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should return 405 Method Not Allowed
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHealthEndpoint(t *testing.T) {
	server := setupTestServer(t)
	router := server.Router()
	
	req := httptest.NewRequest("GET", "/health", nil)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should return 200
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
