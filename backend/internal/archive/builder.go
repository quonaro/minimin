package archive

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func createZipArchive(w io.Writer, filesByType map[string][]string) error {
	zw := zip.NewWriter(w)
	defer func() { _ = zw.Close() }()

	zipDirMap := map[string]string{
		"mods":          "mods/",
		"resourcepacks": "resourcepacks/",
		"shaderpacks":   "shaderpacks/",
	}

	for t, files := range filesByType {
		prefix := zipDirMap[t]
		for _, path := range files {
			info, err := os.Stat(path)
			if err != nil {
				continue
			}
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				continue
			}
			header.Name = prefix + filepath.Base(path)
			header.Method = zip.Deflate

			entry, err := zw.CreateHeader(header)
			if err != nil {
				continue
			}
			f, err := os.Open(path)
			if err != nil {
				continue
			}
			_, _ = io.Copy(entry, f)
			_ = f.Close()
		}
	}
	return nil
}

func createMrpackArchive(w io.Writer, packName string, filesByType map[string][]string) error {
	zw := zip.NewWriter(w)
	defer func() { _ = zw.Close() }()

	type mrpackFile struct {
		Path   string            `json:"path"`
		Hashes map[string]string `json:"hashes"`
		Size   int64             `json:"size"`
	}
	type mrpackIndex struct {
		FormatVersion int          `json:"formatVersion"`
		Game          string       `json:"game"`
		VersionID     string       `json:"versionId"`
		Name          string       `json:"name"`
		Files         []mrpackFile `json:"files"`
	}

	index := mrpackIndex{
		FormatVersion: 1,
		Game:          "minecraft",
		VersionID:     "1.0.0",
		Name:          packName,
		Files:         []mrpackFile{},
	}

	dirMap := map[string]string{
		"mods":          "overrides/mods/",
		"resourcepacks": "overrides/resourcepacks/",
		"shaderpacks":   "overrides/shaderpacks/",
	}

	for t, files := range filesByType {
		prefix := dirMap[t]
		for _, path := range files {
			name := filepath.Base(path)
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			index.Files = append(index.Files, mrpackFile{
				Path:   prefix + name,
				Hashes: map[string]string{},
				Size:   info.Size(),
			})

			header := &zip.FileHeader{Name: prefix + name, Method: zip.Deflate, Modified: info.ModTime()}
			entry, err := zw.CreateHeader(header)
			if err != nil {
				continue
			}
			f, err := os.Open(path)
			if err != nil {
				continue
			}
			_, _ = io.Copy(entry, f)
			_ = f.Close()
		}
	}

	indexData, _ := json.MarshalIndent(index, "", "  ")
	entry, err := zw.Create("modrinth.index.json")
	if err != nil {
		return err
	}
	_, _ = entry.Write(indexData)
	return nil
}

func createCurseForgeArchive(w io.Writer, packName, gameVersion, engineType, loaderVersion string, filesByType map[string][]string) error {
	zw := zip.NewWriter(w)
	defer func() { _ = zw.Close() }()

	engine := strings.ToLower(engineType)
	loaderID := engine
	if loaderVersion != "" {
		loaderID = engine + "-" + loaderVersion
	}

	manifest := map[string]any{
		"minecraft": map[string]any{
			"version":    gameVersion,
			"modLoaders": []map[string]any{{"id": loaderID, "primary": true}},
		},
		"manifestType":    "minecraftModpack",
		"manifestVersion": 1,
		"name":            packName,
		"version":         "1.0.0",
		"author":          "MiniMin",
		"files":           []any{},
		"overrides":       "overrides",
	}

	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	entry, err := zw.Create("manifest.json")
	if err != nil {
		return err
	}
	_, _ = entry.Write(manifestData)

	dirMap := map[string]string{
		"mods":          "overrides/mods/",
		"resourcepacks": "overrides/resourcepacks/",
		"shaderpacks":   "overrides/shaderpacks/",
	}

	for t, files := range filesByType {
		prefix := dirMap[t]
		for _, path := range files {
			name := filepath.Base(path)
			info, err := os.Stat(path)
			if err != nil {
				continue
			}
			header := &zip.FileHeader{Name: prefix + name, Method: zip.Deflate, Modified: info.ModTime()}
			entry, err := zw.CreateHeader(header)
			if err != nil {
				continue
			}
			f, err := os.Open(path)
			if err != nil {
				continue
			}
			_, _ = io.Copy(entry, f)
			_ = f.Close()
		}
	}
	return nil
}

