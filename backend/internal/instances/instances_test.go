package instances

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func createZip(t *testing.T, entries map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDetectFormat(t *testing.T) {
	cases := []struct {
		name     string
		entries  map[string]string
		expected Format
	}{
		{
			name: "prism",
			entries: map[string]string{
				"instance.cfg":  "name=Test\n",
				"mmc-pack.json": `{"components":[]}`,
			},
			expected: FormatPrism,
		},
		{
			name: "mrpack",
			entries: map[string]string{
				"modrinth.index.json": `{"formatVersion":1,"game":"minecraft","versionId":"1","name":"Test"}`,
			},
			expected: FormatMrpack,
		},
		{
			name: "curseforge",
			entries: map[string]string{
				"manifest.json": `{"manifestType":"minecraftModpack","minecraft":{"version":"1.20.1"}}`,
			},
			expected: FormatCurseForge,
		},
		{
			name: "plain",
			entries: map[string]string{
				"mods/foo.jar": "jar",
			},
			expected: FormatPlain,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := createZip(t, tc.entries)
			zr, _, err := openZip(path)
			if err != nil {
				t.Fatal(err)
			}
			defer zr.Close()
			format, err := DetectFormat(zr.Reader)
			if err != nil {
				t.Fatal(err)
			}
			if format != tc.expected {
				t.Fatalf("expected %s, got %s", tc.expected, format)
			}
		})
	}
}

func TestParsePrism(t *testing.T) {
	entries := map[string]string{
		"instance.cfg": "name=My Prism Instance\n",
		"mmc-pack.json": `{
			"components": [
				{"uid": "net.minecraft", "version": "1.20.1"},
				{"uid": "net.fabricmc.fabric-loader", "version": "0.15.0"}
			]
		}`,
	}
	path := createZip(t, entries)
	zr, _, err := openZip(path)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()

	meta, err := ParseMetadata(zr.Reader, FormatPrism)
	if err != nil {
		t.Fatal(err)
	}
	if meta.InstanceName != "My Prism Instance" {
		t.Errorf("instance name: got %q", meta.InstanceName)
	}
	if meta.GameVersion != "1.20.1" {
		t.Errorf("game version: got %q", meta.GameVersion)
	}
	if meta.EngineType != "FABRIC" {
		t.Errorf("engine type: got %q", meta.EngineType)
	}
	if meta.LoaderVersion != "0.15.0" {
		t.Errorf("loader version: got %q", meta.LoaderVersion)
	}
}

