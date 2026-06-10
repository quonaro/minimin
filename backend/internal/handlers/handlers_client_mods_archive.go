package handlers

import (
	"archive/zip"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ArchiveToken represents a generated archive with expiration.
type ArchiveToken struct {
	Token          string    `json:"token"`
	ServerID       string    `json:"serverId"`
	ServerName     string    `json:"serverName"`
	ZipPath        string    `json:"zipPath"`
	MrpackPath     string    `json:"mrpackPath"`
	CurseForgePath string    `json:"curseForgePath"`
	PrismPath      string    `json:"prismPath"`
	ExpiresAt      time.Time `json:"expiresAt"`
	CreatedAt      time.Time `json:"createdAt"`
	Formats        []string  `json:"formats"`
}

var (
	archiveTokens   = make(map[string]*ArchiveToken)
	archiveTokensMu sync.RWMutex
)

func init() {
	go cleanupExpiredArchives()
}

func cleanupExpiredArchives() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		archiveTokensMu.Lock()
		now := time.Now()
		for token, entry := range archiveTokens {
			if now.After(entry.ExpiresAt) {
				_ = os.Remove(entry.ZipPath)
				_ = os.Remove(entry.MrpackPath)
				_ = os.Remove(entry.CurseForgePath)
				_ = os.Remove(entry.PrismPath)
				delete(archiveTokens, token)
			}
		}
		archiveTokensMu.Unlock()
	}
}

func generateToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// HandleCreateClientArchive generates .zip and/or .mrpack from client mods.
func (h *Handler) HandleCreateClientArchive(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if s.VolumePath == "" {
		jsonError(w, "server volume not initialized", http.StatusConflict)
		return
	}

	var req struct {
		Formats []string `json:"formats"` // "zip", "mrpack", "curseforge", "prism"
		TTL     int      `json:"ttl"`     // hours
		Include []string `json:"include"` // "mods", "resourcepacks", "shaderpacks"
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.Formats) == 0 {
		req.Formats = []string{"zip"}
	}
	if req.TTL <= 0 {
		req.TTL = 24
	}
	if len(req.Include) == 0 {
		req.Include = []string{"mods"}
	}

	typeDirMap := map[string]string{
		"mods":          filepath.Join(s.VolumePath, "mods-client"),
		"resourcepacks": filepath.Join(s.VolumePath, "resourcepacks"),
		"shaderpacks":   filepath.Join(s.VolumePath, "shaderpacks"),
	}
	extMap := map[string]string{
		"mods":          ".jar",
		"resourcepacks": ".zip",
		"shaderpacks":   ".zip",
	}

	filesByType := make(map[string][]string)
	for _, t := range req.Include {
		dir, ok := typeDirMap[t]
		if !ok {
			continue
		}
		ext := extMap[t]
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			jsonError(w, fmt.Sprintf("failed to read %s directory", t), http.StatusInternalServerError)
			return
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := strings.ToLower(e.Name())
			if strings.HasSuffix(name, ext) || strings.HasSuffix(name, ext+".disabled") {
				filesByType[t] = append(filesByType[t], filepath.Join(dir, e.Name()))
			}
		}
	}

	if len(filesByType) == 0 {
		jsonError(w, "no files found for selected types", http.StatusNotFound)
		return
	}

	tmpDir := filepath.Join(os.TempDir(), "webui-archives")
	_ = os.MkdirAll(tmpDir, 0o755)

	token := generateToken()
	expiresAt := time.Now().Add(time.Duration(req.TTL) * time.Hour)

	archive := &ArchiveToken{
		Token:      token,
		ServerID:   id,
		ServerName: s.ServerID,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		Formats:    req.Formats,
	}

	for _, format := range req.Formats {
		switch format {
		case "zip":
			zipPath := filepath.Join(tmpDir, token+".zip")
			if err := createZipArchive(zipPath, filesByType); err != nil {
				jsonError(w, fmt.Sprintf("failed to create zip: %v", err), http.StatusInternalServerError)
				return
			}
			archive.ZipPath = zipPath
		case "mrpack":
			mrpackPath := filepath.Join(tmpDir, token+".mrpack")
			if err := createMrpackArchive(mrpackPath, s.ServerID, filesByType); err != nil {
				jsonError(w, fmt.Sprintf("failed to create mrpack: %v", err), http.StatusInternalServerError)
				return
			}
			archive.MrpackPath = mrpackPath
		case "curseforge":
			cfPath := filepath.Join(tmpDir, token+"-curseforge.zip")
			if err := createCurseForgeArchive(cfPath, s.ServerID, s.GameVersion, s.EngineType, s.LoaderVersion, filesByType); err != nil {
				jsonError(w, fmt.Sprintf("failed to create curseforge archive: %v", err), http.StatusInternalServerError)
				return
			}
			archive.CurseForgePath = cfPath
		case "prism":
			prismPath := filepath.Join(tmpDir, token+"-prism.zip")
			if err := createPrismArchive(prismPath, s.ServerID, s.GameVersion, s.EngineType, s.LoaderVersion, filesByType); err != nil {
				jsonError(w, fmt.Sprintf("failed to create prism archive: %v", err), http.StatusInternalServerError)
				return
			}
			archive.PrismPath = prismPath
		}
	}

	archiveTokensMu.Lock()
	archiveTokens[token] = archive
	archiveTokensMu.Unlock()

	jsonResponse(w, map[string]any{
		"token":      token,
		"expiresAt":  expiresAt.Format(time.RFC3339),
		"serverName": s.ServerID,
		"formats":    req.Formats,
	})
}

