// Package persistent provides a generic bbolt-backed key-value store with TTL support.
package persistent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

// DB wraps a bbolt database with typed get/set/delete operations.
type DB struct {
	db *bolt.DB
}

// Open opens (or creates) a bbolt database at the given path.
func Open(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	db, err := bolt.Open(path, 0o644, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bbolt db: %w", err)
	}
	return &DB{db: db}, nil
}

// Close closes the underlying database.
func (d *DB) Close() error {
	return d.db.Close()
}

// Get retrieves a value from the given bucket and unmarshals it into v.
// Returns true if the key exists and has not expired.
func (d *DB) Get(bucket, key string, v any) (bool, error) {
	var raw []byte
	found := false
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		data := b.Get([]byte(key))
		if data == nil {
			return nil
		}
		raw = make([]byte, len(data))
		copy(raw, data)
		found = true
		return nil
	})
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}

	var entry struct {
		Data      json.RawMessage `json:"data"`
		ExpiresAt int64           `json:"expires_at"`
	}
	if err := json.Unmarshal(raw, &entry); err != nil {
		return false, fmt.Errorf("unmarshal cache entry: %w", err)
	}
	if time.Now().Unix() > entry.ExpiresAt {
		_ = d.Delete(bucket, key)
		return false, nil
	}

	if err := json.Unmarshal(entry.Data, v); err != nil {
		return false, fmt.Errorf("unmarshal cached data: %w", err)
	}
	return true, nil
}

// Set marshals v and stores it in the given bucket with the specified TTL.
func (d *DB) Set(bucket, key string, v any, ttl time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}
	entry := struct {
		Data      json.RawMessage `json:"data"`
		ExpiresAt int64           `json:"expires_at"`
	}{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}
	raw, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal cache entry: %w", err)
	}

	return d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("create bucket %q: %w", bucket, err)
		}
		if err := b.Put([]byte(key), raw); err != nil {
			return fmt.Errorf("put key: %w", err)
		}
		return nil
	})
}

// Delete removes a key from the given bucket.
func (d *DB) Delete(bucket, key string) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(key))
	})
}

// EvictStale removes all expired entries from the given bucket.
func (d *DB) EvictStale(bucket string) error {
	now := time.Now().Unix()
	var keysToDelete [][]byte

	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var entry struct {
				ExpiresAt int64 `json:"expires_at"`
			}
			if err := json.Unmarshal(v, &entry); err != nil {
				keysToDelete = append(keysToDelete, append([]byte(nil), k...))
				continue
			}
			if now > entry.ExpiresAt {
				keysToDelete = append(keysToDelete, append([]byte(nil), k...))
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(keysToDelete) == 0 {
		return nil
	}

	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		for _, k := range keysToDelete {
			if err := b.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
}

// Compact triggers a database compaction. Call periodically (e.g. on startup).
func (d *DB) Compact() error {
	return d.db.Sync()
}
