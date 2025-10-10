package store

import (
	"testing"
	"time"
)

// Milestone 7: Unit Tests - Comprehensive device logic and uptime calculation testing

// TestUptimeCalculation tests critical uptime calculation scenarios
func TestUptimeCalculation(t *testing.T) {
	// Test Case 1: Single heartbeat → 100%
	t.Run("SingleHeartbeat100Percent", func(t *testing.T) {
		device := &Device{
			id:               "test-device-1",
			heartbeatBuckets: make(map[time.Time]bool),
		}
		
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		device.AddHeartbeat(baseTime)
		
		uptime := device.CalculateUptime()
		expected := 100.0
		
		if uptime != expected {
			t.Errorf("Single heartbeat: expected %.1f%%, got %.1f%%", expected, uptime)
		}
	})
	
	// Test Case 2: Multiple heartbeats → correct percentage
	t.Run("MultipleHeartbeatsCorrectPercentage", func(t *testing.T) {
		device := &Device{
			id:               "test-device-2",
			heartbeatBuckets: make(map[time.Time]bool),
		}
		
		// Heartbeats: 10:00, 10:02, 10:04 (3 buckets over 5 minutes = 60%)
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		device.AddHeartbeat(baseTime)                       // 10:00
		device.AddHeartbeat(baseTime.Add(2 * time.Minute)) // 10:02
		device.AddHeartbeat(baseTime.Add(4 * time.Minute)) // 10:04
		
		uptime := device.CalculateUptime()
		expected := 60.0 // (3 buckets / 5 minutes) * 100
		
		if uptime != expected {
			t.Errorf("Multiple heartbeats: expected %.1f%%, got %.1f%%", expected, uptime)
		}
	})
	
	// Test Case 3: Gaps in minutes → offline periods counted
	t.Run("GapsInMinutesOfflinePeriodsCount", func(t *testing.T) {
		device := &Device{
			id:               "test-device-3",
			heartbeatBuckets: make(map[time.Time]bool),
		}
		
		// Heartbeats: 10:00, 10:10 (2 buckets over 11 minutes = 18.18%)
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		device.AddHeartbeat(baseTime)                        // 10:00
		device.AddHeartbeat(baseTime.Add(10 * time.Minute)) // 10:10
		
		uptime := device.CalculateUptime()
		expected := (2.0 / 11.0) * 100 // 18.18%
		
		if uptime < expected-0.1 || uptime > expected+0.1 {
			t.Errorf("Gaps in minutes: expected %.2f%%, got %.2f%%", expected, uptime)
		}
	})
	
	// Test Case 4: Out-of-order timestamps → use actual values
	t.Run("OutOfOrderTimestampsUseActualValues", func(t *testing.T) {
		device := &Device{
			id:               "test-device-4",
			heartbeatBuckets: make(map[time.Time]bool),
		}
		
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		
		// Add heartbeats out of order
		device.AddHeartbeat(baseTime.Add(4 * time.Minute)) // 10:04 (added first)
		device.AddHeartbeat(baseTime)                       // 10:00 (added second)
		device.AddHeartbeat(baseTime.Add(2 * time.Minute)) // 10:02 (added third)
		
		uptime := device.CalculateUptime()
		expected := 60.0 // Should still be (3 buckets / 5 minutes) * 100
		
		if uptime != expected {
			t.Errorf("Out-of-order timestamps: expected %.1f%%, got %.1f%%", expected, uptime)
		}
		
		// Verify first/last heartbeat tracking works correctly
		expectedFirst := baseTime
		expectedLast := baseTime.Add(4 * time.Minute)
		
		if !device.firstHeartbeat.Equal(expectedFirst) {
			t.Errorf("First heartbeat: expected %v, got %v", expectedFirst, device.firstHeartbeat)
		}
		if !device.lastHeartbeat.Equal(expectedLast) {
			t.Errorf("Last heartbeat: expected %v, got %v", expectedLast, device.lastHeartbeat)
		}
	})
	
	// Test Case 5: Consecutive minutes → 100%
	t.Run("ConsecutiveMinutes100Percent", func(t *testing.T) {
		device := &Device{
			id:               "test-device-5",
			heartbeatBuckets: make(map[time.Time]bool),
		}
		
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		
		// Add heartbeats for consecutive minutes: 10:00, 10:01, 10:02, 10:03, 10:04
		for i := 0; i < 5; i++ {
			device.AddHeartbeat(baseTime.Add(time.Duration(i) * time.Minute))
		}
		
		uptime := device.CalculateUptime()
		expected := 100.0 // (5 buckets / 5 minutes) * 100
		
		if uptime != expected {
			t.Errorf("Consecutive minutes: expected %.1f%%, got %.1f%%", expected, uptime)
		}
	})
	
	// Test Case 6: Same minute multiple heartbeats → single bucket
	t.Run("SameMinuteMultipleHeartbeatsSingleBucket", func(t *testing.T) {
		device := &Device{
			id:               "test-device-6",
			heartbeatBuckets: make(map[time.Time]bool),
		}
		
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		
		// Add multiple heartbeats in the same minute
		device.AddHeartbeat(baseTime)                       // 10:00:00
		device.AddHeartbeat(baseTime.Add(15 * time.Second)) // 10:00:15
		device.AddHeartbeat(baseTime.Add(30 * time.Second)) // 10:00:30
		device.AddHeartbeat(baseTime.Add(45 * time.Second)) // 10:00:45
		
		uptime := device.CalculateUptime()
		expected := 100.0 // Single heartbeat = 100%
		
		if uptime != expected {
			t.Errorf("Same minute multiple heartbeats: expected %.1f%%, got %.1f%%", expected, uptime)
		}
		
		// Verify only one bucket was created
		if len(device.heartbeatBuckets) != 1 {
			t.Errorf("Expected 1 heartbeat bucket, got %d", len(device.heartbeatBuckets))
		}
	})
}

