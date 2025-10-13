package server

import (
    "encoding/json"
    "net/http"
    "strings"
    "time"

    "fleet-management-server/internal/store"
)

// Server represents the HTTP server
type Server struct {
    store  *store.Store
    sseHub *SSEHub
}

// handleStatsPost processes POST /api/v1/devices/{device_id}/stats
func (s *Server) handleStatsPost(w http.ResponseWriter, r *http.Request, deviceID string) {
    var req StatsRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.sendError(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Parse sent_at timestamp (RFC3339 format) - currently not used but validated
    if _, err := time.Parse(time.RFC3339, req.SentAt); err != nil {
        s.sendError(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Note: upload_time = int64 nanoseconds -> time.Duration
    s.store.AddUploadTime(deviceID, time.Duration(req.UploadTime))

    // Success: 204 No Content
    w.WriteHeader(http.StatusNoContent)
}

// New creates a new Server instance
func New(store *store.Store) *Server {
    s := &Server{
        store:  store,
        sseHub: NewSSEHub(),
    }
    // Periodic health broadcast for SSE subscribers (more frequent for quicker detection)
    go func() {
        ticker := time.NewTicker(2 * time.Second)
        defer ticker.Stop()
        for range ticker.C {
            s.broadcastHealth()
        }
    }()
    return s
}

// Close immediately tears down any server-managed long-lived connections (e.g., SSE)
func (s *Server) Close() {
    if s != nil && s.sseHub != nil {
        s.sseHub.CloseAll()
    }
}

// Router returns the HTTP router with all endpoints
func (s *Server) Router() http.Handler {
    mux := http.NewServeMux()

	// Milestone 5: HTTP API endpoints with exact OpenAPI contract compliance
	mux.HandleFunc("/api/v1/devices", s.handleDevicesList)
	mux.HandleFunc("/api/v1/devices/", s.handleDeviceEndpoints)

	    // Health check endpoint
    mux.HandleFunc("/health", s.handleHealth)

    // SSE events endpoint
    mux.HandleFunc("/events", s.handleEvents)

    return mux
}

// handleDeviceEndpoints handles all device-related endpoints with proper routing
func (s *Server) handleDeviceEndpoints(w http.ResponseWriter, r *http.Request) {
    // Parse URL path to extract device_id and endpoint
    path := r.URL.Path

    // Expected patterns:
    // POST /api/v1/devices/{device_id}/heartbeat
    // POST /api/v1/devices/{device_id}/stats
    // GET  /api/v1/devices/{device_id}/stats

    // Remove /api/v1/devices/ prefix
    prefix := "/api/v1/devices/"
    if len(path) < len(prefix) || path[:len(prefix)] != prefix {
        s.sendError(w, "Invalid endpoint", http.StatusNotFound)
        return
    }

    remaining := path[len(prefix):] // Remove "/api/v1/devices/"
    parts := strings.Split(remaining, "/")

    if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
        s.sendError(w, "Invalid endpoint format", http.StatusNotFound)
        return
    }

    deviceID := parts[0]
    endpoint := parts[1]

    // Validate device exists first
    if _, exists := s.store.GetDevice(deviceID); !exists {
        s.sendError(w, "Device not found", http.StatusNotFound)
        return
    }

    switch endpoint {
    case "heartbeat":
        if r.Method == http.MethodPost {
            s.handleHeartbeat(w, r, deviceID)
        } else {
            s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    case "stats":
        if r.Method == http.MethodPost {
            s.handleStatsPost(w, r, deviceID)
        } else if r.Method == http.MethodGet {
            s.handleStatsGet(w, r, deviceID)
        } else {
            s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    default:
        s.sendError(w, "Invalid endpoint", http.StatusNotFound)
    }
}

// HeartbeatRequest represents a heartbeat request for POST /api/v1/devices/{device_id}/heartbeat
type HeartbeatRequest struct {
    SentAt string `json:"sent_at"` // RFC3339 timestamp format
}

// StatsRequest represents a stats request for POST /api/v1/devices/{device_id}/stats
type StatsRequest struct {
    SentAt     string `json:"sent_at"`     // RFC3339 timestamp format
    UploadTime int64  `json:"upload_time"` // int64 nanoseconds -> time.Duration
}

// StatsResponse represents the response for GET /api/v1/devices/{device_id}/stats
type StatsResponse struct {
    AvgUploadTime string  `json:"avg_upload_time"` // time.Duration.String() format (e.g., "250ms")
    Uptime        float64 `json:"uptime"`          // float64 with 3 decimals (e.g., 98.999)
}

// ErrorResponse represents error responses
type ErrorResponse struct {
    Message string `json:"msg"`
}

// sendError sends a JSON error response
func (s *Server) sendError(w http.ResponseWriter, message string, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}

// handleHeartbeat processes POST /api/v1/devices/{device_id}/heartbeat
func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request, deviceID string) {
    var req HeartbeatRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.sendError(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Parse sent_at timestamp (RFC3339 format)
    sentAt, err := time.Parse(time.RFC3339, req.SentAt)
    if err != nil {
        s.sendError(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Add heartbeat to store
    s.store.AddHeartbeat(deviceID, sentAt)

    // Success: 204 No Content
    w.WriteHeader(http.StatusNoContent)

    // Broadcast device delta and health snapshot via SSE
    if s.sseHub != nil {
        s.sseHub.BroadcastJSON("devices:update", map[string]interface{}{
            "id":        deviceID,
            "last_seen": sentAt.Format(time.RFC3339),
            "status":    "online",
        })
        s.broadcastHealth()
    }
}

// handleStatsGet processes GET /api/v1/devices/{device_id}/stats
func (s *Server) handleStatsGet(w http.ResponseWriter, r *http.Request, deviceID string) {
	uptime := s.store.CalculateUptime(deviceID)
	avgUploadTime := s.store.CalculateAverageUploadTime(deviceID)

	// Check if device has no heartbeats AND no uploads
	_, heartbeatCount, exists := s.store.GetDeviceStats(deviceID)
	if !exists || (heartbeatCount == 0 && avgUploadTime == 0) {
		// No heartbeats AND no uploads: 204 No Content
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Has data: 200 + JSON response
	// Format: time.Duration.String() + float64 with 3 decimals
	response := StatsResponse{
		AvgUploadTime: avgUploadTime.String(), // time.Duration.String() format (e.g., "250ms")
		Uptime:        uptime,                 // float64 with 3 decimals (e.g., 98.999)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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
