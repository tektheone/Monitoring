package store

import (
	"os"
	"sync"
	"testing"
	"time"
)

func TestStoreLoadFromCSV(t *testing.T) {
	// Create a temporary CSV file with the exact 5 devices
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
	
	// Test loading
	store := New()
	err = store.LoadFromCSV(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load CSV: %v", err)
	}
	
	// Verify exactly 5 devices loaded
	if store.DeviceCount() != 5 {
		t.Errorf("Expected 5 devices, got %d", store.DeviceCount())
	}
	
	// Verify all expected devices are present
	expectedDevices := []string{
		"60-6b-44-84-dc-64",
		"b4-45-52-a2-f1-3c", 
		"26-9a-66-01-33-83",
		"18-b8-87-e7-1f-06",
		"38-4e-73-e0-33-59",
	}
	
	for _, deviceID := range expectedDevices {
		device, exists := store.GetDevice(deviceID)
		if !exists {
			t.Errorf("Expected device %s not found", deviceID)
		}
		if device.id != deviceID {
			t.Errorf("Device ID mismatch: expected %s, got %s", deviceID, device.id)
		}
	}
}

func TestStoreLoadFromCSVValidation(t *testing.T) {
	// Test with wrong number of devices
	csvContent := `device_id
60-6b-44-84-dc-64
b4-45-52-a2-f1-3c
26-9a-66-01-33-83
18-b8-87-e7-1f-06
`
	
	tmpFile, err := os.CreateTemp("", "devices_invalid_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()
	
	store := New()
	err = store.LoadFromCSV(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for missing devices, got nil")
	}
}

func TestStoreLoadFromCSVUnexpectedDevice(t *testing.T) {
	// Test with unexpected device ID
	csvContent := `device_id
60-6b-44-84-dc-64
b4-45-52-a2-f1-3c
26-9a-66-01-33-83
18-b8-87-e7-1f-06
unexpected-device-id
`
	
	tmpFile, err := os.CreateTemp("", "devices_unexpected_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()
	
	store := New()
	err = store.LoadFromCSV(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for unexpected device ID, got nil")
	}
}

func TestStoreConcurrentOperations(t *testing.T) {
	// Create store with test devices
	csvContent := `device_id
60-6b-44-84-dc-64
b4-45-52-a2-f1-3c
26-9a-66-01-33-83
18-b8-87-e7-1f-06
38-4e-73-e0-33-59
`
	
	tmpFile, err := os.CreateTemp("", "devices_concurrent_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()
	
	store := New()
	err = store.LoadFromCSV(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load CSV: %v", err)
	}
	
	// Test concurrent reads and writes
	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100
	
	// Concurrent heartbeat additions
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			deviceID := "60-6b-44-84-dc-64"
			
			for j := 0; j < numOperations; j++ {
				timestamp := time.Now().Add(time.Duration(j) * time.Second)
				store.AddHeartbeat(deviceID, timestamp)
			}
		}(i)
	}
	
	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			deviceID := "60-6b-44-84-dc-64"
			
			for j := 0; j < numOperations; j++ {
				store.CalculateUptime(deviceID)
				store.GetDeviceStats(deviceID)
			}
		}()
	}
	
	// Concurrent upload time additions
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			deviceID := "b4-45-52-a2-f1-3c"
			
			for j := 0; j < numOperations; j++ {
				duration := time.Duration(j+1) * time.Millisecond
				store.AddUploadTime(deviceID, duration)
			}
		}()
	}
	
	wg.Wait()
	
	// Verify operations completed without race conditions
	uptime := store.CalculateUptime("60-6b-44-84-dc-64")
	if uptime < 0 || uptime > 100 {
		t.Errorf("Invalid uptime after concurrent operations: %f", uptime)
	}
	
	avgUpload := store.CalculateAverageUploadTime("b4-45-52-a2-f1-3c")
	if avgUpload <= 0 {
		t.Errorf("Invalid average upload time after concurrent operations: %v", avgUpload)
	}
}

func TestStoreInvalidCSVHeader(t *testing.T) {
	// Test with invalid header
	csvContent := `invalid_header
60-6b-44-84-dc-64
`
	
	tmpFile, err := os.CreateTemp("", "devices_invalid_header_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()
	
	store := New()
	err = store.LoadFromCSV(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid CSV header, got nil")
	}
}
