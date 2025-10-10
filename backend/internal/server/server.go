package server

import (
	"encoding/json"
	"net/http"
	"time"

	"fleet-management-server/internal/store"
)

// Server represents the HTTP server
type Server struct {
	store *store.Store
}

// New creates a new Server instance
func New(store *store.Store) *Server {
	return &Server{
		store: store,
	}
}

// Router returns the HTTP router with all endpoints
func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// API endpoints with base path /api/v1
	mux.HandleFunc("/api/v1/heartbeat", s.handleHeartbeat)
	mux.HandleFunc("/api/v1/upload", s.handleUpload)
	mux.HandleFunc("/api/v1/devices", s.handleDevices)
	mux.HandleFunc("/api/v1/devices/", s.handleDeviceDetail)

	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)

	return mux
}

// HeartbeatRequest represents a heartbeat request
type HeartbeatRequest struct {
	DeviceID  string `json:"device_id"`
	Timestamp int64  `json:"timestamp"`
}

// UploadRequest represents an upload request
type UploadRequest struct {
	DeviceID     string `json:"device_id"`
	UploadTimeMs int64  `json:"upload_time_ms"`
}

// DeviceResponse represents a device in API responses
type DeviceResponse struct {
	ID                string  `json:"id"`
	LastSeen          int64   `json:"last_seen"`
	UptimePercentage  float64 `json:"uptime_percentage"`
	AvgUploadTimeMs   int64   `json:"avg_upload_time_ms"`
	HeartbeatCount    int     `json:"heartbeat_count"`
}

// DevicesResponse represents the devices list response
type DevicesResponse struct {
	Devices []DeviceResponse `json:"devices"`
	Count   int              `json:"count"`
}

// handleHeartbeat processes heartbeat requests
func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.DeviceID == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}

	// Convert timestamp to time.Time
	timestamp := time.Unix(req.Timestamp, 0)
	if req.Timestamp == 0 {
		timestamp = time.Now()
	}

	// Add heartbeat to store
	s.store.AddHeartbeat(req.DeviceID, timestamp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleUpload processes upload time requests
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.DeviceID == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}

	if req.UploadTimeMs <= 0 {
		http.Error(w, "upload_time_ms must be positive", http.StatusBadRequest)
		return
	}

	// Convert milliseconds to duration
	uploadDuration := time.Duration(req.UploadTimeMs) * time.Millisecond

	// Add upload time to store
	s.store.AddUploadTime(req.DeviceID, uploadDuration)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleDevices returns all devices with their metrics
func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	devices := s.store.GetAllDevices()
	deviceResponses := make([]DeviceResponse, 0, len(devices))

	for deviceID := range devices {
		uptime := s.store.CalculateUptime(deviceID)
		avgUploadTime := s.store.CalculateAverageUploadTime(deviceID)
		lastSeen, heartbeatCount, exists := s.store.GetDeviceStats(deviceID)

		if !exists {
			continue
		}

		deviceResp := DeviceResponse{
			ID:                deviceID,
			LastSeen:          lastSeen.Unix(),
			UptimePercentage:  uptime,
			AvgUploadTimeMs:   avgUploadTime.Nanoseconds() / 1000000, // Convert to milliseconds
			HeartbeatCount:    heartbeatCount,
		}

		deviceResponses = append(deviceResponses, deviceResp)
	}

	response := DevicesResponse{
		Devices: deviceResponses,
		Count:   len(deviceResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDeviceDetail returns details for a specific device
func (s *Server) handleDeviceDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract device ID from URL path
	deviceID := r.URL.Path[len("/api/v1/devices/"):]
	if deviceID == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}

	lastSeen, heartbeatCount, exists := s.store.GetDeviceStats(deviceID)
	if !exists {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	uptime := s.store.CalculateUptime(deviceID)
	avgUploadTime := s.store.CalculateAverageUploadTime(deviceID)

	deviceResp := DeviceResponse{
		ID:                deviceID,
		LastSeen:          lastSeen.Unix(),
		UptimePercentage:  uptime,
		AvgUploadTimeMs:   avgUploadTime.Nanoseconds() / 1000000, // Convert to milliseconds
		HeartbeatCount:    heartbeatCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deviceResp)
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	deviceCount := s.store.DeviceCount()
	
	response := map[string]interface{}{
		"status":       "healthy",
		"timestamp":    time.Now().Unix(),
		"device_count": deviceCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
