package handlers

import (
	"archive/zip"
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

// HandleDownloadClientModFromURL downloads a mod from a remote URL and saves it into the server's mods-client directory.
func (h *Handler) HandleDownloadClientModFromURL(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("panic in HandleDownloadClientModFromURL", "error", rec)
			jsonError(w, fmt.Sprintf("internal error: %v", rec), http.StatusInternalServerError)
		}
	}()
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	modsDir, err := h.getClientModsDir(id)
	if err != nil {
		jsonError(w, "failed to access client mods directory", http.StatusInternalServerError)
		return
	}
	if modsDir == "" {
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
		slog.Error("failed to download client mod", "url", req.URL, "error", err)
		jsonError(w, fmt.Sprintf("failed to download: %v", err), http.StatusInternalServerError)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		jsonError(w, fmt.Sprintf("upstream returned %d", resp.StatusCode), http.StatusBadGateway)
		return
	}

	if err := runner.FixVolumeOwnership(r.Context(), h.Cli, s.VolumePath, os.Getuid(), os.Getgid()); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", id, "path", s.VolumePath, "error", err)
	}

	if err := os.MkdirAll(modsDir, 0o775); err != nil {
		jsonError(w, "failed to create client mods directory", http.StatusInternalServerError)
		return
	}

	filename := req.Filename
	if filename == "" {
		filename = filepath.Base(req.URL)
	}
	if filename == "" || filename == "." {
		filename = "mod.jar"
	}

	targetPath := filepath.Join(modsDir, filename)
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

func (h *Handler) getClientModsDir(serverID string) (string, error) {
	s, ok := h.Instance.Get(serverID)
	if !ok {
		return "", nil
	}
	if s.VolumePath == "" {
		return "", nil
	}
	return filepath.Join(s.VolumePath, "mods-client"), nil
}

// HandleListClientMods returns metadata for every .jar in the server's mods-client directory.
func (h *Handler) HandleListClientMods(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("panic in HandleListClientMods", "error", rec)
			jsonError(w, fmt.Sprintf("internal error: %v", rec), http.StatusInternalServerError)
		}
	}()
	id := r.PathValue("id")
	modsDir, err := h.getClientModsDir(id)
	if err != nil {
		slog.Error("getClientModsDir failed", "server", id, "error", err)
		jsonError(w, "failed to access client mods directory", http.StatusInternalServerError)
		return
	}
	if modsDir == "" {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}

	entries, err := os.ReadDir(modsDir)
	if err != nil {
		if os.IsNotExist(err) {
			jsonResponse(w, map[string]any{"mods": []ModInfo{}})
			return
		}
		jsonError(w, "failed to read client mods directory", http.StatusInternalServerError)
		return
	}

	var mods []ModInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		lowerName := strings.ToLower(e.Name())
		isDeactivated := strings.HasSuffix(lowerName, ".deactivated")
		if !strings.HasSuffix(lowerName, ".jar") && !isDeactivated {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		modPath := filepath.Join(modsDir, e.Name())
		modInfo, _ := ParseModInfo(modPath, info.Size())
		if modInfo != nil {
			modInfo.Filename = e.Name()
			modInfo.Enabled = !isDeactivated
			if isDeactivated && strings.HasSuffix(modInfo.Name, ".deactivated") {
				modInfo.Name = strings.TrimSuffix(modInfo.Name, ".deactivated")
			}
			mods = append(mods, *modInfo)
		}
	}

	jsonResponse(w, map[string]any{"mods": mods})
}

