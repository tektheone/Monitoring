package store

import (
	"sync"
	"testing"
	"time"
)

// TestConcurrentSimulatorRequests tests the optimized concurrency implementation
func TestConcurrentSimulatorRequests(t *testing.T) {
	// Create store with test devices
	store := New()
	
	// Add test devices manually for concurrency testing
	store.mu.Lock()
	for _, deviceID := range []string{
		"60-6b-44-84-dc-64",
		"b4-45-52-a2-f1-3c", 
		"26-9a-66-01-33-83",
		"18-b8-87-e7-1f-06",
		"38-4e-73-e0-33-59",
	} {
		store.devices[deviceID] = &Device{
			id:               deviceID,
			heartbeatBuckets: make(map[time.Time]bool),
		}
	}
	store.mu.Unlock()
	
	// Simulate high-concurrency scenario with multiple simulators
	numGoroutines := 50 // Simulate 50 concurrent simulator requests
	numOperationsPerGoroutine := 1000
	
	var wg sync.WaitGroup
	
	// Test concurrent heartbeat operations across all devices
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			deviceIDs := []string{
				"60-6b-44-84-dc-64",
				"b4-45-52-a2-f1-3c", 
				"26-9a-66-01-33-83",
				"18-b8-87-e7-1f-06",
				"38-4e-73-e0-33-59",
			}
			
			for j := 0; j < numOperationsPerGoroutine; j++ {
				deviceID := deviceIDs[j%len(deviceIDs)]
				timestamp := time.Now().Add(time.Duration(j) * time.Second)
				
				// Global RWMutex: Devices map read/write
				// Per-device Mutex: Individual device updates
				// O(1) Operations: Direct map access for heartbeats
				store.AddHeartbeat(deviceID, timestamp)
			}
		}(i)
	}
	
	// Test concurrent upload operations across all devices
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			deviceIDs := []string{
				"60-6b-44-84-dc-64",
				"b4-45-52-a2-f1-3c", 
				"26-9a-66-01-33-83",
				"18-b8-87-e7-1f-06",
				"38-4e-73-e0-33-59",
			}
			
			for j := 0; j < numOperationsPerGoroutine; j++ {
				deviceID := deviceIDs[j%len(deviceIDs)]
				duration := time.Duration(j+1) * time.Millisecond
				
				// Minimize lock duration
				// No global contention during updates
				store.AddUploadTime(deviceID, duration)
			}
		}(i)
	}
	
	// Test concurrent read operations (uptime calculations)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			deviceIDs := []string{
				"60-6b-44-84-dc-64",
				"b4-45-52-a2-f1-3c", 
				"26-9a-66-01-33-83",
				"18-b8-87-e7-1f-06",
				"38-4e-73-e0-33-59",
			}
			
			for j := 0; j < numOperationsPerGoroutine; j++ {
				deviceID := deviceIDs[j%len(deviceIDs)]
				
				// Efficient minute-bucket operations
				uptime := store.CalculateUptime(deviceID)
				avgUpload := store.CalculateAverageUploadTime(deviceID)
				
				// Verify operations don't return invalid values
				if uptime < 0 || uptime > 100 {
					t.Errorf("Invalid uptime: %f", uptime)
				}
				if avgUpload < 0 {
					t.Errorf("Invalid average upload time: %v", avgUpload)
				}
			}
		}(i)
	}
	
	// Wait for all concurrent operations to complete
	wg.Wait()
	
	// Verify final state consistency
	for _, deviceID := range []string{
		"60-6b-44-84-dc-64",
		"b4-45-52-a2-f1-3c", 
		"26-9a-66-01-33-83",
		"18-b8-87-e7-1f-06",
		"38-4e-73-e0-33-59",
	} {
		uptime := store.CalculateUptime(deviceID)
		avgUpload := store.CalculateAverageUploadTime(deviceID)
		
		// Verify data integrity after concurrent operations
		if uptime < 0 || uptime > 100 {
			t.Errorf("Device %s: Invalid final uptime: %f", deviceID, uptime)
		}
		if avgUpload < 0 {
			t.Errorf("Device %s: Invalid final average upload time: %v", deviceID, avgUpload)
		}
	}
}

// TestConcurrencyLockingStrategy tests the specific locking strategy requirements
func TestConcurrencyLockingStrategy(t *testing.T) {
	store := New()
	
	// Add test device
	store.mu.Lock()
	store.devices["test-device"] = &Device{
		id:               "test-device",
		heartbeatBuckets: make(map[time.Time]bool),
	}
	store.mu.Unlock()
	
	// Test that operations complete without deadlocks
	var wg sync.WaitGroup
	numConcurrentOps := 100
	
	// Test Global RWMutex for device lookup
	for i := 0; i < numConcurrentOps; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// This should use Global RWMutex read lock
			store.AddHeartbeat("test-device", time.Now())
		}()
	}
	
	// Test Per-device Mutex for individual updates
	for i := 0; i < numConcurrentOps; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// This should use Per-device Mutex
			store.AddUploadTime("test-device", time.Millisecond*100)
		}()
	}
	
	// Test concurrent reads (should not block writes)
	for i := 0; i < numConcurrentOps; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// These should use minimal lock duration
			store.CalculateUptime("test-device")
			store.CalculateAverageUploadTime("test-device")
		}()
	}
	
	// All operations should complete without deadlock
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()
	
	select {
	case <-done:
		// Success: All operations completed
	case <-time.After(5 * time.Second):
		t.Fatal("Deadlock detected: Operations did not complete within 5 seconds")
	}
}

// TestPerformanceOptimizations tests the performance requirements
func TestPerformanceOptimizations(t *testing.T) {
	store := New()
	
	// Add test device
	store.mu.Lock()
	store.devices["perf-test"] = &Device{
		id:               "perf-test",
		heartbeatBuckets: make(map[time.Time]bool),
	}
	store.mu.Unlock()
	
	// Measure performance of concurrent operations
	numOperations := 10000
	start := time.Now()
	
	var wg sync.WaitGroup
	
	// Test O(1) heartbeat operations
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			timestamp := time.Now().Add(time.Duration(i) * time.Second)
			store.AddHeartbeat("perf-test", timestamp)
		}(i)
	}
	
	// Test O(1) upload operations
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			duration := time.Duration(i+1) * time.Millisecond
			store.AddUploadTime("perf-test", duration)
		}(i)
	}
	
	wg.Wait()
	elapsed := time.Since(start)
	
	// Performance should be reasonable for concurrent operations
	opsPerSecond := float64(numOperations*2) / elapsed.Seconds()
	
	t.Logf("Concurrent operations performance: %.0f ops/second", opsPerSecond)
	
	// Should handle at least 1000 ops/second for simulator requests
	if opsPerSecond < 1000 {
		t.Errorf("Performance too slow: %.0f ops/second (expected > 1000)", opsPerSecond)
	}
}
