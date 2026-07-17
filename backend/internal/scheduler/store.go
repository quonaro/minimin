package scheduler

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Store persists and loads the actions file.
type Store struct {
	mu   sync.RWMutex
	path string
	data ActionsFile
}

// NewStore creates a new Store backed by the given YAML file.
func NewStore(path string) (*Store, error) {
	s := &Store{
		path: path,
		data: ActionsFile{Tasks: []Task{}},
	}
	if err := s.Load(); err != nil {
		return nil, err
	}
	return s, nil
}

// Load reads the actions file from disk.
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read actions file: %w", err)
	}

	if err := yaml.Unmarshal(data, &s.data); err != nil {
		return fmt.Errorf("failed to parse actions file: %w", err)
	}
	if s.data.Tasks == nil {
		s.data.Tasks = []Task{}
	}
	return nil
}

// Save writes the current state back to disk atomically.
func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := yaml.Marshal(&s.data)
	if err != nil {
		return fmt.Errorf("failed to marshal actions file: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write actions file: %w", err)
	}
	return nil
}

// All returns a snapshot of all tasks.
func (s *Store) All() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Task, len(s.data.Tasks))
	copy(out, s.data.Tasks)
	return out
}

// AllForServer returns tasks filtered by server ID.
func (s *Store) AllForServer(serverID string) []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Task
	for _, t := range s.data.Tasks {
		if t.ServerID == serverID {
			out = append(out, t)
		}
	}
	return out
}

// Get returns a task by ID.
func (s *Store) Get(id string) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.data.Tasks {
		if t.ID == id {
			return t, true
		}
	}
	return Task{}, false
}

// Put inserts or updates a task.
func (s *Store) Put(t Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, existing := range s.data.Tasks {
		if existing.ID == t.ID {
			s.data.Tasks[i] = t
			return
		}
	}
	s.data.Tasks = append(s.data.Tasks, t)
}

// Delete removes a task by ID.
func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.data.Tasks {
		if t.ID == id {
			s.data.Tasks = append(s.data.Tasks[:i], s.data.Tasks[i+1:]...)
			return true
		}
	}
	return false
}
