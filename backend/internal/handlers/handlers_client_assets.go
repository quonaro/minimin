package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"orchestrator/internal/runner"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (h *Handler) getClientAssetsDir(serverID string, assetType string) (string, error) {
	s, ok := h.Instance.Get(serverID)
	if !ok {
		return "", nil
	}
	if s.VolumePath == "" {
		return "", nil
	}
	return filepath.Join(s.VolumePath, assetType), nil
}

// HandleListClientAssets returns files from resourcepacks or shaderpacks directory.
func (h *Handler) HandleListClientAssets(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("panic in HandleListClientAssets", "error", rec)
			jsonError(w, fmt.Sprintf("internal error: %v", rec), http.StatusInternalServerError)
		}
	}()
	id := r.PathValue("id")
	assetType := r.URL.Query().Get("type")
	if assetType == "" {
		assetType = "resourcepacks"
	}
	switch assetType {
	case "resourcepacks", "shaderpacks":
	default:
		jsonError(w, "invalid asset type", http.StatusBadRequest)
		return
	}
	dir, err := h.getClientAssetsDir(id, assetType)
	if err != nil {
		jsonError(w, "failed to access directory", http.StatusInternalServerError)
		return
	}
	if dir == "" {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			jsonResponse(w, map[string]any{"assets": []any{}})
			return
		}
		jsonError(w, "failed to read directory", http.StatusInternalServerError)
		return
	}
	var assets []map[string]any
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		isDeactivated := strings.HasSuffix(name, ".disabled")
		if !strings.HasSuffix(name, ".zip") && !isDeactivated {
			continue
		}
		info, _ := e.Info()
		assets = append(assets, map[string]any{
			"filename": e.Name(),
			"size":     info.Size(),
			"enabled":  !isDeactivated,
		})
	}
	jsonResponse(w, map[string]any{"assets": assets})
}

