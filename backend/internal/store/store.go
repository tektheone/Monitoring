package store

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Device represents a device in the fleet
type Device struct {
	ID          string
	LastSeen    time.Time
	Heartbeats  []time.Time
	UploadTimes []time.Duration
	mu          sync.RWMutex
}

// Store manages the device data with thread-safe operations
type Store struct {
	devices map[string]*Device
	mu      sync.RWMutex
}

// New creates a new Store instance
func New() *Store {
	return &Store{
		devices: make(map[string]*Device),
	}
}

// LoadFromCSV loads devices from a CSV file
func (s *Store) LoadFromCSV(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
	// Read header
	_, err = reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

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
			s.devices[deviceID] = &Device{
				ID:          deviceID,
				Heartbeats:  make([]time.Time, 0),
				UploadTimes: make([]time.Duration, 0),
			}
		}
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

// AddHeartbeat adds a heartbeat timestamp for a device
func (s *Store) AddHeartbeat(deviceID string, timestamp time.Time) {
	s.mu.RLock()
	device, exists := s.devices[deviceID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	device.mu.Lock()
	defer device.mu.Unlock()
	
	device.Heartbeats = append(device.Heartbeats, timestamp)
	device.LastSeen = timestamp
}

// AddUploadTime adds an upload duration for a device
func (s *Store) AddUploadTime(deviceID string, duration time.Duration) {
	s.mu.RLock()
	device, exists := s.devices[deviceID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	device.mu.Lock()
	defer device.mu.Unlock()
	
	device.UploadTimes = append(device.UploadTimes, duration)
}

// CalculateUptime calculates uptime percentage for a device using the exact formula
// uptime = (sumHeartbeats / numMinutesBetweenFirstAndLastHeartbeat) * 100
func (s *Store) CalculateUptime(deviceID string) float64 {
	s.mu.RLock()
	device, exists := s.devices[deviceID]
	s.mu.RUnlock()

	if !exists {
		return 0.0
	}

	device.mu.RLock()
	defer device.mu.RUnlock()

	if len(device.Heartbeats) == 0 {
		return 0.0
	}

	if len(device.Heartbeats) == 1 {
		return 100.0 // Single heartbeat = 100% uptime
	}

	// Find first and last heartbeat times
	firstHeartbeat := device.Heartbeats[0]
	lastHeartbeat := device.Heartbeats[0]

	for _, hb := range device.Heartbeats {
		if hb.Before(firstHeartbeat) {
			firstHeartbeat = hb
		}
		if hb.After(lastHeartbeat) {
			lastHeartbeat = hb
		}
	}

	// Calculate minutes between first and last heartbeat
	duration := lastHeartbeat.Sub(firstHeartbeat)
	numMinutes := duration.Minutes()

	if numMinutes == 0 {
		return 100.0 // All heartbeats in same minute
	}

	// Apply exact uptime formula
	sumHeartbeats := float64(len(device.Heartbeats))
	uptime := (sumHeartbeats / numMinutes) * 100

	// Cap at 100%
	if uptime > 100.0 {
		uptime = 100.0
	}

	return uptime
}

// CalculateAverageUploadTime calculates average upload time for a device
func (s *Store) CalculateAverageUploadTime(deviceID string) time.Duration {
	s.mu.RLock()
	device, exists := s.devices[deviceID]
	s.mu.RUnlock()

	if !exists {
		return 0
	}

	device.mu.RLock()
	defer device.mu.RUnlock()

	if len(device.UploadTimes) == 0 {
		return 0
	}

	var total time.Duration
	for _, uploadTime := range device.UploadTimes {
		total += uploadTime
	}

	return total / time.Duration(len(device.UploadTimes))
}

// GetDeviceStats returns device statistics safely
func (s *Store) GetDeviceStats(deviceID string) (lastSeen time.Time, heartbeatCount int, exists bool) {
	s.mu.RLock()
	device, exists := s.devices[deviceID]
	s.mu.RUnlock()

	if !exists {
		return time.Time{}, 0, false
	}

	device.mu.RLock()
	defer device.mu.RUnlock()

	return device.LastSeen, len(device.Heartbeats), true
}
