package clientmods

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"orchestrator/internal/mods"
	"orchestrator/internal/runner"
)

// MoveMod moves a mod file between mods/ and mods-client/ directories.
func (s *Service) MoveMod(serverID string, req struct {
	Filename string `json:"filename"`
	Target   string `json:"target"`
}) error {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return mods.ErrNotFound
	}
	if st.VolumePath == "" {
		return mods.ErrVolumeNotInitialized
	}

	serverDir := filepath.Join(st.VolumePath, "mods")
	clientDir := filepath.Join(st.VolumePath, "mods-client")

	uid, gid := runner.ContainerUIDGID()
	if err := runner.FixVolumeOwnership(context.Background(), s.cli, st.VolumePath, uid, gid); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", serverID, "path", st.VolumePath, "error", err)
	}
	if err := os.MkdirAll(clientDir, 0o775); err != nil {
		if info, statErr := os.Stat(clientDir); statErr != nil || !info.IsDir() {
			return fmt.Errorf("failed to create client mods directory: %w", err)
		}
	}

	filename := filepath.Base(req.Filename)
	var srcDir, dstDir string
	switch req.Target {
	case "client":
		srcDir = serverDir
		dstDir = clientDir
	case "server":
		srcDir = clientDir
		dstDir = serverDir
	default:
		return fmt.Errorf("target must be 'server' or 'client'")
	}

	srcPath := filepath.Join(srcDir, filename)
	dstPath := filepath.Join(dstDir, filename)
	if !strings.HasPrefix(srcPath, srcDir+string(filepath.Separator)) && srcPath != srcDir {
		return mods.ErrInvalidFilename
	}

	if err := os.Rename(srcPath, dstPath); err != nil {
		if os.IsNotExist(err) {
			return mods.ErrNotFound
		}
		return fmt.Errorf("failed to move mod: %w", err)
	}
	return nil
}

// CopyMod copies a mod file between mods/ and mods-client/ directories.
func (s *Service) CopyMod(serverID string, req struct {
	Filename string `json:"filename"`
	Source   string `json:"source"`
	Target   string `json:"target"`
}) error {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return mods.ErrNotFound
	}
	if st.VolumePath == "" {
		return mods.ErrVolumeNotInitialized
	}

	serverDir := filepath.Join(st.VolumePath, "mods")
	clientDir := filepath.Join(st.VolumePath, "mods-client")

	uid, gid := runner.ContainerUIDGID()
	if err := runner.FixVolumeOwnership(context.Background(), s.cli, st.VolumePath, uid, gid); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", serverID, "path", st.VolumePath, "error", err)
	}
	if err := os.MkdirAll(clientDir, 0o775); err != nil {
		if info, statErr := os.Stat(clientDir); statErr != nil || !info.IsDir() {
			return fmt.Errorf("failed to create client mods directory: %w", err)
		}
	}
	if err := os.MkdirAll(serverDir, 0o775); err != nil {
		if info, statErr := os.Stat(serverDir); statErr != nil || !info.IsDir() {
			return fmt.Errorf("failed to create server mods directory: %w", err)
		}
	}

	filename := filepath.Base(req.Filename)
	var srcDir, dstDir string
	switch req.Source {
	case "server":
		srcDir = serverDir
	case "client":
		srcDir = clientDir
	default:
		return fmt.Errorf("source must be 'server' or 'client'")
	}
	switch req.Target {
	case "server":
		dstDir = serverDir
	case "client":
		dstDir = clientDir
	default:
		return fmt.Errorf("target must be 'server' or 'client'")
	}

	srcPath := filepath.Join(srcDir, filename)
	dstPath := filepath.Join(dstDir, filename)
	if !strings.HasPrefix(srcPath, srcDir+string(filepath.Separator)) && srcPath != srcDir {
		return mods.ErrInvalidFilename
	}
	if !strings.HasPrefix(dstPath, dstDir+string(filepath.Separator)) && dstPath != dstDir {
		return mods.ErrInvalidFilename
	}

	src, err := os.Open(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return mods.ErrNotFound
		}
		return fmt.Errorf("failed to open mod: %w", err)
	}
	defer func() { _ = src.Close() }()

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = dst.Close() }()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy mod: %w", err)
	}
	return nil
}