// TestUploadAveraging tests comprehensive upload averaging scenarios
func TestUploadAveraging(t *testing.T) {
	// Test Case 1: Zero uploads + heartbeats → "0s"
	t.Run("ZeroUploadsHeartbeats0s", func(t *testing.T) {
		device := &Device{
			id:               "test-device-upload-1",
			heartbeatBuckets: make(map[time.Time]bool),
		}
		
		// Add heartbeats but no uploads
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		device.AddHeartbeat(baseTime)
		device.AddHeartbeat(baseTime.Add(1 * time.Minute))
		
		avgUpload := device.CalculateAvgUploadTime()
		expected := time.Duration(0)
		
		if avgUpload != expected {
			t.Errorf("Zero uploads: expected %v, got %v", expected, avgUpload)
		}
	})
	
	// Test Case 2: Multiple uploads → correct average
	t.Run("MultipleUploadsCorrectAverage", func(t *testing.T) {
		device := &Device{
			id:               "test-device-upload-2",
			heartbeatBuckets: make(map[time.Time]bool),
		}
		
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		
		// Add uploads: 100ms, 200ms, 300ms → average = 200ms
		device.AddUploadStat(baseTime, (100 * time.Millisecond).Nanoseconds())
		device.AddUploadStat(baseTime.Add(1*time.Minute), (200 * time.Millisecond).Nanoseconds())
		device.AddUploadStat(baseTime.Add(2*time.Minute), (300 * time.Millisecond).Nanoseconds())
		
		avgUpload := device.CalculateAvgUploadTime()
		expected := 200 * time.Millisecond
		
		if avgUpload != expected {
			t.Errorf("Multiple uploads average: expected %v, got %v", expected, avgUpload)
		}
	})
	
	// Test Case 3: Nanosecond precision handling
	t.Run("NanosecondPrecisionHandling", func(t *testing.T) {
		device := &Device{
			id:               "test-device-upload-3",
			heartbeatBuckets: make(map[time.Time]bool),
		}
		
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		
		// Add uploads with nanosecond precision: 1.5ms, 2.5ms → average = 2ms
		device.AddUploadStat(baseTime, (1500000)) // 1.5ms in nanoseconds
		device.AddUploadStat(baseTime.Add(1*time.Minute), (2500000)) // 2.5ms in nanoseconds
		
		avgUpload := device.CalculateAvgUploadTime()
		expected := 2 * time.Millisecond // (1.5ms + 2.5ms) / 2 = 2ms
		
		if avgUpload != expected {
			t.Errorf("Nanosecond precision: expected %v, got %v", expected, avgUpload)
		}
	})
	
	// Test Case 4: Single upload → same value returned
	t.Run("SingleUploadSameValue", func(t *testing.T) {
		device := &Device{
			id:               "test-device-upload-4",
			heartbeatBuckets: make(map[time.Time]bool),
		}
		
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		uploadTime := 250 * time.Millisecond
		
		device.AddUploadStat(baseTime, uploadTime.Nanoseconds())
		
		avgUpload := device.CalculateAvgUploadTime()
		
		if avgUpload != uploadTime {
			t.Errorf("Single upload: expected %v, got %v", uploadTime, avgUpload)
		}
	})
	
	// Test Case 5: Large number of uploads → correct average
	t.Run("LargeNumberUploadsCorrectAverage", func(t *testing.T) {
		device := &Device{
			id:               "test-device-upload-5",
			heartbeatBuckets: make(map[time.Time]bool),
		}
		
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		
		// Add 1000 uploads of 100ms each → average should be 100ms
		numUploads := 1000
		uploadTime := 100 * time.Millisecond
		
		for i := 0; i < numUploads; i++ {
			device.AddUploadStat(baseTime.Add(time.Duration(i)*time.Second), uploadTime.Nanoseconds())
		}
		
		avgUpload := device.CalculateAvgUploadTime()
		
		if avgUpload != uploadTime {
			t.Errorf("Large number uploads: expected %v, got %v", uploadTime, avgUpload)
		}
		
		// Verify upload count
		if device.uploadCount != int64(numUploads) {
			t.Errorf("Upload count: expected %d, got %d", numUploads, device.uploadCount)
		}
	})
}

