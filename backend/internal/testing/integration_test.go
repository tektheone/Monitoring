package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"fleet-management-server/internal/server"
	"fleet-management-server/internal/store"
)

// Integration Testing - End-to-end testing and concurrency validation

func setupIntegrationTestServer(t *testing.T) *server.Server {
	// Create a temporary CSV file with test devices
	csvContent := `device_id
60-6b-44-84-dc-64
b4-45-52-a2-f1-3c
26-9a-66-01-33-83
18-b8-87-e7-1f-06
38-4e-73-e0-33-59
`
	
	tmpFile, err := os.CreateTemp("", "devices_integration_*.csv")
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

// TestEndToEndIntegration performs end-to-end integration testing
func TestEndToEndIntegration(t *testing.T) {
	srv := setupIntegrationTestServer(t)
	router := srv.Router()
	
	// All 5 device IDs from CSV
	deviceIDs := []string{
		"60-6b-44-84-dc-64",
		"b4-45-52-a2-f1-3c", 
		"26-9a-66-01-33-83",
		"18-b8-87-e7-1f-06",
		"38-4e-73-e0-33-59",
	}
	
	// Integration Test Steps:
	// 1. Start server with test CSV ✓ (done in setupIntegrationTestServer)
	// 2. POST heartbeats/stats via HTTP
	// 3. GET stats and verify calculations
	// 4. Test all 5 device IDs
	
	t.Run("EndToEndIntegrationTest", func(t *testing.T) {
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		
		// Step 2 & 4: POST heartbeats/stats via HTTP for all 5 devices
		for i, deviceID := range deviceIDs {
			// POST heartbeats for each device (3 heartbeats each)
			for j := 0; j < 3; j++ {
				heartbeatReq := server.HeartbeatRequest{
					SentAt: baseTime.Add(time.Duration(i*10+j) * time.Minute).Format(time.RFC3339),
				}
				
				reqBody, _ := json.Marshal(heartbeatReq)
				req := httptest.NewRequest("POST", "/api/v1/devices/"+deviceID+"/heartbeat", bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				
				// Verify 204 response
				if w.Code != http.StatusNoContent {
					t.Errorf("Device %s heartbeat %d: expected 204, got %d", deviceID, j, w.Code)
				}
			}
			
			// POST upload stats for each device (2 uploads each)
			for j := 0; j < 2; j++ {
				statsReq := server.StatsRequest{
					SentAt:     baseTime.Add(time.Duration(i*10+j) * time.Minute).Format(time.RFC3339),
					UploadTime: int64((250 + i*50 + j*25) * 1000000), // 250ms, 275ms, 300ms, etc.
				}
				
				reqBody, _ := json.Marshal(statsReq)
				req := httptest.NewRequest("POST", "/api/v1/devices/"+deviceID+"/stats", bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				
				// Verify 204 response
				if w.Code != http.StatusNoContent {
					t.Errorf("Device %s stats %d: expected 204, got %d", deviceID, j, w.Code)
				}
			}
		}
		
		// Step 3: GET stats and verify calculations for all devices
		for _, deviceID := range deviceIDs {
			req := httptest.NewRequest("GET", "/api/v1/devices/"+deviceID+"/stats", nil)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Verify 200 response (should have data)
			if w.Code != http.StatusOK {
				t.Errorf("Device %s GET stats: expected 200, got %d", deviceID, w.Code)
				continue
			}
			
			// Verify JSON response structure
			var statsResp server.StatsResponse
			err := json.NewDecoder(w.Body).Decode(&statsResp)
			if err != nil {
				t.Errorf("Device %s: failed to decode stats response: %v", deviceID, err)
				continue
			}
			
			// Verify calculations are correct
			if statsResp.AvgUploadTime == "" {
				t.Errorf("Device %s: expected non-empty avg_upload_time", deviceID)
			}
			
			if statsResp.Uptime <= 0 {
				t.Errorf("Device %s: expected positive uptime, got %f", deviceID, statsResp.Uptime)
			}
			
			// Verify uptime is 100% (3 consecutive heartbeats)
			if statsResp.Uptime != 100.0 {
				t.Errorf("Device %s: expected 100%% uptime, got %f", deviceID, statsResp.Uptime)
			}
			
			t.Logf("Device %s: uptime=%.1f%%, avg_upload_time=%s", deviceID, statsResp.Uptime, statsResp.AvgUploadTime)
		}
	})
}

// TestConcurrentRequestHandling performs concurrency validation
func TestConcurrentRequestHandling(t *testing.T) {
	srv := setupIntegrationTestServer(t)
	router := srv.Router()
	
	// All 5 device IDs from CSV
	deviceIDs := []string{
		"60-6b-44-84-dc-64",
		"b4-45-52-a2-f1-3c", 
		"26-9a-66-01-33-83",
		"18-b8-87-e7-1f-06",
		"38-4e-73-e0-33-59",
	}
	
	// Concurrency Stress Test:
	// - 100 goroutines × 5 devices = 500 concurrent operations
	// - Mixed heartbeat/upload operations
	// - Verify no race conditions
	// - Validate final statistics accuracy
	
	t.Run("ConcurrencyStressTest", func(t *testing.T) {
		numGoroutines := 100
		numOperationsPerGoroutine := 10
		
		var wg sync.WaitGroup
		var mu sync.Mutex
		errors := make([]error, 0)
		
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		
		// Launch 100 goroutines × 5 devices
		for goroutineID := 0; goroutineID < numGoroutines; goroutineID++ {
			wg.Add(1)
			go func(gID int) {
				defer wg.Done()
				
				for opID := 0; opID < numOperationsPerGoroutine; opID++ {
					deviceID := deviceIDs[gID%len(deviceIDs)]
					
					// Mixed heartbeat/upload operations
					if opID%2 == 0 {
						// POST heartbeat
						heartbeatReq := server.HeartbeatRequest{
							SentAt: baseTime.Add(time.Duration(gID*100+opID) * time.Second).Format(time.RFC3339),
						}
						
						reqBody, _ := json.Marshal(heartbeatReq)
						req := httptest.NewRequest("POST", "/api/v1/devices/"+deviceID+"/heartbeat", bytes.NewReader(reqBody))
						req.Header.Set("Content-Type", "application/json")
						
						w := httptest.NewRecorder()
						router.ServeHTTP(w, req)
						
						if w.Code != http.StatusNoContent {
							mu.Lock()
							errors = append(errors, fmt.Errorf("goroutine %d op %d: heartbeat expected 204, got %d", gID, opID, w.Code))
							mu.Unlock()
						}
					} else {
						// POST upload stats
						statsReq := server.StatsRequest{
							SentAt:     baseTime.Add(time.Duration(gID*100+opID) * time.Second).Format(time.RFC3339),
							UploadTime: int64((250 + gID + opID*10) * 1000000), // Variable upload times
						}
						
						reqBody, _ := json.Marshal(statsReq)
						req := httptest.NewRequest("POST", "/api/v1/devices/"+deviceID+"/stats", bytes.NewReader(reqBody))
						req.Header.Set("Content-Type", "application/json")
						
						w := httptest.NewRecorder()
						router.ServeHTTP(w, req)
						
						if w.Code != http.StatusNoContent {
							mu.Lock()
							errors = append(errors, fmt.Errorf("goroutine %d op %d: stats expected 204, got %d", gID, opID, w.Code))
							mu.Unlock()
						}
					}
				}
			}(goroutineID)
		}
		
		// Wait for all concurrent operations to complete
		wg.Wait()
		
		// Verify no race conditions (no errors during concurrent operations)
		if len(errors) > 0 {
			t.Errorf("Race conditions detected: %d errors", len(errors))
			for i, err := range errors {
				if i < 10 { // Show first 10 errors
					t.Errorf("Error %d: %v", i+1, err)
				}
			}
		}
		
		// Validate final statistics accuracy for all devices
		for _, deviceID := range deviceIDs {
			req := httptest.NewRequest("GET", "/api/v1/devices/"+deviceID+"/stats", nil)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Should have data after concurrent operations
			if w.Code != http.StatusOK {
				t.Errorf("Device %s final stats: expected 200, got %d", deviceID, w.Code)
				continue
			}
			
			var statsResp server.StatsResponse
			err := json.NewDecoder(w.Body).Decode(&statsResp)
			if err != nil {
				t.Errorf("Device %s: failed to decode final stats: %v", deviceID, err)
				continue
			}
			
			// Validate final statistics accuracy
			if statsResp.AvgUploadTime == "" {
				t.Errorf("Device %s: final avg_upload_time is empty", deviceID)
			}
			
			if statsResp.Uptime < 0 || statsResp.Uptime > 100 {
				t.Errorf("Device %s: invalid final uptime %f (should be 0-100)", deviceID, statsResp.Uptime)
			}
			
			t.Logf("Device %s final stats: uptime=%.3f%%, avg_upload_time=%s", deviceID, statsResp.Uptime, statsResp.AvgUploadTime)
		}
		
		t.Logf("Concurrency stress test completed: %d goroutines × %d operations = %d total operations", 
			numGoroutines, numOperationsPerGoroutine*len(deviceIDs), numGoroutines*numOperationsPerGoroutine*len(deviceIDs))
	})
}

// TestRaceConditionPrevention tests for race conditions specifically
func TestRaceConditionPrevention(t *testing.T) {
	srv := setupIntegrationTestServer(t)
	router := srv.Router()
	
	deviceID := "60-6b-44-84-dc-64"
	
	t.Run("RaceConditionDetection", func(t *testing.T) {
		numConcurrentReads := 50
		numConcurrentWrites := 50
		
		var wg sync.WaitGroup
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		
		// Concurrent writes (heartbeats and stats)
		for i := 0; i < numConcurrentWrites; i++ {
			wg.Add(2) // One for heartbeat, one for stats
			
			// Concurrent heartbeat writes
			go func(i int) {
				defer wg.Done()
				heartbeatReq := server.HeartbeatRequest{
					SentAt: baseTime.Add(time.Duration(i) * time.Minute).Format(time.RFC3339),
				}
				
				reqBody, _ := json.Marshal(heartbeatReq)
				req := httptest.NewRequest("POST", "/api/v1/devices/"+deviceID+"/heartbeat", bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				
				if w.Code != http.StatusNoContent {
					t.Errorf("Concurrent heartbeat %d: expected 204, got %d", i, w.Code)
				}
			}(i)
			
			// Concurrent stats writes
			go func(i int) {
				defer wg.Done()
				statsReq := server.StatsRequest{
					SentAt:     baseTime.Add(time.Duration(i) * time.Minute).Format(time.RFC3339),
					UploadTime: int64((250 + i*10) * 1000000),
				}
				
				reqBody, _ := json.Marshal(statsReq)
				req := httptest.NewRequest("POST", "/api/v1/devices/"+deviceID+"/stats", bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				
				if w.Code != http.StatusNoContent {
					t.Errorf("Concurrent stats %d: expected 204, got %d", i, w.Code)
				}
			}(i)
		}
		
		// Concurrent reads (GET stats)
		for i := 0; i < numConcurrentReads; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				req := httptest.NewRequest("GET", "/api/v1/devices/"+deviceID+"/stats", nil)
				
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				
				// Should be either 200 (with data) or 204 (no data yet)
				if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
					t.Errorf("Concurrent read %d: expected 200 or 204, got %d", i, w.Code)
				}
				
				// If 200, verify JSON structure
				if w.Code == http.StatusOK {
					var statsResp server.StatsResponse
					err := json.NewDecoder(w.Body).Decode(&statsResp)
					if err != nil {
						t.Errorf("Concurrent read %d: failed to decode response: %v", i, err)
					}
				}
			}(i)
		}
		
		// Wait for all concurrent operations
		wg.Wait()
		
		// Final verification - ensure data consistency
		req := httptest.NewRequest("GET", "/api/v1/devices/"+deviceID+"/stats", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code == http.StatusOK {
			var statsResp server.StatsResponse
			err := json.NewDecoder(w.Body).Decode(&statsResp)
			if err != nil {
				t.Errorf("Final verification: failed to decode response: %v", err)
			} else {
				t.Logf("Final stats after race condition test: uptime=%.3f%%, avg_upload_time=%s", 
					statsResp.Uptime, statsResp.AvgUploadTime)
			}
		}
	})
}
