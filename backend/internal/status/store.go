package status

import (
	"sync"
	"time"
)

// AgentStatus represents the online status of a single agent.
type AgentStatus struct {
	ID     string `json:"id"`
	Online bool   `json:"online"`
}

// Store holds in-memory agent health statuses.
type Store struct {
	mu            sync.RWMutex
	statuses      map[string]bool
	lastCheckAt   time.Time
	checkInterval time.Duration
}

// NewStore creates a new Store.
func NewStore() *Store {
	return &Store{
		statuses: make(map[string]bool),
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
