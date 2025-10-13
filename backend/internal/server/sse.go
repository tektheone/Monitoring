package server

import (
	"bufio"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type sseEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// CloseAll force-disconnects all SSE clients
func (h *SSEHub) CloseAll() {
    h.mu.Lock()
    for ch := range h.clients {
        delete(h.clients, ch)
        close(ch)
    }
    h.mu.Unlock()
}

type SSEHub struct {
    mu      sync.RWMutex
    clients map[chan []byte]struct{}
}

func NewSSEHub() *SSEHub {
	return &SSEHub{clients: make(map[chan []byte]struct{})}
}
func (h *SSEHub) Subscribe() (ch chan []byte, closeFn func()) {
	ch = make(chan []byte, 16)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		if _, ok := h.clients[ch]; ok {
			delete(h.clients, ch)
			close(ch)
		}
		h.mu.Unlock()
	}
}

func (h *SSEHub) BroadcastJSON(eventType string, data interface{}) {
	msg, _ := json.Marshal(sseEvent{Type: eventType, Data: data})
	// Wrap for SSE: only using default event field with data
	payload := append([]byte("data: "), msg...)
	payload = append(payload, []byte("\n\n")...)

	h.mu.RLock()
	for ch := range h.clients {
		select {
		case ch <- payload:
		default:
			// drop message if client is slow
		}
	}
	h.mu.RUnlock()
}

// handleEvents serves the /events SSE endpoint
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
        return
    }

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch, closeFn := s.sseHub.Subscribe()
	defer closeFn()

    bw := bufio.NewWriter(w)
    bw.WriteString(": open\n\n")
    bw.Flush()
    flusher.Flush()
    s.broadcastHealth()
    // Also push current devices list so UI has immediate data
    if s.sseHub != nil {
        summaries := s.buildDeviceSummaries()
        s.sseHub.BroadcastJSON("devices:update", map[string]interface{}{
            "devices": summaries,
        })
    }

	// Ping comments to keep connection alive on proxies (more frequent for quick detection)
	pingTicker := time.NewTicker(2 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			if _, err := bw.Write(msg); err != nil {
				return
			}
			bw.Flush()
			flusher.Flush()
		case <-pingTicker.C:
			bw.WriteString(": ping\n\n")
			bw.Flush()
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// broadcastHealth snapshots current health and emits health:update
func (s *Server) broadcastHealth() {
	count := s.store.DeviceCount()
	s.sseHub.BroadcastJSON("health:update", map[string]interface{}{
		"status":       "healthy",
		"timestamp":    time.Now().Unix(),
		"device_count": count,
	})
}
