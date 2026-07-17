package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"orchestrator/internal/state"
)

func TestCreateBackupZeroRetentionLeavesFile(t *testing.T) {
	tmpDir := t.TempDir()
	backupsDir := filepath.Join(tmpDir, "backups")
	serversDir := filepath.Join(tmpDir, "servers")

	worldDir := filepath.Join(serversDir, "srv1", "world")
	if err := os.MkdirAll(filepath.Join(worldDir, "region"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worldDir, "level.dat"), []byte("level"), 0o644); err != nil {
		t.Fatal(err)
	}

	instance := &state.InstanceFile{
		Servers: map[string]state.ServerState{
			"srv1": {
				ServerID:     "srv1",
				HostPath:     filepath.Join(serversDir, "srv1"),
				ServerStatus: "stopped",
			},
		},
	}

	svc := NewService(instance, backupsDir, serversDir)

	bk, err := svc.Create("srv1", 0, 0)
	if err != nil {
		t.Fatalf("create backup: %v", err)
	}

	path := filepath.Join(backupsDir, "srv1", bk.Name)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("backup file should exist: %v", err)
	}

	list, err := svc.List("srv1")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 backup, got %d", len(list))
	}
}

func TestApplyRetentionKeepLast(t *testing.T) {
	tmpDir := t.TempDir()
	backupsDir := filepath.Join(tmpDir, "backups")
	svc := NewService(&state.InstanceFile{}, backupsDir, "")

	serverID := "srv1"
	backupDir := filepath.Join(backupsDir, serverID)
	_ = os.MkdirAll(backupDir, 0o755)

	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("2026010%d_000000.tar.gz", i+1)
		path := filepath.Join(backupDir, name)
		if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
			t.Fatal(err)
		}
		modTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * time.Hour)
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatal(err)
		}
	}

	svc.applyRetention(serverID, 2, 0)

	list, err := svc.List(serverID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 backups, got %d", len(list))
	}
	for _, bk := range list {
		if bk.Name != "20260105_000000.tar.gz" && bk.Name != "20260104_000000.tar.gz" {
			t.Fatalf("unexpected backup remaining: %s", bk.Name)
		}
	}
}

func TestApplyRetentionKeepDays(t *testing.T) {
	tmpDir := t.TempDir()
	backupsDir := filepath.Join(tmpDir, "backups")
	svc := NewService(&state.InstanceFile{}, backupsDir, "")

	serverID := "srv1"
	backupDir := filepath.Join(backupsDir, serverID)
	_ = os.MkdirAll(backupDir, 0o755)

	now := time.Now().UTC()

	recentName := "recent.tar.gz"
	recentPath := filepath.Join(backupDir, recentName)
	_ = os.WriteFile(recentPath, []byte("data"), 0o644)
	_ = os.Chtimes(recentPath, now.Add(-1*time.Hour), now.Add(-1*time.Hour))

	oldName := "old.tar.gz"
	oldPath := filepath.Join(backupDir, oldName)
	_ = os.WriteFile(oldPath, []byte("data"), 0o644)
	_ = os.Chtimes(oldPath, now.Add(-48*time.Hour), now.Add(-48*time.Hour))

	svc.applyRetention(serverID, 0, 1)

	list, err := svc.List(serverID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 backup, got %d", len(list))
	}
	if list[0].Name != recentName {
		t.Fatalf("expected %s, got %s", recentName, list[0].Name)
	}
}
