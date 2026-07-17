package actionlog

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

const maxEntries = 2000

// Store persists action log entries to a YAML file.
type Store struct {
	path string
	mu   sync.Mutex
}

// NewStore creates an action log store.
func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.ensureFile(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) ensureFile() error {
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		return s.saveLocked(&LogFile{Entries: []Entry{}})
	}
	return nil
}

// Append adds a new entry and trims old ones if over limit.
func (s *Store) Append(e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadLocked()
	if err != nil {
		return err
	}
	data.Entries = append(data.Entries, e)
	if len(data.Entries) > maxEntries {
		data.Entries = data.Entries[len(data.Entries)-maxEntries:]
	}
	return s.saveLocked(data)
}

// List returns entries for a server in reverse chronological order.
func (s *Store) List(serverID string) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadLocked()
	if err != nil {
		return nil, err
	}
	var out []Entry
	for i := len(data.Entries) - 1; i >= 0; i-- {
		if data.Entries[i].ServerID == serverID {
			out = append(out, data.Entries[i])
		}
	}
	return out, nil
}

func (s *Store) loadLocked() (*LogFile, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("read action log: %w", err)
	}
	var data LogFile
	if err := yaml.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("parse action log: %w", err)
	}
	return &data, nil
}

func (s *Store) saveLocked(data *LogFile) error {
	b, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal action log: %w", err)
	}
	if err := os.WriteFile(s.path, b, 0o644); err != nil {
		return fmt.Errorf("write action log: %w", err)
	}
	return nil
}
