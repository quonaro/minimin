package instances

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Extract unpacks an archive into a target directory, applying format-specific
// path normalization and an allowlist of top-level directories.
// It also mirrors server mods to mods-client so the client archive is ready.
func Extract(r *zip.Reader, format Format, targetDir string) error {
	strip := stripComponents(format)
	allowed := map[string]struct{}{}
	for _, d := range allowedTopLevelDirs {
		allowed[d] = struct{}{}
	}

	for _, f := range r.File {
		name := filepath.ToSlash(f.Name)
		if name == "" {
			continue
		}

		rel, ok := normalizeAndAllow(name, strip, allowed)
		if !ok {
			continue
		}
		if rel == "" {
			if f.FileInfo().IsDir() {
				continue
			}
			// Skip files at the archive root that are not in an allowed dir.
			continue
		}

		absPath, err := safeJoin(targetDir, rel)
		if err != nil {
			return fmt.Errorf("invalid path %q: %w", rel, err)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(absPath, 0o775); err != nil {
				return fmt.Errorf("create directory %s: %w", absPath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(absPath), 0o775); err != nil {
			return fmt.Errorf("create directory %s: %w", filepath.Dir(absPath), err)
		}
		if err := writeZipEntry(f, absPath); err != nil {
			return fmt.Errorf("extract %s: %w", rel, err)
		}
	}

	if err := mirrorModsToClientMods(targetDir); err != nil {
		return fmt.Errorf("mirror mods to client: %w", err)
	}
	return nil
}

// mirrorModsToClientMods copies .jar files from the server mods directory into
// the mods-client directory so the client archive is immediately available.
func mirrorModsToClientMods(targetDir string) error {
	modsDir, err := safeJoin(targetDir, "mods")
	if err != nil {
		return err
	}
	clientDir, err := safeJoin(targetDir, "mods-client")
	if err != nil {
		return err
	}

	info, err := os.Stat(modsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}

	if err := os.MkdirAll(clientDir, 0o775); err != nil {
		return err
	}

	entries, err := os.ReadDir(modsDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if !strings.HasSuffix(name, ".jar") && !strings.HasSuffix(name, ".jar.disabled") {
			continue
		}
		src := filepath.Join(modsDir, e.Name())
		dst := filepath.Join(clientDir, e.Name())
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("copy %s: %w", e.Name(), err)
		}
	}
	return nil
}

// copyFile copies a file from src to dst, preserving mode.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o664)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// stripComponents returns how many leading path components to strip for each
// format.
func stripComponents(format Format) int {
	switch format {
	case FormatMrpack, FormatCurseForge:
		// overrides/mods -> mods
		return 1
	case FormatPrism:
		// .minecraft/mods -> mods
		return 1
	default:
		return 0
	}
}

// normalizeAndAllow checks whether a zip entry path is safe and maps it to a
// relative path inside the target directory. It returns the relative path and
// true when the entry should be extracted.
func normalizeAndAllow(name string, strip int, allowed map[string]struct{}) (string, bool) {
	// Clean and split the path.
	parts := strings.Split(strings.TrimPrefix(name, "/"), "/")
	if len(parts) == 0 {
		return "", false
	}

	// Strip leading components.
	if len(parts) <= strip {
		return "", true
	}
	parts = parts[strip:]

	// Reject path traversal attempts.
	for _, p := range parts {
		if p == ".." || p == "" && len(parts) > 1 {
			return "", false
		}
	}

	// The first remaining component must be in the allowlist.
	if len(parts) == 0 {
		return "", true
	}
	if _, ok := allowed[strings.ToLower(parts[0])]; !ok {
		return "", false
	}

	return strings.Join(parts, "/"), true
}

// safeJoin joins a base directory with a relative path and verifies the result
// stays within the base directory.
func safeJoin(base, rel string) (string, error) {
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(filepath.Join(absBase, rel))
	if err != nil {
		return "", err
	}
	if absPath != absBase && !strings.HasPrefix(absPath, absBase+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes base directory")
	}
	return absPath, nil
}

// writeZipEntry writes a single zip file to disk with the correct mode.
func writeZipEntry(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o664)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, rc); err != nil {
		return err
	}
	return out.Close()
}