// HandleUploadClientAsset uploads a .zip into resourcepacks or shaderpacks directory.
func (h *Handler) HandleUploadClientAsset(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("panic in HandleUploadClientAsset", "error", rec)
			jsonError(w, fmt.Sprintf("internal error: %v", rec), http.StatusInternalServerError)
		}
	}()
	id := r.PathValue("id")
	assetType := r.URL.Query().Get("type")
	if assetType == "" {
		assetType = "resourcepacks"
	}
	switch assetType {
	case "resourcepacks", "shaderpacks":
	default:
		jsonError(w, "invalid asset type", http.StatusBadRequest)
		return
	}
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	dir, err := h.getClientAssetsDir(id, assetType)
	if err != nil {
		jsonError(w, "failed to access directory", http.StatusInternalServerError)
		return
	}
	if dir == "" {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	maxMB := h.ModUploadMaxMB
	if maxMB <= 0 {
		maxMB = 1024
	}
	maxSize := int64(maxMB) << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := r.ParseMultipartForm(maxSize); err != nil {
		slog.Warn("multipart parse failed", "error", err)
		jsonError(w, fmt.Sprintf("invalid multipart form or file too large: %s", err.Error()), http.StatusBadRequest)
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()
	file, header, err := r.FormFile("file")
	if err != nil {
		jsonError(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()
	filename := filepath.Base(header.Filename)
	if !strings.HasSuffix(strings.ToLower(filename), ".zip") {
		jsonError(w, "only .zip files are allowed", http.StatusBadRequest)
		return
	}
	uid, gid := runner.ContainerUIDGID()
	if err := runner.FixVolumeOwnership(r.Context(), h.Cli, s.VolumePath, uid, gid); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", id, "path", s.VolumePath, "error", err)
	}
	if err := os.MkdirAll(dir, 0o775); err != nil {
		if info, statErr := os.Stat(dir); statErr != nil || !info.IsDir() {
			jsonError(w, "failed to create directory: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	targetPath := filepath.Join(dir, filename)
	out, err := os.Create(targetPath)
	if err != nil {
		jsonError(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	if _, err := out.ReadFrom(file); err != nil {
		_ = out.Close()
		jsonError(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	_ = out.Close()
	jsonResponse(w, map[string]any{"success": true, "filename": filename})
}

// HandleDownloadClientAssetFromURL downloads a .zip from a remote URL and saves it into the server's resourcepacks or shaderpacks directory.
func (h *Handler) HandleDownloadClientAssetFromURL(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("panic in HandleDownloadClientAssetFromURL", "error", rec)
			jsonError(w, fmt.Sprintf("internal error: %v", rec), http.StatusInternalServerError)
		}
	}()
	id := r.PathValue("id")
	assetType := r.URL.Query().Get("type")
	if assetType == "" {
		assetType = "resourcepacks"
	}
	switch assetType {
	case "resourcepacks", "shaderpacks":
	default:
		jsonError(w, "invalid asset type", http.StatusBadRequest)
		return
	}
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	dir, err := h.getClientAssetsDir(id, assetType)
	if err != nil {
		jsonError(w, "failed to access directory", http.StatusInternalServerError)
		return
	}
	if dir == "" {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}

	var req struct {
		URL      string `json:"url"`
		Filename string `json:"filename"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		jsonError(w, "url is required", http.StatusBadRequest)
		return
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(req.URL)
	if err != nil {
		slog.Error("failed to download client asset", "url", req.URL, "error", err)
		jsonError(w, fmt.Sprintf("failed to download: %v", err), http.StatusInternalServerError)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		jsonError(w, fmt.Sprintf("upstream returned %d", resp.StatusCode), http.StatusBadGateway)
		return
	}

	filename := req.Filename
	if filename == "" {
		filename = filepath.Base(req.URL)
	}
	if filename == "" || filename == "." {
		filename = "asset.zip"
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".zip") {
		jsonError(w, "only .zip files are allowed", http.StatusBadRequest)
		return
	}

	uid, gid := runner.ContainerUIDGID()
	if err := runner.FixVolumeOwnership(r.Context(), h.Cli, s.VolumePath, uid, gid); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", id, "path", s.VolumePath, "error", err)
	}
	if err := os.MkdirAll(dir, 0o775); err != nil {
		if info, statErr := os.Stat(dir); statErr != nil || !info.IsDir() {
			jsonError(w, "failed to create directory: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	targetPath := filepath.Join(dir, filename)
	out, err := os.Create(targetPath)
	if err != nil {
		jsonError(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, resp.Body); err != nil {
		jsonError(w, "failed to save file", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{"success": "true", "filename": filename})
}

// HandleDeleteClientAsset removes a file from resourcepacks or shaderpacks directory.
func (h *Handler) HandleDeleteClientAsset(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("panic in HandleDeleteClientAsset", "error", rec)
			jsonError(w, fmt.Sprintf("internal error: %v", rec), http.StatusInternalServerError)
		}
	}()
	id := r.PathValue("id")
	assetType := r.URL.Query().Get("type")
	if assetType == "" {
		assetType = "resourcepacks"
	}
	switch assetType {
	case "resourcepacks", "shaderpacks":
	default:
		jsonError(w, "invalid asset type", http.StatusBadRequest)
		return
	}
	dir, err := h.getClientAssetsDir(id, assetType)
	if err != nil {
		jsonError(w, "failed to access directory", http.StatusInternalServerError)
		return
	}
	if dir == "" {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	filename := filepath.Base(r.PathValue("filename"))
	if filename == "" || filename == "." || filename == ".." {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}
	assetPath := filepath.Join(dir, filename)
	if !strings.HasPrefix(assetPath, dir+string(filepath.Separator)) && assetPath != dir {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}
	if err := os.Remove(assetPath); err != nil {
		if os.IsNotExist(err) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to delete", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleToggleClientAsset renames a file between .zip and .zip.disabled.
func (h *Handler) HandleToggleClientAsset(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("panic in HandleToggleClientAsset", "error", rec)
			jsonError(w, fmt.Sprintf("internal error: %v", rec), http.StatusInternalServerError)
		}
	}()
	id := r.PathValue("id")
	assetType := r.URL.Query().Get("type")
	if assetType == "" {
		assetType = "resourcepacks"
	}
	switch assetType {
	case "resourcepacks", "shaderpacks":
	default:
		jsonError(w, "invalid asset type", http.StatusBadRequest)
		return
	}
	dir, err := h.getClientAssetsDir(id, assetType)
	if err != nil {
		jsonError(w, "failed to access directory", http.StatusInternalServerError)
		return
	}
	if dir == "" {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	filename := filepath.Base(r.PathValue("filename"))
	if filename == "" || filename == "." || filename == ".." {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}
	assetPath := filepath.Join(dir, filename)
	if !strings.HasPrefix(assetPath, dir+string(filepath.Separator)) && assetPath != dir {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}
	lowerPath := strings.ToLower(assetPath)
	var newPath string
	enabled := false
	if strings.HasSuffix(lowerPath, ".disabled") {
		newPath = strings.TrimSuffix(assetPath, ".disabled")
		enabled = true
	} else {
		newPath = assetPath + ".disabled"
	}
	if err := os.Rename(assetPath, newPath); err != nil {
		if os.IsNotExist(err) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to toggle", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]any{"filename": filepath.Base(newPath), "enabled": enabled})
}