func createZipArchive(zipPath string, filesByType map[string][]string) error {
	zf, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer func() { _ = zf.Close() }()

	zw := zip.NewWriter(zf)
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

			w, err := zw.CreateHeader(header)
			if err != nil {
				continue
			}
			f, err := os.Open(path)
			if err != nil {
				continue
			}
			_, _ = io.Copy(w, f)
			_ = f.Close()
		}
	}
	return nil
}

func createMrpackArchive(mrpackPath, packName string, filesByType map[string][]string) error {
	zf, err := os.Create(mrpackPath)
	if err != nil {
		return err
	}
	defer func() { _ = zf.Close() }()

	zw := zip.NewWriter(zf)
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
			f, err := os.Open(path)
			if err != nil {
				continue
			}
			_ = f.Close()

			index.Files = append(index.Files, mrpackFile{
				Path:   prefix + name,
				Hashes: map[string]string{},
				Size:   info.Size(),
			})

			header := &zip.FileHeader{Name: prefix + name, Method: zip.Deflate, Modified: info.ModTime()}
			w, err := zw.CreateHeader(header)
			if err != nil {
				continue
			}
			f2, err := os.Open(path)
			if err != nil {
				continue
			}
			_, _ = io.Copy(w, f2)
			_ = f2.Close()
		}
	}

	indexData, _ := json.MarshalIndent(index, "", "  ")
	w, err := zw.Create("modrinth.index.json")
	if err != nil {
		return err
	}
	_, _ = w.Write(indexData)
	return nil
}

func createCurseForgeArchive(zipPath, packName, gameVersion, engineType, loaderVersion string, filesByType map[string][]string) error {
	zf, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer func() { _ = zf.Close() }()

	zw := zip.NewWriter(zf)
	defer func() { _ = zw.Close() }()

	engine := strings.ToLower(engineType)
	loaderID := engine
	if loaderVersion != "" {
		loaderID = engine + "-" + loaderVersion
	}

	manifest := map[string]any{
		"minecraft": map[string]any{
			"version": gameVersion,
			"modLoaders": []map[string]any{
				{"id": loaderID, "primary": true},
			},
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
	w, err := zw.Create("manifest.json")
	if err != nil {
		return err
	}
	_, _ = w.Write(manifestData)

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
			w, err := zw.CreateHeader(header)
			if err != nil {
				continue
			}
			f, err := os.Open(path)
			if err != nil {
				continue
			}
			_, _ = io.Copy(w, f)
			_ = f.Close()
		}
	}
	return nil
}

func createPrismArchive(zipPath, packName, gameVersion, engineType, loaderVersion string, filesByType map[string][]string) error {
	zf, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer func() { _ = zf.Close() }()

	zw := zip.NewWriter(zf)
	defer func() { _ = zw.Close() }()

	// instance.cfg
	cfg := fmt.Sprintf("[General]\nname=%s\nInstanceType=OneSix\niconKey=default\n", packName)
	w, err := zw.Create("instance.cfg")
	if err != nil {
		return err
	}
	_, _ = w.Write([]byte(cfg))

	// mmc-pack.json
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
	w, err = zw.Create("mmc-pack.json")
	if err != nil {
		return err
	}
	_, _ = w.Write(mmcData)

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
			w, err := zw.CreateHeader(header)
			if err != nil {
				continue
			}
			f, err := os.Open(path)
			if err != nil {
				continue
			}
			_, _ = io.Copy(w, f)
			_ = f.Close()
		}
	}
	return nil
}

// HandleDownloadClientArchive serves a generated archive by token (public, no auth).
func (h *Handler) HandleDownloadClientArchive(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "zip"
	}

	archiveTokensMu.RLock()
	archive, ok := archiveTokens[token]
	archiveTokensMu.RUnlock()

	if !ok || time.Now().After(archive.ExpiresAt) {
		jsonError(w, "archive not found or expired", http.StatusNotFound)
		return
	}

	var filePath, contentType string
	switch format {
	case "zip":
		filePath = archive.ZipPath
		contentType = "application/zip"
	case "mrpack":
		filePath = archive.MrpackPath
		contentType = "application/octet-stream"
	case "curseforge":
		filePath = archive.CurseForgePath
		contentType = "application/zip"
	case "prism":
		filePath = archive.PrismPath
		contentType = "application/zip"
	default:
		jsonError(w, "invalid format", http.StatusBadRequest)
		return
	}

	if filePath == "" {
		jsonError(w, "requested format not available", http.StatusNotFound)
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		jsonError(w, "failed to open archive", http.StatusInternalServerError)
		return
	}
	defer func() { _ = f.Close() }()

	stat, _ := f.Stat()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s-%s.%s\"", archive.ServerName, archive.CreatedAt.Format("20060102"), format))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, f)
}

// HandleGetClientArchiveInfo returns metadata about an archive (public, no auth).
func (h *Handler) HandleGetClientArchiveInfo(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	archiveTokensMu.RLock()
	archive, ok := archiveTokens[token]
	archiveTokensMu.RUnlock()

	if !ok || time.Now().After(archive.ExpiresAt) {
		jsonError(w, "archive not found or expired", http.StatusNotFound)
		return
	}

	var formats []string
	if archive.ZipPath != "" {
		formats = append(formats, "zip")
	}
	if archive.MrpackPath != "" {
		formats = append(formats, "mrpack")
	}
	if archive.CurseForgePath != "" {
		formats = append(formats, "curseforge")
	}
	if archive.PrismPath != "" {
		formats = append(formats, "prism")
	}

	jsonResponse(w, map[string]any{
		"token":      token,
		"serverName": archive.ServerName,
		"expiresAt":  archive.ExpiresAt.Format(time.RFC3339),
		"createdAt":  archive.CreatedAt.Format(time.RFC3339),
		"formats":    formats,
	})
}
