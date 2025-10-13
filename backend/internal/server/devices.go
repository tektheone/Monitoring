package server

import (
	"encoding/json"
	"net/http"
	"time"
)

// DeviceSummary represents a concise view of a device for GET /api/v1/devices
type DeviceSummary struct {
	ID       string `json:"id"`
	LastSeen string `json:"last_seen,omitempty"`
	Status   string `json:"status,omitempty"`
}

// handleDevicesList processes GET /api/v1/devices (list all devices)
func (s *Server) handleDevicesList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	devices := s.store.GetAllDevices()
	now := time.Now()

	out := make([]DeviceSummary, 0, len(devices))
	for id := range devices {
		lastSeen, _, exists := s.store.GetDeviceStats(id)
		status := "offline"
		if exists && !lastSeen.IsZero() {
			// Convert heartbeat time to local timezone to match server time
			lastSeenLocal := lastSeen.In(now.Location())
			timeDiff := now.Sub(lastSeenLocal)
			if timeDiff < 2*time.Minute {
				status = "online"
			}
		}

		var lastSeenStr string
		if !lastSeen.IsZero() {
			lastSeenStr = lastSeen.Format(time.RFC3339)
		}

		out = append(out, DeviceSummary{
			ID:       id,
			LastSeen: lastSeenStr,
			Status:   status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
