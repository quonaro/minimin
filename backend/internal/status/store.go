package status

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// AgentStatus represents the online status of a single agent.
type AgentStatus struct {
	ID     string `json:"id"`
	Online bool   `json:"online"`
}

// Store holds in-memory agent health statuses and manages WebSocket clients.
type Store struct {
	mu            sync.RWMutex
	statuses      map[string]bool
	clients       map[*websocket.Conn]bool
	upgrader      websocket.Upgrader
	lastCheckAt   time.Time
	checkInterval time.Duration
}

// NewStore creates a new Store.
func NewStore() *Store {
	return &Store{
		statuses: make(map[string]bool),
		clients:  make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// SetCheckInterval sets the interval between health checks.
func (s *Store) SetCheckInterval(d time.Duration) {
	s.mu.Lock()
	s.checkInterval = d
	s.mu.Unlock()
}

// SetLastCheck records the time of the most recent health check.
func (s *Store) SetLastCheck(t time.Time) {
	s.mu.Lock()
	s.lastCheckAt = t
	s.mu.Unlock()
}

// NextCheckAt returns the estimated time of the next health check.
func (s *Store) NextCheckAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastCheckAt.Add(s.checkInterval)
}

// Set updates the online status for a given agent.
func (s *Store) Set(agentID string, online bool) {
	s.mu.Lock()
	s.statuses[agentID] = online
	s.mu.Unlock()
}

// Get returns the online status for a given agent.
func (s *Store) Get(agentID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.statuses[agentID]
}

// All returns a snapshot of all agent statuses.
func (s *Store) All() []AgentStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]AgentStatus, 0, len(s.statuses))
	for id, online := range s.statuses {
		result = append(result, AgentStatus{ID: id, Online: online})
	}
	return result
}

// Broadcast marshals the current statuses and sends them to all connected clients.
func (s *Store) Broadcast() {
	payload := struct {
		Statuses    []AgentStatus `json:"statuses"`
		NextCheckAt time.Time     `json:"next_check_at"`
	}{
		Statuses:    s.All(),
		NextCheckAt: s.NextCheckAt(),
	}
	msg, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal agent statuses", "error", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for client := range s.clients {
		if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
			client.Close()
			delete(s.clients, client)
		}
	}
}

// AddClient registers a new WebSocket client and immediately sends current statuses.
func (s *Store) AddClient(conn *websocket.Conn) {
	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	payload := struct {
		Statuses    []AgentStatus `json:"statuses"`
		NextCheckAt time.Time     `json:"next_check_at"`
	}{
		Statuses:    s.All(),
		NextCheckAt: s.NextCheckAt(),
	}
	msg, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal initial statuses", "error", err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		slog.Warn("failed to send initial statuses", "error", err)
	}
}

// RemoveClient unregisters a WebSocket client.
func (s *Store) RemoveClient(conn *websocket.Conn) {
	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()
	conn.Close()
}

// ServeHTTP upgrades the HTTP connection to WebSocket and keeps it open.
func (s *Store) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("websocket upgrade failed", "error", err)
		return
	}
	s.AddClient(conn)
	defer s.RemoveClient(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
