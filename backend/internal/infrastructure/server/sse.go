package server

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Hub broadcasts a "bookings changed" ping to all connected SSE clients.
type Hub struct {
	mu      sync.RWMutex
	clients map[chan struct{}]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[chan struct{}]struct{})}
}

func (h *Hub) subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) unsubscribe(ch chan struct{}) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
}

// Notify wakes up all connected SSE clients.
func (h *Hub) Notify() {
	h.mu.RLock()
	for ch := range h.clients {
		select {
		case ch <- struct{}{}:
		default: // already has a pending signal
		}
	}
	h.mu.RUnlock()
}

// Handler returns the SSE endpoint handler for GET /bookings/events.
func (h *Hub) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		// Disable per-connection write deadline so the server's WriteTimeout
		// doesn't kill the long-lived SSE connection.
		rc := http.NewResponseController(w)
		if err := rc.SetWriteDeadline(time.Time{}); err != nil {
			http.Error(w, "cannot clear deadline", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		// Tell nginx not to buffer this response.
		w.Header().Set("X-Accel-Buffering", "no")
		// Custom retry so the browser reconnects after 5 s on disconnect.
		fmt.Fprintf(w, "retry: 5000\n\n")
		flusher.Flush()

		ch := h.subscribe()
		defer h.unsubscribe(ch)

		// Heartbeat every 25 s keeps the connection alive through nginx's
		// proxy_read_timeout (30 s).
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ch:
				fmt.Fprintf(w, "event: update\ndata: {}\n\n")
				flusher.Flush()
			case <-ticker.C:
				fmt.Fprintf(w, ": heartbeat\n\n")
				flusher.Flush()
			}
		}
	}
}
