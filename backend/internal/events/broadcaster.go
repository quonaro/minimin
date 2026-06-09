package events

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Event is a generic notification payload sent over SSE.
type Event struct {
	Type               string `json:"type"`
	AgentID            string `json:"agentId,omitempty"`
	ServerID           string `json:"serverId,omitempty"`
	Message            string `json:"message"`
	Timestamp          string `json:"timestamp"`
	OldStatus          string `json:"oldStatus,omitempty"`
	NewStatus          string `json:"newStatus,omitempty"`
	DesiredStatus      string `json:"desiredStatus,omitempty"`
	OldContainerStatus string `json:"oldContainerStatus,omitempty"`
	NewContainerStatus string `json:"newContainerStatus,omitempty"`
	OldServerStatus    string `json:"oldServerStatus,omitempty"`
	NewServerStatus    string `json:"newServerStatus,omitempty"`
	ContainerStartedAt string `json:"containerStartedAt,omitempty"`
	ServerStartedAt    string `json:"serverStartedAt,omitempty"`
}

// Broadcaster manages SSE clients and distributes events.
type Broadcaster struct {
	mu      sync.RWMutex
	clients map[chan Event]struct{}
}

// NewBroadcaster creates a new event broadcaster.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[chan Event]struct{}),
	}
}

// Broadcast sends an event to all connected SSE clients.
func (b *Broadcaster) Broadcast(evt Event) {
	if evt.Timestamp == "" {
		evt.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.clients {
		select {
		case ch <- evt:
		default:
			// Drop event if client buffer is full.
		}
	}
}

// AddClient registers a new client channel.
func (b *Broadcaster) AddClient(ch chan Event) {
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
}

// RemoveClient unregisters a client channel.
func (b *Broadcaster) RemoveClient(ch chan Event) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
	close(ch)
}

// ServeHTTP handles SSE connections at GET /api/events.
func (b *Broadcaster) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		slog.Warn("streaming not supported")
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ch := make(chan Event, 64)
	b.AddClient(ch)
	defer b.RemoveClient(ch)

	// Send a keep-alive ping every 30 seconds.
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case evt, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(evt)
			if err != nil {
				slog.Warn("failed to marshal event", "error", err)
				continue
			}
			_, _ = fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-ticker.C:
			_, _ = fmt.Fprint(w, ":ping\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