// TestEdgeCasesAndBoundaryConditions tests edge cases for comprehensive coverage
func TestEdgeCasesAndBoundaryConditions(t *testing.T) {
	// Test Case 1: No heartbeats → 0% uptime
	t.Run("NoHeartbeats0Percent", func(t *testing.T) {
		device := &Device{
			id:               "test-edge-1",
			heartbeatBuckets: make(map[time.Time]bool),
		}
		
		uptime := device.CalculateUptime()
		expected := 0.0
		
		if uptime != expected {
			t.Errorf("No heartbeats: expected %.1f%%, got %.1f%%", expected, uptime)
		}
	})
	
	// Test Case 2: Heartbeats spanning multiple hours
	t.Run("HeartbeatsSpanningMultipleHours", func(t *testing.T) {
		device := &Device{
			id:               "test-edge-2",
			heartbeatBuckets: make(map[time.Time]bool),
		}
		
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		
		// Heartbeats: 10:00, 11:00, 12:00 (3 buckets over 121 minutes)
		device.AddHeartbeat(baseTime)                       // 10:00
		device.AddHeartbeat(baseTime.Add(1 * time.Hour))   // 11:00
		device.AddHeartbeat(baseTime.Add(2 * time.Hour))   // 12:00
		
		uptime := device.CalculateUptime()
		expected := (3.0 / 121.0) * 100 // 3 buckets over 121 minutes (10:00 to 12:00)
		
		if uptime < expected-0.1 || uptime > expected+0.1 {
			t.Errorf("Multiple hours: expected %.2f%%, got %.2f%%", expected, uptime)
		}
	})
	
	// Test Case 3: Very small upload times
	t.Run("VerySmallUploadTimes", func(t *testing.T) {
		device := &Device{
			id:               "test-edge-3",
			heartbeatBuckets: make(map[time.Time]bool),
		}
		
		baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
		
		// Add very small upload times: 1ns, 2ns, 3ns → average = 2ns
		device.AddUploadStat(baseTime, 1) // 1 nanosecond
		device.AddUploadStat(baseTime.Add(1*time.Minute), 2) // 2 nanoseconds
		device.AddUploadStat(baseTime.Add(2*time.Minute), 3) // 3 nanoseconds
		
		avgUpload := device.CalculateAvgUploadTime()
		expected := 2 * time.Nanosecond
		
		if avgUpload != expected {
			t.Errorf("Very small uploads: expected %v, got %v", expected, avgUpload)
		}
	})
}
