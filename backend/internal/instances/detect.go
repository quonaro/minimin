package instances

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
)

// DetectFormat inspects a zip archive and returns its detected format.
func DetectFormat(r *zip.Reader) (Format, error) {
	hasInstanceCfg := false
	hasMmcPack := false
	hasMrpackIndex := false
	hasCurseManifest := false

	for _, f := range r.File {
		name := strings.ToLower(filepath.ToSlash(f.Name))
		base := filepath.Base(name)
		switch base {
		case "instance.cfg":
			hasInstanceCfg = true
		case "mmc-pack.json":
			hasMmcPack = true
		case "modrinth.index.json":
			hasMrpackIndex = true
		case "manifest.json":
			if isCurseManifest(f) {
				hasCurseManifest = true
			}
		}
	}

	switch {
	case hasInstanceCfg && hasMmcPack:
		return FormatPrism, nil
	case hasMrpackIndex:
		return FormatMrpack, nil
	case hasCurseManifest:
		return FormatCurseForge, nil
	default:
		return FormatPlain, nil
	}
}

// isCurseManifest reads the first part of a manifest.json to decide whether it
// is a CurseForge modpack manifest.
func isCurseManifest(f *zip.File) bool {
	rc, err := f.Open()
	if err != nil {
		return false
	}
	defer func() { _ = rc.Close() }()

	data, err := io.ReadAll(io.LimitReader(rc, 64*1024))
	if err != nil {
		return false
	}
	var manifest struct {
		ManifestType string `json:"manifestType"`
	}
	_ = json.Unmarshal(data, &manifest)
	return manifest.ManifestType == "minecraftModpack"
}

// DetectPaths scans the archive and returns the top-level directory names that
// are relevant for a server import.
func DetectPaths(r *zip.Reader, format Format) []string {
	seen := make(map[string]struct{})
	for _, f := range r.File {
		name := strings.ToLower(filepath.ToSlash(f.Name))
		parts := strings.Split(strings.TrimPrefix(name, "/"), "/")
		if len(parts) == 0 {
			continue
		}

		top := parts[0]
		switch format {
		case FormatPrism:
			if top == ".minecraft" && len(parts) > 1 {
				top = parts[1]
			}
		case FormatMrpack, FormatCurseForge:
			if top == "overrides" && len(parts) > 1 {
				top = parts[1]
			}
		}
		if isAllowedDir(top) {
			seen[top] = struct{}{}
		}
	}

	paths := make([]string, 0, len(seen))
	for d := range seen {
		paths = append(paths, d)
	}
	return paths
}

func isAllowedDir(name string) bool {
	for _, d := range allowedTopLevelDirs {
		if d == name {
			return true
		}
	}
	return false
}

// readZipEntry reads the full content of a zip entry.
func readZipEntry(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()
	return io.ReadAll(rc)
}

// findFile returns the first zip entry whose name (case-insensitive) matches
// the given base name.
func findFile(r *zip.Reader, base string) *zip.File {
	for _, f := range r.File {
		if strings.EqualFold(filepath.Base(f.Name), base) {
			return f
		}
	}
	return nil
}

// firstNonEmpty returns the first non-empty string.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// trimPrefix removes the first n path components from name.
func trimPrefix(name string, n int) string {
	parts := strings.Split(strings.TrimPrefix(name, "/"), "/")
	if len(parts) <= n {
		return ""
	}
	return strings.Join(parts[n:], "/")
}

// peekBytes reads the first n bytes of a zip entry.
func peekBytes(f *zip.File, n int) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()
	return io.ReadAll(io.LimitReader(rc, int64(n)))
}

// hasPrefixBytes reports whether the first len(prefix) bytes of data equal prefix.
func hasPrefixBytes(data, prefix []byte) bool {
	return len(data) >= len(prefix) && bytes.Equal(data[:len(prefix)], prefix)
}