func TestExtractPrism(t *testing.T) {
	entries := map[string]string{
		"instance.cfg":            "name=Test\n",
		"mmc-pack.json":           `{"components":[]}`,
		".minecraft/mods/foo.jar": "jar",
		".minecraft/mods/bar.jar": "jar",
	}
	path := createZip(t, entries)
	zr, _, err := openZip(path)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()

	target := t.TempDir()
	if err := Extract(zr.Reader, FormatPrism, ExtractOptions{TargetDir: target}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(target, "mods", "foo.jar")); err != nil {
		t.Errorf("expected mods/foo.jar: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "mods", "bar.jar")); err != nil {
		t.Errorf("expected mods/bar.jar: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "mods-client", "foo.jar")); err != nil {
		t.Errorf("expected mods-client/foo.jar: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "mods-client", "bar.jar")); err != nil {
		t.Errorf("expected mods-client/bar.jar: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "config", "bar.cfg")); err == nil {
		t.Errorf("config/ should not be extracted")
	}
	if _, err := os.Stat(filepath.Join(target, "instance.cfg")); err == nil {
		t.Errorf("instance.cfg should not be extracted")
	}
}

func TestExtractMrpack(t *testing.T) {
	entries := map[string]string{
		"modrinth.index.json":    `{"formatVersion":1,"game":"minecraft","versionId":"1","name":"Test"}`,
		"overrides/mods/foo.jar": "jar",
	}
	path := createZip(t, entries)
	zr, _, err := openZip(path)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()

	target := t.TempDir()
	if err := Extract(zr.Reader, FormatMrpack, ExtractOptions{TargetDir: target}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(target, "mods", "foo.jar")); err != nil {
		t.Errorf("expected mods/foo.jar: %v", err)
	}
}

func TestExtractRejectsTraversal(t *testing.T) {
	entries := map[string]string{
		"mods/foo.jar":     "jar",
		"mods/../evil.txt": "evil",
	}
	path := createZip(t, entries)
	zr, _, err := openZip(path)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()

	target := t.TempDir()
	if err := Extract(zr.Reader, FormatPlain, ExtractOptions{TargetDir: target}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(target, "evil.txt")); err == nil {
		t.Errorf("path traversal file should not be extracted")
	}
}

func TestSafeJoin(t *testing.T) {
	base := t.TempDir()
	cases := []struct {
		rel     string
		allowed bool
	}{
		{"mods/foo.jar", true},
		{"../foo.jar", false},
		{"mods/../../foo.jar", false},
	}
	for _, tc := range cases {
		path, err := safeJoin(base, tc.rel)
		if tc.allowed {
			if err != nil {
				t.Errorf("%q: unexpected error: %v", tc.rel, err)
			}
			if path == "" {
				t.Errorf("%q: expected path", tc.rel)
			}
		} else {
			if err == nil {
				t.Errorf("%q: expected error", tc.rel)
			}
		}
	}
}

func TestStoreSaveAndGet(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	token, err := store.Save("test.zip", 4, func(dst *os.File) error {
		_, err := dst.Write([]byte("test"))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	entry, err := store.Get(token)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Size != 4 {
		t.Errorf("size: got %d", entry.Size)
	}
	data, err := os.ReadFile(entry.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte("test")) {
		t.Errorf("stored data mismatch")
	}
}

func TestServicePrepareUsesStoreToken(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(store)

	path := createZip(t, map[string]string{
		"instance.cfg":          "name=SvcTest\n",
		"mmc-pack.json":         `{"components":[{"uid":"net.minecraft","version":"1.20.1"}]}`,
		".minecraft/mods/a.jar": "jar",
	})
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	token, err := svc.Save("svc.zip", int64(len(data)), bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	meta, err := svc.Prepare(token)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Format != FormatPrism {
		t.Errorf("format: got %s", meta.Format)
	}
	if meta.InstanceName != "SvcTest" {
		t.Errorf("instance name: got %q", meta.InstanceName)
	}
}

func TestDetectWorlds(t *testing.T) {
	entries := map[string]string{
		".minecraft/saves/MyWorld/level.dat": "dat",
		".minecraft/saves/Other/level.dat":   "dat",
		".minecraft/mods/foo.jar":            "jar",
	}
	path := createZip(t, entries)
	zr, _, err := openZip(path)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()

	worlds := DetectWorlds(zr.Reader, FormatPrism)
	if len(worlds) != 2 {
		t.Fatalf("worlds: got %d, want 2", len(worlds))
	}
	names := map[string]bool{}
	for _, w := range worlds {
		names[w.Name] = true
	}
	if !names["MyWorld"] || !names["Other"] {
		t.Errorf("expected MyWorld and Other, got %v", names)
	}
}

func TestExtractWorld(t *testing.T) {
	entries := map[string]string{
		".minecraft/saves/MyWorld/level.dat":        "dat",
		".minecraft/saves/MyWorld/region/r.0.0.mca": "mca",
		".minecraft/mods/foo.jar":                   "jar",
	}
	path := createZip(t, entries)
	zr, _, err := openZip(path)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()

	target := t.TempDir()
	if err := Extract(zr.Reader, FormatPrism, ExtractOptions{
		TargetDir: target,
		World:     ".minecraft/saves/MyWorld",
		LevelName: "MyWorld",
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(target, "MyWorld", "level.dat")); err != nil {
		t.Errorf("expected MyWorld/level.dat: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "MyWorld", "region", "r.0.0.mca")); err != nil {
		t.Errorf("expected MyWorld/region/r.0.0.mca: %v", err)
	}
}