func createPrismArchive(w io.Writer, packName, gameVersion, engineType, loaderVersion, token, baseURL string, filesByType map[string][]string) error {
	zw := zip.NewWriter(w)
	defer func() { _ = zw.Close() }()

	cfg := fmt.Sprintf("[General]\nname=%s\nInstanceType=OneSix\niconKey=default\n", packName)
	entry, err := zw.Create("instance.cfg")
	if err != nil {
		return err
	}
	_, _ = entry.Write([]byte(cfg))

	engine := strings.ToLower(engineType)

	minecraftComponent := map[string]any{
		"cachedName":    "Minecraft",
		"cachedVersion": gameVersion,
		"important":     true,
		"uid":           "net.minecraft",
		"version":       gameVersion,
	}

	var components []map[string]any

	switch {
	case engine == "fabric" && loaderVersion != "":
		components = []map[string]any{
			minecraftComponent,
			{
				"cachedName":     "Intermediary Mappings",
				"cachedRequires": []map[string]any{{"equals": gameVersion, "uid": "net.minecraft"}},
				"cachedVersion":  gameVersion,
				"dependencyOnly": true,
				"uid":            "net.fabricmc.intermediary",
				"version":        gameVersion,
			},
			{
				"cachedName":     "Fabric Loader",
				"cachedRequires": []map[string]any{{"uid": "net.fabricmc.intermediary"}},
				"cachedVersion":  loaderVersion,
				"important":      true,
				"uid":            "net.fabricmc.fabric-loader",
				"version":        loaderVersion,
			},
		}
	case engine != "vanilla" && loaderVersion != "":
		var uid, cachedName string
		switch engine {
		case "forge":
			uid = "net.minecraftforge"
			cachedName = "Forge"
		case "neoforge":
			uid = "net.neoforged"
			cachedName = "NeoForge"
		default:
			uid = engine
			cachedName = engine
		}
		components = []map[string]any{
			minecraftComponent,
			{
				"cachedName":     cachedName,
				"cachedRequires": []map[string]any{{"equals": gameVersion, "uid": "net.minecraft"}},
				"cachedVersion":  loaderVersion,
				"important":      true,
				"uid":            uid,
				"version":        loaderVersion,
			},
		}
	default:
		components = []map[string]any{minecraftComponent}
	}

	mmcPack := map[string]any{
		"components":    components,
		"formatVersion": 1,
	}
	mmcData, _ := json.MarshalIndent(mmcPack, "", "  ")
	entry, err = zw.Create("mmc-pack.json")
	if err != nil {
		return err
	}
	_, _ = entry.Write(mmcData)

	marker := map[string]any{
		"serverId":   packName,
		"token":      token,
		"baseUrl":    baseURL,
		"lastSyncAt": time.Now().UTC().Format(time.RFC3339),
	}
	markerData, _ := json.MarshalIndent(marker, "", "  ")
	entry, err = zw.Create(".minimin.json")
	if err != nil {
		return err
	}
	_, _ = entry.Write(markerData)

	dirMap := map[string]string{
		"mods":          ".minecraft/mods/",
		"resourcepacks": ".minecraft/resourcepacks/",
		"shaderpacks":   ".minecraft/shaderpacks/",
	}

	for t, files := range filesByType {
		prefix := dirMap[t]
		for _, path := range files {
			name := filepath.Base(path)
			info, err := os.Stat(path)
			if err != nil {
				continue
			}
			header := &zip.FileHeader{Name: prefix + name, Method: zip.Deflate, Modified: info.ModTime()}
			entry, err := zw.CreateHeader(header)
			if err != nil {
				continue
			}
			f, err := os.Open(path)
			if err != nil {
				continue
			}
			_, _ = io.Copy(entry, f)
			_ = f.Close()
		}
	}
	return nil
}
