package mods

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

	"orchestrator/internal/runner"
)

// UploadServerMod saves a .jar or .zip into the server's mods directory.
// It returns the list of extracted filenames.
func (s *Service) UploadServerMod(serverID string, file io.Reader, filename string, size int64) ([]string, error) {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return nil, ErrNotFound
	}
	if st.VolumePath == "" {
		return nil, ErrVolumeNotInitialized
	}

	modsDir := filepath.Join(st.VolumePath, "mods")
	uid, gid := runner.ContainerUIDGID()
	if err := runner.FixVolumeOwnership(context.Background(), s.cli, st.VolumePath, uid, gid); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", serverID, "path", st.VolumePath, "error", err)
	}
	if err := os.MkdirAll(modsDir, 0o775); err != nil {
		return nil, err
	}

	isZip := strings.HasSuffix(strings.ToLower(filename), ".zip")
	isJar := strings.HasSuffix(strings.ToLower(filename), ".jar")
	if !isZip && !isJar {
		return nil, fmt.Errorf("only .jar and .zip files are allowed")
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

	// .zip: save to temp, extract jars, delete temp
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

	extracted, extractErr := ExtractZipJars(tmpPath, modsDir)
	_ = os.Remove(tmpPath)
	if extractErr != nil {
		return nil, extractErr
	}
	return extracted, nil
}

// DownloadModFromURL downloads a mod from a remote URL and saves it.
func (s *Service) DownloadModFromURL(serverID, urlStr, filename string) (string, error) {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return "", ErrNotFound
	}
	if st.VolumePath == "" {
		return "", ErrVolumeNotInitialized
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(urlStr)
	if err != nil {
		slog.Error("failed to download mod", "url", urlStr, "error", err)
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("upstream returned %d", resp.StatusCode)
	}

	modsDir := filepath.Join(st.VolumePath, "mods")
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

// CopyAllServerMods copies all .jar files from server mods to client mods directory.
func (s *Service) CopyAllServerMods(serverID string) (copied, skipped, errs []string, err error) {
	st, ok := s.instance.Get(serverID)
	if !ok {
		return nil, nil, nil, ErrNotFound
	}
	if st.VolumePath == "" {
		return nil, nil, nil, ErrVolumeNotInitialized
	}

	serverDir := filepath.Join(st.VolumePath, "mods")
	clientDir := filepath.Join(st.VolumePath, "mods-client")

	uid, gid := runner.ContainerUIDGID()
	if err := runner.FixVolumeOwnership(context.Background(), s.cli, st.VolumePath, uid, gid); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", serverID, "path", st.VolumePath, "error", err)
	}
	if err := os.MkdirAll(clientDir, 0o775); err != nil {
		return nil, nil, nil, err
	}

	entries, err := os.ReadDir(serverDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, []string{}, []string{}, nil
		}
		return nil, nil, nil, err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		lower := strings.ToLower(name)
		if !strings.HasSuffix(lower, ".jar") && !strings.HasSuffix(lower, ".jar.deactivated") {
			continue
		}

		srcPath := filepath.Join(serverDir, name)
		dstPath := filepath.Join(clientDir, name)

		if _, statErr := os.Stat(dstPath); statErr == nil {
			skipped = append(skipped, name)
			continue
		}

		src, err := os.Open(srcPath)
		if err != nil {
			errs = append(errs, fmt.Sprintf("open %s: %v", name, err))
			continue
		}
		dst, err := os.Create(dstPath)
		if err != nil {
			src.Close()
			errs = append(errs, fmt.Sprintf("create %s: %v", name, err))
			continue
		}
		_, cpyErr := io.Copy(dst, src)
		src.Close()
		dst.Close()
		if cpyErr != nil {
			errs = append(errs, fmt.Sprintf("copy %s: %v", name, cpyErr))
			continue
		}
		copied = append(copied, name)
	}
	return copied, skipped, errs, nil
}
