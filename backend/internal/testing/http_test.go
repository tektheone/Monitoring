package testing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"fleet-management-server/internal/server"
	"fleet-management-server/internal/store"
)

func setupHTTPTestServer(t *testing.T) *server.Server {
	// Create a temporary CSV file with test devices
	csvContent := `device_id
60-6b-44-84-dc-64
b4-45-52-a2-f1-3c
26-9a-66-01-33-83
18-b8-87-e7-1f-06
38-4e-73-e0-33-59
`
	
	tmpFile, err := os.CreateTemp("", "devices_http_test_*.csv")
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
	
	return server.New(s)
}

// HTTP Tests - Validate exact HTTP status codes and JSON responses

// TestPOSTSuccessResponses tests POST success scenarios with 204 responses
func TestPOSTSuccessResponses(t *testing.T) {
	srv := setupHTTPTestServer(t)
	router := srv.Router()
	
	// Test POST /api/v1/devices/{device_id}/heartbeat success
	t.Run("HeartbeatPOSTSuccess204", func(t *testing.T) {
		heartbeatReq := server.HeartbeatRequest{
			SentAt: "2025-10-10T12:15:00Z",
		}
		
		reqBody, _ := json.Marshal(heartbeatReq)
		req := httptest.NewRequest("POST", "/api/v1/devices/60-6b-44-84-dc-64/heartbeat", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Verify exact 204 No Content response
		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204 No Content, got %d", w.Code)
		}
		
		// Verify no response body for 204
		if w.Body.Len() != 0 {
			t.Errorf("Expected empty body for 204 response, got: %s", w.Body.String())
		}
	})
	
	// Test POST /api/v1/devices/{device_id}/stats success
	t.Run("StatsPOSTSuccess204", func(t *testing.T) {
		statsReq := server.StatsRequest{
			SentAt:     "2025-10-10T12:15:00Z",
			UploadTime: 250000000, // 250ms in nanoseconds
		}
		
		reqBody, _ := json.Marshal(statsReq)
		req := httptest.NewRequest("POST", "/api/v1/devices/60-6b-44-84-dc-64/stats", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Verify exact 204 No Content response
		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204 No Content, got %d", w.Code)
		}
		
		// Verify no response body for 204
		if w.Body.Len() != 0 {
			t.Errorf("Expected empty body for 204 response, got: %s", w.Body.String())
		}
	})
}

// TestDeviceNotFoundHandling tests 404 handling for unknown device IDs
func TestDeviceNotFoundHandling(t *testing.T) {
	srv := setupHTTPTestServer(t)
	router := srv.Router()
	
	unknownDeviceID := "unknown-device-12345"
	
	// Test POST heartbeat with unknown device → 404
	t.Run("HeartbeatPOSTUnknownDevice404", func(t *testing.T) {
		heartbeatReq := server.HeartbeatRequest{
			SentAt: "2025-10-10T12:15:00Z",
		}
		
		reqBody, _ := json.Marshal(heartbeatReq)
		req := httptest.NewRequest("POST", "/api/v1/devices/"+unknownDeviceID+"/heartbeat", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Verify exact 404 Not Found response
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404 Not Found, got %d", w.Code)
		}
		
		// Verify exact JSON format: {"msg": "Device not found"}
		var errorResp server.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResp)
		if err != nil {
			t.Fatalf("Failed to decode error response: %v", err)
		}
		
		expectedMsg := "Device not found"
		if errorResp.Message != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, errorResp.Message)
		}
		
		// Verify Content-Type header
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}
	})
	
	// Test GET stats with unknown device → 404
	t.Run("StatsGETUnknownDevice404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/devices/"+unknownDeviceID+"/stats", nil)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Verify exact 404 Not Found response
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404 Not Found, got %d", w.Code)
		}
		
		// Verify exact JSON format: {"msg": "Device not found"}
		var errorResp server.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResp)
		if err != nil {
			t.Fatalf("Failed to decode error response: %v", err)
		}
		
		expectedMsg := "Device not found"
		if errorResp.Message != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, errorResp.Message)
		}
	})
}

// TestInvalidJSONHandling tests 400 handling for invalid JSON payloads
func TestInvalidJSONHandling(t *testing.T) {
	srv := setupHTTPTestServer(t)
	router := srv.Router()
	
	validDeviceID := "60-6b-44-84-dc-64"
	
	// Test POST heartbeat with invalid JSON → 400
	t.Run("HeartbeatPOSTInvalidJSON400", func(t *testing.T) {
		invalidJSON := `{"sent_at": "invalid-timestamp-format"}`
		
		req := httptest.NewRequest("POST", "/api/v1/devices/"+validDeviceID+"/heartbeat", strings.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Verify exact 400 Bad Request response
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 Bad Request, got %d", w.Code)
		}
		
		// Verify exact JSON format: {"msg": "Invalid request payload"}
		var errorResp server.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResp)
		if err != nil {
			t.Fatalf("Failed to decode error response: %v", err)
		}
		
		expectedMsg := "Invalid request payload"
		if errorResp.Message != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, errorResp.Message)
		}
	})
	
	// Test POST heartbeat with malformed JSON → 400
	t.Run("HeartbeatPOSTMalformedJSON400", func(t *testing.T) {
		malformedJSON := `{"sent_at": "2025-10-10T12:15:00Z"` // Missing closing brace
		
		req := httptest.NewRequest("POST", "/api/v1/devices/"+validDeviceID+"/heartbeat", strings.NewReader(malformedJSON))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Verify exact 400 Bad Request response
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 Bad Request, got %d", w.Code)
		}
		
		// Verify exact JSON format: {"msg": "Invalid request payload"}
		var errorResp server.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResp)
		if err != nil {
			t.Fatalf("Failed to decode error response: %v", err)
		}
		
		expectedMsg := "Invalid request payload"
		if errorResp.Message != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, errorResp.Message)
		}
	})
}

