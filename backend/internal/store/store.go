package store

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Device represents a device in the fleet with minute-bucket heartbeat tracking
type Device struct {
	mu                sync.Mutex
	id                string
	heartbeatBuckets  map[time.Time]bool // minute buckets (truncated to minute)
	firstHeartbeat    time.Time          // earliest heartbeat timestamp
	lastHeartbeat     time.Time          // latest heartbeat timestamp
	uploadCount       int64              // total samples
	uploadSum         time.Duration      // sum of durations
}

// Store manages the device data with concurrent-safe operations
type Store struct {
	mu      sync.RWMutex
	devices map[string]*Device
}

// New creates a new Store instance
func New() *Store {
	return &Store{
		devices: make(map[string]*Device),
	}
}

// LoadFromCSV loads exactly 5 devices from a CSV file with concurrent-safe operations
func (s *Store) LoadFromCSV(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
	// Read header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}
	
	// Validate header
	if len(header) == 0 || header[0] != "device_id" {
		return fmt.Errorf("invalid CSV header: expected 'device_id', got %v", header)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear existing devices
	s.devices = make(map[string]*Device)
	
	// Expected device IDs from the specification
	expectedDevices := map[string]bool{
		"60-6b-44-84-dc-64": false,
		"b4-45-52-a2-f1-3c": false,
		"26-9a-66-01-33-83": false,
		"18-b8-87-e7-1f-06": false,
		"38-4e-73-e0-33-59": false,
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read CSV record: %w", err)
		}

		if len(record) > 0 && record[0] != "" {
			deviceID := record[0]
			
			// Validate device ID is expected
			if _, expected := expectedDevices[deviceID]; !expected {
				return fmt.Errorf("unexpected device ID in CSV: %s", deviceID)
			}
			
			// Mark as found
			expectedDevices[deviceID] = true
			
			// Create device with concurrent-safe initialization
			s.devices[deviceID] = &Device{
				id:               deviceID,
				heartbeatBuckets: make(map[time.Time]bool),
			}
		}
	}
	
	// Verify all 5 devices were loaded
	loadedCount := 0
	var missingDevices []string
	for deviceID, found := range expectedDevices {
		if found {
			loadedCount++
		} else {
			missingDevices = append(missingDevices, deviceID)
		}
	}
	
	if loadedCount != 5 {
		return fmt.Errorf("expected exactly 5 devices, loaded %d. Missing: %v", loadedCount, missingDevices)
	}

	return nil
}

// DeviceCount returns the number of devices in the store
func (s *Store) DeviceCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.devices)
}

// GetDevice returns a device by ID
func (s *Store) GetDevice(deviceID string) (*Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	device, exists := s.devices[deviceID]
	return device, exists
}

// GetAllDevices returns all devices
func (s *Store) GetAllDevices() map[string]*Device {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Return a copy to prevent external modification
	devices := make(map[string]*Device)
	for id, device := range s.devices {
		devices[id] = device
	}
	return devices
}

// AddHeartbeat adds a heartbeat with O(1) bucket insertion
func (d *Device) AddHeartbeat(sentAt time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	// Truncate to minute for bucket key
	minuteBucket := sentAt.Truncate(time.Minute)
	d.heartbeatBuckets[minuteBucket] = true
	
	// Update first/last heartbeat tracking
	if d.firstHeartbeat.IsZero() || sentAt.Before(d.firstHeartbeat) {
		d.firstHeartbeat = sentAt
	}
	if d.lastHeartbeat.IsZero() || sentAt.After(d.lastHeartbeat) {
		d.lastHeartbeat = sentAt
	}
}

// AddUploadStat adds upload statistics with O(1) accumulation
func (d *Device) AddUploadStat(sentAt time.Time, uploadNanos int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.uploadCount++
	d.uploadSum += time.Duration(uploadNanos)
}

// CalculateUptime returns uptime percentage using exact formula implementation
func (d *Device) CalculateUptime() float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if len(d.heartbeatBuckets) == 0 {
		return 0.0
	}
	
	if d.firstHeartbeat.IsZero() || d.lastHeartbeat.IsZero() {
		return 0.0
	}
	
	// Calculate minutes between first and last heartbeat
	duration := d.lastHeartbeat.Sub(d.firstHeartbeat)
	numMinutesBetweenFirstAndLastHeartbeat := duration.Minutes()
	
	if numMinutesBetweenFirstAndLastHeartbeat == 0 {
		return 100.0 // All heartbeats in same minute
	}
	
	// Exact formula: uptime = (sumHeartbeats / numMinutesBetweenFirstAndLastHeartbeat) * 100
	sumHeartbeats := float64(len(d.heartbeatBuckets))
	uptime := (sumHeartbeats / numMinutesBetweenFirstAndLastHeartbeat) * 100
	
	// Cap at 100%
	if uptime > 100.0 {
		uptime = 100.0
	}
	
	return uptime
}

// CalculateAvgUploadTime returns simple average of upload times
func (d *Device) CalculateAvgUploadTime() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if d.uploadCount == 0 {
		return 0
	}
	
	return d.uploadSum / time.Duration(d.uploadCount)
}

// Store wrapper methods for backward compatibility
func (s *Store) AddHeartbeat(deviceID string, timestamp time.Time) {
	s.mu.RLock()
	device, exists := s.devices[deviceID]
	s.mu.RUnlock()

	if exists {
		device.AddHeartbeat(timestamp)
	}
}

func (s *Store) AddUploadTime(deviceID string, duration time.Duration) {
	s.mu.RLock()
	device, exists := s.devices[deviceID]
	s.mu.RUnlock()

	if exists {
		device.AddUploadStat(time.Now(), duration.Nanoseconds())
	}
}

func (s *Store) CalculateUptime(deviceID string) float64 {
	s.mu.RLock()
	device, exists := s.devices[deviceID]
	s.mu.RUnlock()

	if !exists {
		return 0.0
	}

	return device.CalculateUptime()
}

func (s *Store) CalculateAverageUploadTime(deviceID string) time.Duration {
	s.mu.RLock()
	device, exists := s.devices[deviceID]
	s.mu.RUnlock()

	if !exists {
		return 0
	}

	return device.CalculateAvgUploadTime()
}

// GetDeviceStats returns device statistics safely
func (s *Store) GetDeviceStats(deviceID string) (lastSeen time.Time, heartbeatCount int, exists bool) {
	s.mu.RLock()
	device, exists := s.devices[deviceID]
	s.mu.RUnlock()

	if !exists {
		return time.Time{}, 0, false
	}

	device.mu.Lock()
	defer device.mu.Unlock()

	return device.lastHeartbeat, len(device.heartbeatBuckets), true
}
