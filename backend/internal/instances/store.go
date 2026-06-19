package instances

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// defaultTTL is how long a prepared archive is kept before it must be used.
const defaultTTL = 30 * time.Minute

// Store keeps uploaded archives in a temporary directory while the client
// decides whether to create a server from them.
type Store struct {
	dir   string
	mu    sync.RWMutex
	items map[string]*TempEntry
}

// NewStore creates a temporary file store under the given directory.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create temp store directory: %w", err)
	}
	return &Store{dir: dir, items: make(map[string]*TempEntry)}, nil
}

// Save writes the uploaded data to a temporary file and returns a token.
func (s *Store) Save(name string, size int64, r func(dst *os.File) error) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}

	path := filepath.Join(s.dir, token+".zip")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return "", err
	}
	if err := r(f); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return "", err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}

	entry := &TempEntry{
		Token:     token,
		Path:      path,
		CreatedAt: time.Now(),
		Size:      size,
	}

	s.mu.Lock()
	s.items[token] = entry
	s.mu.Unlock()
	return token, nil
}

// Get returns a prepared archive by token.
func (s *Store) Get(token string) (*TempEntry, error) {
	s.mu.RLock()
	entry, ok := s.items[token]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("instance token not found")
	}
	if time.Since(entry.CreatedAt) > defaultTTL {
		_ = s.Remove(token)
		return nil, fmt.Errorf("instance token expired")
	}
	return entry, nil
}

// Remove deletes a temporary archive.
func (s *Store) Remove(token string) error {
	s.mu.Lock()
	entry, ok := s.items[token]
	if ok {
		delete(s.items, token)
	}
	s.mu.Unlock()
	if entry != nil {
		_ = os.Remove(entry.Path)
	}
	return nil
}

// Cleanup removes expired entries.
func (s *Store) Cleanup() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for token, entry := range s.items {
		if now.Sub(entry.CreatedAt) > defaultTTL {
			delete(s.items, token)
			_ = os.Remove(entry.Path)
		}
	}
}

// generateToken creates a random 16-byte hex token.
func generateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
