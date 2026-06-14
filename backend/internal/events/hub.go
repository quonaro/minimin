package events

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"
)

// Event is a single SSE event sent to all connected clients.
type Event struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

// Client represents a single SSE subscriber.
type client struct {
	ch chan Event
}

// Hub manages a set of SSE client channels.
type Hub struct {
	mu          sync.RWMutex
	clients     map[int]*client
	nextID      int
	closed      bool
	lastMetrics sync.Map // map[string]MetricsPayload
}

// NewHub creates a new event hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[int]*client),
	}
}

// Register creates a new buffered client channel and returns a client ID
// together with the receive-only channel.
func (h *Hub) Register() (int, <-chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return 0, nil
	}
	id := h.nextID
	h.nextID++
	c := &client{ch: make(chan Event, 16)}
	h.clients[id] = c
	return id, c.ch
}

// Unregister removes a client from the hub and closes its channel.
func (h *Hub) Unregister(id int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	c, ok := h.clients[id]
	if !ok {
		return
	}
	delete(h.clients, id)
	close(c.ch)
}

// Broadcast sends an event to all registered clients.
// If a client's channel is full, the event is dropped for that client.
func (h *Hub) Broadcast(ev Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return
	}
	for _, c := range h.clients {
		select {
		case c.ch <- ev:
		default:
			slog.Debug("SSE client channel full, dropping event", "type", ev.Type)
		}
	}
}

// HasClients reports whether there is at least one connected client.
func (h *Hub) HasClients() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients) > 0
}

// Close shuts down the hub and closes all client channels.
func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.closed = true
	for _, c := range h.clients {
		close(c.ch)
	}
	h.clients = make(map[int]*client)
}

// BroadcastJSON is a convenience helper that marshals a payload and broadcasts it.
func (h *Hub) BroadcastJSON(eventType string, payload any) {
	h.Broadcast(Event{Type: eventType, Payload: payload})
}

// StoreMetrics saves the latest metrics for a server.
func (h *Hub) StoreMetrics(serverID string, payload MetricsPayload) {
	h.lastMetrics.Store(serverID, payload)
}

// DeleteMetrics removes cached metrics for a server.
func (h *Hub) DeleteMetrics(serverID string) {
	h.lastMetrics.Delete(serverID)
}

// RangeMetrics iterates over the last known metrics for each server.
func (h *Hub) RangeMetrics(f func(serverID string, payload MetricsPayload) bool) {
	h.lastMetrics.Range(func(key, value any) bool {
		return f(key.(string), value.(MetricsPayload))
	})
}

// ServerStatusPayload is broadcast when a server's state changes.
type ServerStatusPayload struct {
	ServerID string `json:"serverId"`
}

// PlayerListPayload is broadcast for online players.
type PlayerListPayload struct {
	ServerID string   `json:"serverId"`
	Players  []string `json:"players"`
	Max      int      `json:"max"`
}

// OpsPayload is broadcast for operators.
type OpsPayload struct {
	ServerID string           `json:"serverId"`
	Ops      []map[string]any `json:"ops"`
}

// BansPayload is broadcast for banned players.
type BansPayload struct {
	ServerID string           `json:"serverId"`
	Bans     []map[string]any `json:"bans"`
}

// WhitelistPayload is broadcast for whitelist entries.
type WhitelistPayload struct {
	ServerID  string           `json:"serverId"`
	Whitelist []map[string]any `json:"whitelist"`
}

// MetricsPayload is broadcast for real-time server metrics.
type MetricsPayload struct {
	ServerID  string    `json:"serverId"`
	RAMUsage  uint64    `json:"ramUsage"`
	RAMLimit  uint64    `json:"ramLimit"`
	CPU       float64   `json:"cpu"`
	Online    int       `json:"online"`
	Max       int       `json:"max"`
	TPS       *float64  `json:"tps,omitempty"`
	RxRate    float64   `json:"rxRate"`
	TxRate    float64   `json:"txRate"`
	Timestamp time.Time `json:"timestamp"`
}

// MarshalEvent formats an Event as an SSE message.
func MarshalEvent(ev Event) ([]byte, error) {
	data, err := json.Marshal(ev.Payload)
	if err != nil {
		return nil, err
	}
	return []byte("event: " + ev.Type + "\ndata: " + string(data) + "\n\n"), nil
}
