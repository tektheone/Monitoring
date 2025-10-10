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

	// Test adding heartbeats in different minutes
	now := time.Now()
	device.AddHeartbeat(now)
	device.AddHeartbeat(now.Add(30 * time.Second)) // Same minute
	device.AddHeartbeat(now.Add(90 * time.Second)) // Different minute

	// Should have 2 buckets (2 different minutes)
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

	// Create a scenario where we know the expected uptime
	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	
	// Add heartbeats at minutes 0, 1, 3 (3 buckets over 3 minutes = 100% uptime)
	device.AddHeartbeat(baseTime)                           // minute 0
	device.AddHeartbeat(baseTime.Add(1 * time.Minute))     // minute 1
	device.AddHeartbeat(baseTime.Add(3 * time.Minute))     // minute 3

	// 3 heartbeat buckets over 3 minutes (from minute 0 to minute 3) = 100%
	uptime := device.CalculateUptime()
	if uptime != 100.0 {
		t.Errorf("Expected 100%% uptime, got %f%%", uptime)
	}
}
