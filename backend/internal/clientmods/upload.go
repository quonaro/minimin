package clientmods

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"orchestrator/internal/mods"
	"orchestrator/internal/runner"
)

// UploadClientMod saves a .jar or .zip into the server's mods-client directory.
func (s *Service) UploadClientMod(serverID string, file io.Reader, filename string, size int64) ([]string, error) {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return nil, mods.ErrNotFound
	}
	modsDir, err := s.getClientModsDir(serverID)
	if err != nil {
		return nil, err
	}
	if modsDir == "" {
		return nil, mods.ErrNotFound
	}

	isZip := strings.HasSuffix(strings.ToLower(filename), ".zip")
	isJar := strings.HasSuffix(strings.ToLower(filename), ".jar")
	if !isZip && !isJar {
		return nil, fmt.Errorf("only .jar and .zip files are allowed")
	}

	uid, gid := runner.ContainerUIDGID()
	if err := runner.FixVolumeOwnership(context.Background(), s.cli, st.VolumePath, uid, gid); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", serverID, "path", st.VolumePath, "error", err)
	}
	if err := os.MkdirAll(modsDir, 0o775); err != nil {
		if info, statErr := os.Stat(modsDir); statErr != nil || !info.IsDir() {
			return nil, err
		}
	}

	if isJar {
		targetPath := filepath.Join(modsDir, filename)
		out, err := os.Create(targetPath)
		if err != nil {
			return nil, err
		}
		if _, err := out.ReadFrom(file); err != nil {
			_ = out.Close()
			return nil, err
		}
		_ = out.Close()
		return []string{filename}, nil
	}

	// .zip
	tmpPath := filepath.Join(modsDir, ".upload_"+filename)
	tmp, err := os.Create(tmpPath)
	if err != nil {
		return nil, err
	}
	if _, err := tmp.ReadFrom(file); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return nil, err
	}
	_ = tmp.Close()

	extracted, extractErr := mods.ExtractZipJars(tmpPath, modsDir)
	_ = os.Remove(tmpPath)
	if extractErr != nil {
		return nil, extractErr
	}
	return extracted, nil
}

// DownloadClientModFromURL downloads a mod from a remote URL and saves it into mods-client.
func (s *Service) DownloadClientModFromURL(serverID, urlStr, filename string) (string, error) {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return "", mods.ErrNotFound
	}
	modsDir, err := s.getClientModsDir(serverID)
	if err != nil {
		return "", err
	}
	if modsDir == "" {
		return "", mods.ErrNotFound
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(urlStr)
	if err != nil {
		slog.Error("failed to download client mod", "url", urlStr, "error", err)
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("upstream returned %d", resp.StatusCode)
	}

	uid, gid := runner.ContainerUIDGID()
	if err := runner.FixVolumeOwnership(context.Background(), s.cli, st.VolumePath, uid, gid); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", serverID, "path", st.VolumePath, "error", err)
	}
	if err := os.MkdirAll(modsDir, 0o775); err != nil {
		return "", err
	}

	if filename == "" {
		filename = filepath.Base(urlStr)
	}
	if filename == "" || filename == "." {
		filename = "mod.jar"
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".jar") {
		return "", fmt.Errorf("only .jar files are allowed")
	}

	targetPath := filepath.Join(modsDir, filename)
	out, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}
	return filename, nil
}