// HandleDeleteClientMod removes a single mod .jar from the server's mods-client directory.
func (h *Handler) HandleDeleteClientMod(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("panic in HandleDeleteClientMod", "error", rec)
			jsonError(w, fmt.Sprintf("internal error: %v", rec), http.StatusInternalServerError)
		}
	}()
	id := r.PathValue("id")
	modsDir, err := h.getClientModsDir(id)
	if err != nil {
		jsonError(w, "failed to access client mods directory", http.StatusInternalServerError)
		return
	}
	if modsDir == "" {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}

	filename := filepath.Base(r.PathValue("filename"))
	if filename == "" || filename == "." || filename == ".." {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}

	modPath := filepath.Join(modsDir, filename)
	if !strings.HasPrefix(modPath, modsDir+string(filepath.Separator)) && modPath != modsDir {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}

	if err := os.Remove(modPath); err != nil {
		if os.IsNotExist(err) {
			jsonError(w, "mod not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to delete mod", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleToggleClientMod renames a mod file between .jar and .jar.deactivated in mods-client.
func (h *Handler) HandleToggleClientMod(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("panic in HandleToggleClientMod", "error", rec)
			jsonError(w, fmt.Sprintf("internal error: %v", rec), http.StatusInternalServerError)
		}
	}()
	id := r.PathValue("id")
	modsDir, err := h.getClientModsDir(id)
	if err != nil {
		jsonError(w, "failed to access client mods directory", http.StatusInternalServerError)
		return
	}
	if modsDir == "" {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}

	filename := filepath.Base(r.PathValue("filename"))
	if filename == "" || filename == "." || filename == ".." {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}

	modPath := filepath.Join(modsDir, filename)
	if !strings.HasPrefix(modPath, modsDir+string(filepath.Separator)) && modPath != modsDir {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}

	lowerPath := strings.ToLower(modPath)
	var newPath string
	enabled := false
	if strings.HasSuffix(lowerPath, ".deactivated") {
		newPath = strings.TrimSuffix(modPath, ".deactivated")
		enabled = true
	} else {
		newPath = modPath + ".deactivated"
	}

	if err := os.Rename(modPath, newPath); err != nil {
		if os.IsNotExist(err) {
			jsonError(w, "mod not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to toggle mod", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]any{"filename": filepath.Base(newPath), "enabled": enabled})
}

// HandleMoveMod moves a mod file between mods/ and mods-client/ directories.
func (h *Handler) HandleMoveMod(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("panic in HandleMoveMod", "error", rec)
			jsonError(w, fmt.Sprintf("internal error: %v", rec), http.StatusInternalServerError)
		}
	}()
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
		Filename string `json:"filename"`
		Target   string `json:"target"` // "server" or "client"
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Filename == "" || req.Target == "" {
		jsonError(w, "filename and target are required", http.StatusBadRequest)
		return
	}

	serverDir := filepath.Join(s.VolumePath, "mods")
	clientDir := filepath.Join(s.VolumePath, "mods-client")

	if err := runner.FixVolumeOwnership(r.Context(), h.Cli, s.VolumePath, os.Getuid(), os.Getgid()); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", id, "path", s.VolumePath, "error", err)
	}

	if err := os.MkdirAll(clientDir, 0o775); err != nil {
		if info, statErr := os.Stat(clientDir); statErr != nil || !info.IsDir() {
			jsonError(w, "failed to create client mods directory: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	filename := filepath.Base(req.Filename)
	var srcDir, dstDir string
	if req.Target == "client" {
		srcDir = serverDir
		dstDir = clientDir
	} else if req.Target == "server" {
		srcDir = clientDir
		dstDir = serverDir
	} else {
		jsonError(w, "target must be 'server' or 'client'", http.StatusBadRequest)
		return
	}

	srcPath := filepath.Join(srcDir, filename)
	dstPath := filepath.Join(dstDir, filename)

	if !strings.HasPrefix(srcPath, srcDir+string(filepath.Separator)) && srcPath != srcDir {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}

	if err := os.Rename(srcPath, dstPath); err != nil {
		if os.IsNotExist(err) {
			jsonError(w, "mod not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to move mod", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]bool{"success": true})
}

// HandleCopyMod copies a mod file between mods/ and mods-client/ directories.
func (h *Handler) HandleCopyMod(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("panic in HandleCopyMod", "error", rec)
			jsonError(w, fmt.Sprintf("internal error: %v", rec), http.StatusInternalServerError)
		}
	}()
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
		Filename string `json:"filename"`
		Source   string `json:"source"` // "server" or "client"
		Target   string `json:"target"` // "server" or "client"
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Filename == "" || req.Source == "" || req.Target == "" {
		jsonError(w, "filename, source and target are required", http.StatusBadRequest)
		return
	}
	if req.Source == req.Target {
		jsonError(w, "source and target must be different", http.StatusBadRequest)
		return
	}

	serverDir := filepath.Join(s.VolumePath, "mods")
	clientDir := filepath.Join(s.VolumePath, "mods-client")

	if err := runner.FixVolumeOwnership(r.Context(), h.Cli, s.VolumePath, os.Getuid(), os.Getgid()); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", id, "path", s.VolumePath, "error", err)
	}

	if err := os.MkdirAll(clientDir, 0o775); err != nil {
		if info, statErr := os.Stat(clientDir); statErr != nil || !info.IsDir() {
			jsonError(w, "failed to create client mods directory: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if err := os.MkdirAll(serverDir, 0o775); err != nil {
		if info, statErr := os.Stat(serverDir); statErr != nil || !info.IsDir() {
			jsonError(w, "failed to create server mods directory: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	filename := filepath.Base(req.Filename)
	var srcDir, dstDir string
	if req.Source == "server" {
		srcDir = serverDir
	} else if req.Source == "client" {
		srcDir = clientDir
	} else {
		jsonError(w, "source must be 'server' or 'client'", http.StatusBadRequest)
		return
	}
	if req.Target == "server" {
		dstDir = serverDir
	} else if req.Target == "client" {
		dstDir = clientDir
	} else {
		jsonError(w, "target must be 'server' or 'client'", http.StatusBadRequest)
		return
	}

	srcPath := filepath.Join(srcDir, filename)
	dstPath := filepath.Join(dstDir, filename)

	if !strings.HasPrefix(srcPath, srcDir+string(filepath.Separator)) && srcPath != srcDir {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(dstPath, dstDir+string(filepath.Separator)) && dstPath != dstDir {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}

	src, err := os.Open(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			jsonError(w, "mod not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to open mod", http.StatusInternalServerError)
		return
	}
	defer func() { _ = src.Close() }()

	dst, err := os.Create(dstPath)
	if err != nil {
		jsonError(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	defer func() { _ = dst.Close() }()

	if _, err := io.Copy(dst, src); err != nil {
		jsonError(w, "failed to copy mod", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]bool{"success": true})
}

// HandleUploadClientMod uploads a .jar into the server's mods-client directory.
func (h *Handler) HandleUploadClientMod(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("panic in HandleUploadClientMod", "error", rec)
			jsonError(w, fmt.Sprintf("internal error: %v", rec), http.StatusInternalServerError)
		}
	}()
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	modsDir, err := h.getClientModsDir(id)
	if err != nil {
		jsonError(w, "failed to access client mods directory", http.StatusInternalServerError)
		return
	}
	if modsDir == "" {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}

	const maxSize = 128 << 20 // 128MB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := r.ParseMultipartForm(maxSize); err != nil {
		jsonError(w, "invalid multipart form or file too large", http.StatusBadRequest)
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
	if !strings.HasSuffix(strings.ToLower(filename), ".jar") {
		jsonError(w, "only .jar files are allowed", http.StatusBadRequest)
		return
	}

	if err := runner.FixVolumeOwnership(r.Context(), h.Cli, s.VolumePath, os.Getuid(), os.Getgid()); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", id, "path", s.VolumePath, "error", err)
	}

	if err := os.MkdirAll(modsDir, 0o775); err != nil {
		if info, statErr := os.Stat(modsDir); statErr != nil || !info.IsDir() {
			jsonError(w, "failed to create client mods directory: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	targetPath := filepath.Join(modsDir, filename)
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

	jsonResponse(w, map[string]any{
		"success":  true,
		"filename": filename,
	})
}

// GetClientModIcon serves the icon image embedded in a client mod .jar.
func (h *Handler) GetClientModIcon(w http.ResponseWriter, r *http.Request) {
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

	filename := filepath.Base(r.PathValue("filename"))
	if filename == "" || filename == "." || filename == ".." {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}

	modsDir := filepath.Join(s.VolumePath, "mods-client")
	modPath := filepath.Join(modsDir, filename)
	if !strings.HasPrefix(modPath, modsDir+string(filepath.Separator)) && modPath != modsDir {
		jsonError(w, "invalid filename", http.StatusBadRequest)
		return
	}

	info, _ := ParseModInfo(modPath, 0)

	zr, err := zip.OpenReader(modPath)
	if err != nil {
		jsonError(w, "failed to open jar", http.StatusInternalServerError)
		return
	}
	defer func() { _ = zr.Close() }()

	iconPath := ""
	if info != nil {
		iconPath = info.Icon
	}
	if iconPath == "" {
		for _, f := range zr.File {
			name := strings.ToLower(f.Name)
			if name == "icon.png" || name == "icon.jpg" || name == "icon.jpeg" {
				iconPath = f.Name
				break
			}
		}
	}
	if iconPath == "" {
		jsonError(w, "no icon found", http.StatusNotFound)
		return
	}

	for _, f := range zr.File {
		if f.Name == iconPath {
			rc, err := f.Open()
			if err != nil {
				jsonError(w, "failed to read icon", http.StatusInternalServerError)
				return
			}
			defer func() { _ = rc.Close() }()

			contentType := "image/png"
			lower := strings.ToLower(iconPath)
			if strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") {
				contentType = "image/jpeg"
			}
			w.Header().Set("Content-Type", contentType)
			w.WriteHeader(http.StatusOK)
			_, _ = io.Copy(w, rc)
			return
		}
	}

	jsonError(w, "icon not found in jar", http.StatusNotFound)
}
