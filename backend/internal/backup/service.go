// Package backup provides server world backup creation, retention and restoration.
package backup

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"orchestrator/internal/runner"
	"orchestrator/internal/state"
)

// Service handles backup operations.
type Service struct {
	instance    *state.InstanceFile
	backupsDir  string
	serversDir  string
	rconTimeout time.Duration
}

// NewService creates a backup service.
func NewService(instance *state.InstanceFile, backupsDir, serversDir string) *Service {
	return &Service{
		instance:    instance,
		backupsDir:  backupsDir,
		serversDir:  serversDir,
		rconTimeout: 10 * time.Second,
	}
}

// Create performs a hot backup of the server's world directory.
// keepLast and keepDays control retention (0 = unlimited).
func (s *Service) Create(serverID string, keepLast, keepDays int) (*Backup, error) {
	start := time.Now()
	slog.Info("backup create started", "server_id", serverID)
	defer func() {
		slog.Info("backup create finished", "server_id", serverID, "elapsed", time.Since(start))
	}()

	server, ok := s.instance.Get(serverID)
	if !ok {
		return nil, fmt.Errorf("server not found")
	}
	if server.HostPath == "" {
		return nil, fmt.Errorf("server has no host path")
	}

	worldPath := filepath.Join(server.HostPath, "world")
	if _, err := os.Stat(worldPath); err != nil {
		// Fallback to serversDir if HostPath world is missing
		if s.serversDir != "" {
			fallback := filepath.Join(s.serversDir, serverID, "world")
			if _, err2 := os.Stat(fallback); err2 == nil {
				worldPath = fallback
			} else {
				return nil, fmt.Errorf("world directory not found: %w", err)
			}
		} else {
			return nil, fmt.Errorf("world directory not found: %w", err)
		}
	}

	// Hot backup via RCON if server is running
	if server.ServerStatus == "running" && server.RconPort > 0 {
		addr := fmt.Sprintf("127.0.0.1:%d", server.RconPort)
		if server.PublicRcon {
			addr = fmt.Sprintf("0.0.0.0:%d", server.RconPort)
		}
		client, err := runner.DialRCON(addr, server.RconPassword, s.rconTimeout)
		if err != nil {
			slog.Warn("rcon dial failed for backup, proceeding without save-off", "server_id", serverID, "error", err)
		} else {
			_, _ = client.Execute("save-off")
			_, _ = client.Execute("save-all")
			_ = client.Close()
		}
		defer func() {
			client2, err := runner.DialRCON(addr, server.RconPassword, s.rconTimeout)
			if err == nil {
				_, _ = client2.Execute("save-on")
				_ = client2.Close()
			}
		}()
	}

	backupDir := filepath.Join(s.backupsDir, serverID)
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return nil, fmt.Errorf("create backup dir: %w", err)
	}

	ts := time.Now().UTC().Format("20060102_150405")
	name := fmt.Sprintf("%s.tar.gz", ts)
	archivePath := filepath.Join(backupDir, name)

	if err := createTarGz(worldPath, archivePath); err != nil {
		return nil, fmt.Errorf("archive world: %w", err)
	}

	info, err := os.Stat(archivePath)
	if err != nil {
		return nil, err
	}

	bk := &Backup{
		Name:      name,
		ServerID:  serverID,
		SizeBytes: info.Size(),
		CreatedAt: time.Now().UTC(),
	}

	// Apply retention
	s.applyRetention(serverID, keepLast, keepDays)

	return bk, nil
}

// List returns all backups for a server.
func (s *Service) List(serverID string) ([]Backup, error) {
	backupDir := filepath.Join(s.backupsDir, serverID)
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Backup{}, nil
		}
		return nil, err
	}

	out := make([]Backup, 0)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".tar.gz") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, Backup{
			Name:      e.Name(),
			ServerID:  serverID,
			SizeBytes: info.Size(),
			CreatedAt: info.ModTime().UTC(),
		})
	}

	sort.Slice(out, func(a, b int) bool {
		return out[a].CreatedAt.After(out[b].CreatedAt)
	})
	return out, nil
}

// Delete removes a backup archive.
func (s *Service) Delete(serverID, name string) error {
	path := filepath.Join(s.backupsDir, serverID, name)
	return os.Remove(path)
}

// DownloadPath returns the absolute filesystem path for a backup.
func (s *Service) DownloadPath(serverID, name string) string {
	return filepath.Join(s.backupsDir, serverID, name)
}

// Restore extracts a backup over the server's world directory.
func (s *Service) Restore(serverID, name string) error {
	server, ok := s.instance.Get(serverID)
	if !ok {
		return fmt.Errorf("server not found")
	}
	if server.HostPath == "" {
		return fmt.Errorf("server has no host path")
	}

	archivePath := filepath.Join(s.backupsDir, serverID, name)
	if _, err := os.Stat(archivePath); err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}

	worldPath := filepath.Join(server.HostPath, "world")
	// Remove existing world dir
	if err := os.RemoveAll(worldPath); err != nil {
		return fmt.Errorf("remove existing world: %w", err)
	}

	if err := extractTarGz(archivePath, server.HostPath); err != nil {
		return fmt.Errorf("extract backup: %w", err)
	}
	return nil
}

// applyRetention enforces keep-last and keep-days limits.
func (s *Service) applyRetention(serverID string, keepLast, keepDays int) {
	if keepLast <= 0 && keepDays <= 0 {
		return
	}

	backups, err := s.List(serverID)
	if err != nil {
		slog.Warn("failed to list backups for retention", "server_id", serverID, "error", err)
		return
	}

	cutoff := time.Time{}
	if keepDays > 0 {
		cutoff = time.Now().UTC().AddDate(0, 0, -keepDays)
	}

	for i, bk := range backups {
		// keepLast: keep the most recent N backups
		if keepLast > 0 && i < keepLast {
			continue
		}
		// keepDays: delete backups older than cutoff
		if !cutoff.IsZero() && bk.CreatedAt.After(cutoff) {
			continue
		}
		if err := s.Delete(serverID, bk.Name); err != nil {
			slog.Warn("failed to delete old backup", "name", bk.Name, "error", err)
		}
	}
}
