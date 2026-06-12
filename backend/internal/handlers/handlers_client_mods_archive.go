package handlers

import (
	"archive/zip"
	"context"
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
	Token          string         `json:"token"`
	ServerID       string         `json:"serverId"`
	ServerName     string         `json:"serverName"`
	Include        []string       `json:"include"`
	ExpiresAt      time.Time      `json:"expiresAt"`
	CreatedAt      time.Time      `json:"createdAt"`
	DownloadCounts map[string]int `json:"downloadCounts"`
	TotalDownloads int            `json:"totalDownloads"`
}

var (
	archiveTokens   = make(map[string]*ArchiveToken)
	archiveTokensMu sync.RWMutex
)

type archiveMetaEntry struct {
	Token          string         `json:"token"`
	ServerID       string         `json:"serverId"`
	ServerName     string         `json:"serverName"`
	Include        []string       `json:"include"`
	ExpiresAt      string         `json:"expiresAt"`
	CreatedAt      string         `json:"createdAt"`
	DownloadCounts map[string]int `json:"downloadCounts"`
	TotalDownloads int            `json:"totalDownloads"`
}

func (a *ArchiveToken) toMetaEntry() archiveMetaEntry {
	return archiveMetaEntry{
		Token:          a.Token,
		ServerID:       a.ServerID,
		ServerName:     a.ServerName,
		Include:        a.Include,
		ExpiresAt:      a.ExpiresAt.Format(time.RFC3339),
		CreatedAt:      a.CreatedAt.Format(time.RFC3339),
		DownloadCounts: a.DownloadCounts,
		TotalDownloads: a.TotalDownloads,
	}
}

func entryToArchiveToken(e archiveMetaEntry) *ArchiveToken {
	expiresAt, _ := time.Parse(time.RFC3339, e.ExpiresAt)
	createdAt, _ := time.Parse(time.RFC3339, e.CreatedAt)
	return &ArchiveToken{
		Token:          e.Token,
		ServerID:       e.ServerID,
		ServerName:     e.ServerName,
		Include:        e.Include,
		ExpiresAt:      expiresAt,
		CreatedAt:      createdAt,
		DownloadCounts: e.DownloadCounts,
		TotalDownloads: e.TotalDownloads,
	}
}

func archiveMetaPath(volumePath string) string {
	return filepath.Join(volumePath, ".webui-archive-meta.json")
}

func (h *Handler) loadServerArchiveMeta(volumePath string) []archiveMetaEntry {
	data, err := os.ReadFile(archiveMetaPath(volumePath))
	if err != nil {
		return nil
	}
	var entries []archiveMetaEntry
	_ = json.Unmarshal(data, &entries)
	return entries
}

func (h *Handler) saveServerArchiveMeta(volumePath string, entries []archiveMetaEntry) {
	path := archiveMetaPath(volumePath)
	data, _ := json.MarshalIndent(entries, "", "  ")
	_ = os.WriteFile(path, data, 0o644)
}

// InitArchives loads archive metadata from all server volumes on startup.
func (h *Handler) InitArchives() {
	archiveTokensMu.Lock()
	defer archiveTokensMu.Unlock()

	for _, s := range h.Instance.All() {
		if s.VolumePath == "" {
			continue
		}
		entries := h.loadServerArchiveMeta(s.VolumePath)
		for _, e := range entries {
			at := entryToArchiveToken(e)
			archiveTokens[at.Token] = at
		}
	}
}

func (h *Handler) saveArchiveMeta(serverID string, archive *ArchiveToken) {
	s, ok := h.Instance.Get(serverID)
	if !ok || s.VolumePath == "" {
		return
	}
	entries := h.loadServerArchiveMeta(s.VolumePath)
	found := false
	for i, e := range entries {
		if e.Token == archive.Token {
			entries[i] = archive.toMetaEntry()
			found = true
			break
		}
	}
	if !found {
		entries = append(entries, archive.toMetaEntry())
	}
	h.saveServerArchiveMeta(s.VolumePath, entries)
}

func (h *Handler) removeArchiveMeta(serverID, token string) {
	s, ok := h.Instance.Get(serverID)
	if !ok || s.VolumePath == "" {
		return
	}
	entries := h.loadServerArchiveMeta(s.VolumePath)
	filtered := entries[:0]
	for _, e := range entries {
		if e.Token != token {
			filtered = append(filtered, e)
		}
	}
	h.saveServerArchiveMeta(s.VolumePath, filtered)
}

func (h *Handler) cleanupExpiredArchives() {
	archiveTokensMu.Lock()
	now := time.Now()
	type del struct{ token, serverID string }
	var toDelete []del
	for token, entry := range archiveTokens {
		if now.After(entry.ExpiresAt) {
			delete(archiveTokens, token)
			toDelete = append(toDelete, del{token, entry.ServerID})
		}
	}
	archiveTokensMu.Unlock()

	for _, d := range toDelete {
		h.removeArchiveMeta(d.serverID, d.token)
	}
}

