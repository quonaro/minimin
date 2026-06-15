package archive

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"context"
	"orchestrator/internal/state"
)

var (
	tokens   = make(map[string]*Token)
	tokensMu sync.RWMutex
)

func metaPath(volumePath string) string {
	return filepath.Join(volumePath, ".webui-archive-meta.json")
}

func loadMeta(volumePath string) []metaEntry {
	data, err := os.ReadFile(metaPath(volumePath))
	if err != nil {
		return nil
	}
	var entries []metaEntry
	_ = json.Unmarshal(data, &entries)
	return entries
}

func saveMeta(volumePath string, entries []metaEntry) {
	path := metaPath(volumePath)
	data, _ := json.MarshalIndent(entries, "", "  ")
	_ = os.WriteFile(path, data, 0o644)
}

// InitArchives loads archive metadata from all server volumes on startup.
func InitArchives(instance *state.InstanceFile) {
	tokensMu.Lock()
	defer tokensMu.Unlock()

	for _, s := range instance.All() {
		if s.VolumePath == "" {
			continue
		}
		entries := loadMeta(s.VolumePath)
		for _, e := range entries {
			t := entryToToken(e)
			tokens[t.Token] = t
		}
	}
}

func saveArchiveMeta(instance *state.InstanceFile, serverID string, archive *Token) {
	s, ok := instance.Get(serverID)
	if !ok || s.VolumePath == "" {
		return
	}
	entries := loadMeta(s.VolumePath)
	found := false
	for i, e := range entries {
		if e.Token == archive.Token {
			entries[i] = archive.toMetaEntry()
			found = true
			break
		}
	}
	if !found {
		entries = append(entries, archive.toMetaEntry())
	}
	saveMeta(s.VolumePath, entries)
}

func removeArchiveMeta(instance *state.InstanceFile, serverID, token string) {
	s, ok := instance.Get(serverID)
	if !ok || s.VolumePath == "" {
		return
	}
	entries := loadMeta(s.VolumePath)
	filtered := entries[:0]
	for _, e := range entries {
		if e.Token != token {
			filtered = append(filtered, e)
		}
	}
	saveMeta(s.VolumePath, filtered)
}

func cleanupExpiredArchives(instance *state.InstanceFile) {
	tokensMu.Lock()
	now := time.Now()
	type del struct{ token, serverID string }
	var toDelete []del
	for token, entry := range tokens {
		if now.After(entry.ExpiresAt) {
			delete(tokens, token)
			toDelete = append(toDelete, del{token, entry.ServerID})
		}
	}
	tokensMu.Unlock()

	for _, d := range toDelete {
		removeArchiveMeta(instance, d.serverID, d.token)
	}
}

// StartCleanup runs a background ticker that removes expired archives.
func StartCleanup(ctx context.Context, instance *state.InstanceFile) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cleanupExpiredArchives(instance)
		case <-ctx.Done():
			return
		}
	}
}
