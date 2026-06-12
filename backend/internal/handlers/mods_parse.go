package handlers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

// modParseCacheKey uniquely identifies a parsed .jar by its path, size and mtime.
type modParseCacheKey struct {
	path  string
	size  int64
	mtime int64 // unix nano
}

var (
	modParseCacheMu sync.RWMutex
	modParseCache   = make(map[modParseCacheKey]*ModInfo)
)

// parseModInfoCached wraps ParseModInfo with a simple in-memory cache.
// If the file hasn't changed (same path, size, mtime) the cached result is returned.
func parseModInfoCached(path string, size int64, mtime time.Time) (*ModInfo, error) {
	key := modParseCacheKey{path, size, mtime.UnixNano()}
	modParseCacheMu.RLock()
	if info, ok := modParseCache[key]; ok {
		modParseCacheMu.RUnlock()
		return info, nil
	}
	modParseCacheMu.RUnlock()

	info, err := ParseModInfo(path, size)
	if err != nil {
		return nil, err
	}

	modParseCacheMu.Lock()
	modParseCache[key] = info
	modParseCacheMu.Unlock()
	return info, nil
}

// ModInfo holds extracted metadata from a mod .jar file.
type ModInfo struct {
	Filename    string `json:"filename"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	ModID       string `json:"modid"`
	Authors     string `json:"authors"`
	Size        int64  `json:"size"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Enabled     bool   `json:"enabled"`
	Environment string `json:"environment"`
}

// ParseModInfo reads a .jar and extracts mod metadata.
func ParseModInfo(path string, size int64) (*ModInfo, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open jar: %w", err)
	}
	defer func() { _ = zr.Close() }()

	info := &ModInfo{Filename: filepath.Base(path), Size: size}

	for _, f := range zr.File {
		if f.Name == "fabric.mod.json" {
			if err := parseFabricModJSON(f, info); err == nil {
				return info, nil
			}
		}
		if f.Name == "META-INF/mods.toml" {
			if err := parseForgeModsTOML(f, info); err == nil {
				return info, nil
			}
		}
		if f.Name == "mcmod.info" {
			if err := parseMcmodInfo(f, info); err == nil {
				return info, nil
			}
		}
	}

	// Fallback: use filename as name
	info.Name = strings.TrimSuffix(filepath.Base(path), ".jar")
	return info, nil
}

func parseFabricModJSON(zf *zip.File, info *ModInfo) error {
	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	var data struct {
		ID          string `json:"id"`
		Version     string `json:"version"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Authors     []any  `json:"authors"`
		Icon        string `json:"icon"`
		Environment string `json:"environment"`
	}
	if err := json.NewDecoder(rc).Decode(&data); err != nil {
		return err
	}
	info.ModID = data.ID
	info.Version = data.Version
	info.Name = data.Name
	if info.Name == "" {
		info.Name = data.ID
	}
	info.Description = data.Description
	info.Authors = joinAuthors(data.Authors)
	info.Icon = data.Icon
	info.Environment = data.Environment
	return nil
}

func parseForgeModsTOML(zf *zip.File, info *ModInfo) error {
	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	var root struct {
		Mods []struct {
			ModID       string `toml:"modId"`
			Version     string `toml:"version"`
			Display     string `toml:"displayName"`
			Authors     string `toml:"authors"`
			Description string `toml:"description"`
			LogoFile    string `toml:"logoFile"`
		} `toml:"mods"`
	}
	if _, err := toml.NewDecoder(rc).Decode(&root); err != nil {
		return err
	}
	if len(root.Mods) == 0 {
		return fmt.Errorf("no mods block")
	}
	m := root.Mods[0]
	info.ModID = m.ModID
	info.Version = m.Version
	info.Name = m.Display
	if info.Name == "" {
		info.Name = m.ModID
	}
	info.Authors = m.Authors
	info.Description = m.Description
	info.Icon = m.LogoFile
	return nil
}

func parseMcmodInfo(zf *zip.File, info *ModInfo) error {
	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	var list []struct {
		ModID       string   `json:"modid"`
		Name        string   `json:"name"`
		Version     string   `json:"version"`
		Authors     []string `json:"authorList"`
		Description string   `json:"description"`
		LogoFile    string   `json:"logoFile"`
	}
	if err := json.NewDecoder(rc).Decode(&list); err != nil {
		return err
	}
	if len(list) == 0 {
		return fmt.Errorf("empty mcmod.info")
	}
	m := list[0]
	info.ModID = m.ModID
	info.Version = m.Version
	info.Name = m.Name
	if info.Name == "" {
		info.Name = m.ModID
	}
	info.Authors = strings.Join(m.Authors, ", ")
	info.Description = m.Description
	info.Icon = m.LogoFile
	return nil
}

func joinAuthors(v []any) string {
	var out []string
	for _, a := range v {
		switch val := a.(type) {
		case string:
			out = append(out, val)
		case map[string]any:
			if name, ok := val["name"].(string); ok {
				out = append(out, name)
			}
		}
	}
	return strings.Join(out, ", ")
}

// createFile ensures the parent directory exists and creates the file.
func createFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return os.Create(path)
}

// extractZipJars extracts every .jar from a zip archive into destDir.
// It validates that each extracted file name ends with .jar.
func extractZipJars(zipPath, destDir string) ([]string, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	defer func() { _ = zr.Close() }()

	var extracted []string
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := strings.ToLower(f.Name)
		if !strings.HasSuffix(name, ".jar") {
			continue
		}
		// Prevent path traversal: flatten to basename inside destDir
		outPath := filepath.Join(destDir, filepath.Base(f.Name))
		rc, err := f.Open()
		if err != nil {
			return extracted, fmt.Errorf("failed to open %s: %w", f.Name, err)
		}
		out, err := createFile(outPath)
		if err != nil {
			_ = rc.Close()
			return extracted, fmt.Errorf("failed to create %s: %w", outPath, err)
		}
		_, copyErr := io.Copy(out, rc)
		_ = rc.Close()
		_ = out.Close()
		if copyErr != nil {
			return extracted, fmt.Errorf("failed to write %s: %w", outPath, copyErr)
		}
		extracted = append(extracted, f.Name)
	}
	return extracted, nil
}