// StartArchiveCleanup runs a background ticker that removes expired archives.
func (h *Handler) StartArchiveCleanup(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			h.cleanupExpiredArchives()
		case <-ctx.Done():
			return
		}
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
		TTL     int      `json:"ttl"`     // hours
		Include []string `json:"include"` // "mods", "resourcepacks", "shaderpacks"
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.TTL <= 0 {
		req.TTL = 24
	}
	const maxTTL = 8760 // 12 months in hours
	if req.TTL > maxTTL {
		req.TTL = maxTTL
	}
	if len(req.Include) == 0 {
		req.Include = []string{"mods"}
	}
	formats := []string{"zip", "mrpack", "curseforge", "prism"}

	token := generateToken()
	expiresAt := time.Now().Add(time.Duration(req.TTL) * time.Hour)

	archive := &ArchiveToken{
		Token:          token,
		ServerID:       id,
		ServerName:     s.ServerID,
		Include:        req.Include,
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now(),
		DownloadCounts: make(map[string]int),
	}

	archiveTokensMu.Lock()
	archiveTokens[token] = archive
	archiveTokensMu.Unlock()

	h.saveArchiveMeta(id, archive)

	jsonResponse(w, map[string]any{
		"token":      token,
		"expiresAt":  expiresAt.Format(time.RFC3339),
		"serverName": s.ServerID,
		"formats":    formats,
	})
}

func (h *Handler) collectClientFiles(serverID string, include []string) (map[string][]string, error) {
	s, ok := h.Instance.Get(serverID)
	if !ok || s.VolumePath == "" {
		return nil, fmt.Errorf("server volume not initialized")
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
	for _, t := range include {
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
			return nil, fmt.Errorf("failed to read %s directory: %w", t, err)
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
		return nil, fmt.Errorf("no files found for selected types")
	}
	return filesByType, nil
}

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

func createPrismArchive(w io.Writer, packName, gameVersion, engineType, loaderVersion string, filesByType map[string][]string) error {
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

	filesByType, err := h.collectClientFiles(archive.ServerID, archive.Include)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	srv, _ := h.Instance.Get(archive.ServerID)

	var contentType, ext string
	switch format {
	case "zip":
		contentType = "application/zip"
		ext = "zip"
	case "mrpack":
		contentType = "application/octet-stream"
		ext = "mrpack"
	case "curseforge":
		contentType = "application/zip"
		ext = "zip"
	case "prism":
		contentType = "application/zip"
		ext = "zip"
	default:
		jsonError(w, "invalid format", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s-%s.%s\"", archive.ServerName, archive.CreatedAt.Format("20060102"), ext))
	w.WriteHeader(http.StatusOK)

	var genErr error
	switch format {
	case "zip":
		genErr = createZipArchive(w, filesByType)
	case "mrpack":
		genErr = createMrpackArchive(w, archive.ServerName, filesByType)
	case "curseforge":
		genErr = createCurseForgeArchive(w, archive.ServerName, srv.GameVersion, srv.EngineType, srv.LoaderVersion, filesByType)
	case "prism":
		genErr = createPrismArchive(w, archive.ServerName, srv.GameVersion, srv.EngineType, srv.LoaderVersion, filesByType)
	}

	if genErr != nil {
		return
	}

	archiveTokensMu.Lock()
	archive.TotalDownloads++
	archive.DownloadCounts[format]++
	archiveTokensMu.Unlock()

	h.saveArchiveMeta(archive.ServerID, archive)
}

// HandleGetClientArchiveInfo returns metadata about an archive (public, no auth).
func (h *Handler) HandleGetClientArchiveInfo(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	archiveTokensMu.RLock()
	archive, ok := archiveTokens[token]
	archiveTokensMu.RUnlock()

	if !ok || archive == nil || time.Now().After(archive.ExpiresAt) {
		jsonError(w, "archive not found or expired", http.StatusNotFound)
		return
	}

	jsonResponse(w, map[string]any{
		"token":          token,
		"serverName":     archive.ServerName,
		"expiresAt":      archive.ExpiresAt.Format(time.RFC3339),
		"createdAt":      archive.CreatedAt.Format(time.RFC3339),
		"formats":        []string{"zip", "mrpack", "curseforge", "prism"},
		"downloadCounts": archive.DownloadCounts,
		"totalDownloads": archive.TotalDownloads,
	})
}

// HandleListServerArchives returns all active archive tokens for a server.
func (h *Handler) HandleListServerArchives(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	archiveTokensMu.RLock()
	defer archiveTokensMu.RUnlock()

	now := time.Now()
	var results []map[string]any
	for _, a := range archiveTokens {
		if a.ServerID != id {
			continue
		}
		if now.After(a.ExpiresAt) {
			continue
		}
		results = append(results, map[string]any{
			"token":          a.Token,
			"serverName":     a.ServerName,
			"expiresAt":      a.ExpiresAt.Format(time.RFC3339),
			"createdAt":      a.CreatedAt.Format(time.RFC3339),
			"formats":        []string{"zip", "mrpack", "curseforge", "prism"},
			"downloadCounts": a.DownloadCounts,
			"totalDownloads": a.TotalDownloads,
		})
	}
	jsonResponse(w, results)
}

// HandleDeleteServerArchive removes an archive token.
func (h *Handler) HandleDeleteServerArchive(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	token := r.PathValue("token")

	archiveTokensMu.Lock()
	archive, ok := archiveTokens[token]
	if !ok || archive.ServerID != id {
		archiveTokensMu.Unlock()
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	delete(archiveTokens, token)
	archiveTokensMu.Unlock()

	h.removeArchiveMeta(id, token)

	jsonResponse(w, map[string]any{"deleted": true})
}