// TestStatsGETLogic tests GET logic for 200 vs 204 response rules
func TestStatsGETLogic(t *testing.T) {
	srv := setupHTTPTestServer(t)
	router := srv.Router()
	
	deviceID := "b4-45-52-a2-f1-3c" // Use different device to avoid conflicts
	
	// Test GET stats with no data → 204 No Content
	t.Run("StatsGETNoData204", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/devices/"+deviceID+"/stats", nil)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Verify exact 204 No Content response (no heartbeats AND no uploads)
		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204 No Content, got %d", w.Code)
		}
		
		// Verify no response body for 204
		if w.Body.Len() != 0 {
			t.Errorf("Expected empty body for 204 response, got: %s", w.Body.String())
		}
	})
	
	// Test GET stats with data → 200 + JSON
	t.Run("StatsGETWithData200", func(t *testing.T) {
		// Add some heartbeats and upload stats first
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		
		// Add data via HTTP requests to simulate real usage
		heartbeatReq := server.HeartbeatRequest{
			SentAt: baseTime.Format(time.RFC3339),
		}
		reqBody, _ := json.Marshal(heartbeatReq)
		req := httptest.NewRequest("POST", "/api/v1/devices/"+deviceID+"/heartbeat", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		statsReq := server.StatsRequest{
			SentAt:     baseTime.Format(time.RFC3339),
			UploadTime: 250000000, // 250ms in nanoseconds
		}
		reqBody, _ = json.Marshal(statsReq)
		req = httptest.NewRequest("POST", "/api/v1/devices/"+deviceID+"/stats", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Now GET the stats
		req = httptest.NewRequest("GET", "/api/v1/devices/"+deviceID+"/stats", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Verify exact 200 OK response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 OK, got %d", w.Code)
		}
		
		// Verify Content-Type header
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}
		
		// Verify JSON response structure
		var statsResp server.StatsResponse
		err := json.NewDecoder(w.Body).Decode(&statsResp)
		if err != nil {
			t.Fatalf("Failed to decode stats response: %v", err)
		}
		
		// Verify response fields are present and valid
		if statsResp.AvgUploadTime == "" {
			t.Error("Expected non-empty avg_upload_time")
		}
		
		if statsResp.Uptime <= 0 {
			t.Errorf("Expected positive uptime, got %f", statsResp.Uptime)
		}
	})
}

// TestJSONResponseFormat tests exact JSON format validation
func TestJSONResponseFormat(t *testing.T) {
	srv := setupHTTPTestServer(t)
	router := srv.Router()
	
	// Test error response JSON structure: {"msg": "..."}
	t.Run("ErrorResponseJSONStructure", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/devices/unknown-device/heartbeat", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Parse response as raw JSON to verify exact structure
		var rawResponse map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&rawResponse)
		if err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}
		
		// Verify exact structure: only "msg" field should be present
		if len(rawResponse) != 1 {
			t.Errorf("Expected exactly 1 field in error response, got %d: %v", len(rawResponse), rawResponse)
		}
		
		// Verify "msg" field exists and is a string
		msgValue, exists := rawResponse["msg"]
		if !exists {
			t.Error("Expected 'msg' field in error response")
		}
		
		msgString, isString := msgValue.(string)
		if !isString {
			t.Errorf("Expected 'msg' field to be string, got %T", msgValue)
		}
		
		if msgString == "" {
			t.Error("Expected non-empty 'msg' field")
		}
	})
	
	// Test stats response JSON structure
	t.Run("StatsResponseJSONStructure", func(t *testing.T) {
		deviceID := "26-9a-66-01-33-83" // Use different device
		
		// Add test data via HTTP
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		
		heartbeatReq := server.HeartbeatRequest{
			SentAt: baseTime.Format(time.RFC3339),
		}
		reqBody, _ := json.Marshal(heartbeatReq)
		req := httptest.NewRequest("POST", "/api/v1/devices/"+deviceID+"/heartbeat", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		statsReq := server.StatsRequest{
			SentAt:     baseTime.Format(time.RFC3339),
			UploadTime: 250000000,
		}
		reqBody, _ = json.Marshal(statsReq)
		req = httptest.NewRequest("POST", "/api/v1/devices/"+deviceID+"/stats", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// GET the stats
		req = httptest.NewRequest("GET", "/api/v1/devices/"+deviceID+"/stats", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Parse response as raw JSON to verify exact structure
		var rawResponse map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&rawResponse)
		if err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}
		
		// Verify exact structure: "avg_upload_time" and "uptime" fields
		if len(rawResponse) != 2 {
			t.Errorf("Expected exactly 2 fields in stats response, got %d: %v", len(rawResponse), rawResponse)
		}
		
		// Verify "avg_upload_time" field exists and is a string
		avgUploadValue, exists := rawResponse["avg_upload_time"]
		if !exists {
			t.Error("Expected 'avg_upload_time' field in stats response")
		}
		
		_, isString := avgUploadValue.(string)
		if !isString {
			t.Errorf("Expected 'avg_upload_time' field to be string, got %T", avgUploadValue)
		}
		
		// Verify "uptime" field exists and is a number
		uptimeValue, exists := rawResponse["uptime"]
		if !exists {
			t.Error("Expected 'uptime' field in stats response")
		}
		
		_, isNumber := uptimeValue.(float64)
		if !isNumber {
			t.Errorf("Expected 'uptime' field to be number, got %T", uptimeValue)
		}
	})
}
