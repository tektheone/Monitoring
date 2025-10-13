package store

import (
	"testing"
	"time"
)

func TestDeviceHeartbeatBuckets(t *testing.T) {
	device := &Device{
		id:               "test-device",
		heartbeatBuckets: make(map[time.Time]bool),
	}

	// Test adding heartbeats in different minutes using fixed time
	baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	device.AddHeartbeat(baseTime)                           // 10:00:00 -> bucket 10:00
	device.AddHeartbeat(baseTime.Add(30 * time.Second))     // 10:00:30 -> bucket 10:00 (same minute)
	device.AddHeartbeat(baseTime.Add(90 * time.Second))     // 10:01:30 -> bucket 10:01 (different minute)

	// Should have 2 buckets (2 different minutes: 10:00 and 10:01)
	if len(device.heartbeatBuckets) != 2 {
		t.Errorf("Expected 2 heartbeat buckets, got %d", len(device.heartbeatBuckets))
	}

	// Test uptime calculation
	uptime := device.CalculateUptime()
	if uptime <= 0 {
		t.Errorf("Expected positive uptime, got %f", uptime)
	}
}

func TestDeviceUploadStats(t *testing.T) {
	device := &Device{
		id:               "test-device",
		heartbeatBuckets: make(map[time.Time]bool),
	}

	// Add upload statistics
	device.AddUploadStat(time.Now(), 1000000000) // 1 second in nanoseconds
	device.AddUploadStat(time.Now(), 2000000000) // 2 seconds in nanoseconds

	// Test average calculation
	avgUpload := device.CalculateAvgUploadTime()
	expected := 1500 * time.Millisecond // Average of 1s and 2s
	if avgUpload != expected {
		t.Errorf("Expected average upload time %v, got %v", expected, avgUpload)
	}
}

func TestExactUptimeFormula(t *testing.T) {
	device := &Device{
		id:               "test-device",
		heartbeatBuckets: make(map[time.Time]bool),
	}

	// Test the exact example from Milestone 4 specification
	// Heartbeats: 10:00:15, 10:02:30, 10:04:45
	// Buckets: 3 (10:00, 10:02, 10:04)
	// Range: 5 minutes (10:00 through 10:04)
	// Uptime: (3/5) * 100 = 60.0%
	
	baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	
	device.AddHeartbeat(baseTime.Add(15 * time.Second))     // 10:00:15 -> bucket 10:00
	device.AddHeartbeat(baseTime.Add(2*time.Minute + 30*time.Second)) // 10:02:30 -> bucket 10:02
	device.AddHeartbeat(baseTime.Add(4*time.Minute + 45*time.Second)) // 10:04:45 -> bucket 10:04

	uptime := device.CalculateUptime()
	expected := 60.0
	if uptime != expected {
		t.Errorf("Expected %.1f%% uptime, got %.1f%%", expected, uptime)
	}
	
	// Verify bucket count
	if len(device.heartbeatBuckets) != 3 {
		t.Errorf("Expected 3 heartbeat buckets, got %d", len(device.heartbeatBuckets))
	}
}

func TestUptimeAlgorithmEdgeCases(t *testing.T) {
	// Test case 1: No heartbeats
	device1 := &Device{
		id:               "test-device-1",
		heartbeatBuckets: make(map[time.Time]bool),
	}
	
	uptime1 := device1.CalculateUptime()
	if uptime1 != 0.0 {
		t.Errorf("Expected 0%% uptime for no heartbeats, got %.1f%%", uptime1)
	}
	
	// Test case 2: Single heartbeat
	device2 := &Device{
		id:               "test-device-2",
		heartbeatBuckets: make(map[time.Time]bool),
	}
	
	baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	device2.AddHeartbeat(baseTime)
	
	uptime2 := device2.CalculateUptime()
	if uptime2 != 100.0 {
		t.Errorf("Expected 100%% uptime for single heartbeat, got %.1f%%", uptime2)
	}
	
	// Test case 3: All heartbeats in same minute
	device3 := &Device{
		id:               "test-device-3",
		heartbeatBuckets: make(map[time.Time]bool),
	}
	
	device3.AddHeartbeat(baseTime.Add(10 * time.Second))
	device3.AddHeartbeat(baseTime.Add(20 * time.Second))
	device3.AddHeartbeat(baseTime.Add(30 * time.Second))
	
	uptime3 := device3.CalculateUptime()
	if uptime3 != 100.0 {
		t.Errorf("Expected 100%% uptime for same minute heartbeats, got %.1f%%", uptime3)
	}
}

func TestUptimeAlgorithmConsecutiveMinutes(t *testing.T) {
	device := &Device{
		id:               "test-device",
		heartbeatBuckets: make(map[time.Time]bool),
	}

	// Test consecutive minutes: 10:00, 10:01, 10:02
	// 3 buckets over 3 minutes = 100%
	baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	
	device.AddHeartbeat(baseTime)                       // 10:00
	device.AddHeartbeat(baseTime.Add(1 * time.Minute)) // 10:01
	device.AddHeartbeat(baseTime.Add(2 * time.Minute)) // 10:02

	uptime := device.CalculateUptime()
	expected := 100.0
	if uptime != expected {
		t.Errorf("Expected %.1f%% uptime for consecutive minutes, got %.1f%%", expected, uptime)
	}
}

func TestUptimeAlgorithmSparseHeartbeats(t *testing.T) {
	device := &Device{
		id:               "test-device",
		heartbeatBuckets: make(map[time.Time]bool),
	}

	// Test sparse heartbeats: 10:00, 10:05, 10:10
	// 3 buckets over 11 minutes (10:00 through 10:10) = 27.27%
	baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	
	device.AddHeartbeat(baseTime)                        // 10:00
	device.AddHeartbeat(baseTime.Add(5 * time.Minute))  // 10:05
	device.AddHeartbeat(baseTime.Add(10 * time.Minute)) // 10:10

	uptime := device.CalculateUptime()
	expected := (3.0 / 11.0) * 100 // 27.27%
	if uptime < expected-0.1 || uptime > expected+0.1 {
		t.Errorf("Expected %.2f%% uptime for sparse heartbeats, got %.2f%%", expected, uptime)
	}
}
