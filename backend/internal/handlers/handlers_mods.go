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

// handleListServerMods returns metadata for every .jar in the server's mods directory.
func (h *Handler) HandleListServerMods(w http.ResponseWriter, r *http.Request) {
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

	modsDir := filepath.Join(s.VolumePath, "mods")
	entries, err := os.ReadDir(modsDir)
	if err != nil {
		if os.IsNotExist(err) {
			jsonResponse(w, map[string]any{"mods": []ModInfo{}})
			return
		}
		jsonError(w, "failed to read mods directory", http.StatusInternalServerError)
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

// handleDeleteServerMod removes a single mod .jar from the server's mods directory.
func (h *Handler) HandleDeleteServerMod(w http.ResponseWriter, r *http.Request) {
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

	modsDir := filepath.Join(s.VolumePath, "mods")
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

// handleToggleServerMod renames a mod file between .jar and .jar.deactivated to enable or disable it.
func (h *Handler) HandleToggleServerMod(w http.ResponseWriter, r *http.Request) {
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

	modsDir := filepath.Join(s.VolumePath, "mods")
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
		enabled = false
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

// GetServerModIcon serves the icon image embedded in a mod .jar.
func (h *Handler) GetServerModIcon(w http.ResponseWriter, r *http.Request) {
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

	modsDir := filepath.Join(s.VolumePath, "mods")
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

// UploadServerMod uploads a .jar or .zip into the server's mods directory.
func (h *Handler) UploadServerMod(w http.ResponseWriter, r *http.Request) {
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

	modsDir := filepath.Join(s.VolumePath, "mods")
	uid, gid := runner.ContainerUIDGID()
	if err := runner.FixVolumeOwnership(r.Context(), h.Cli, s.VolumePath, uid, gid); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", id, "path", s.VolumePath, "error", err)
	}
	if err := os.MkdirAll(modsDir, 0o775); err != nil {
		jsonError(w, "failed to create mods directory", http.StatusInternalServerError)
		return
	}

	filename := filepath.Base(header.Filename)
	isZip := strings.HasSuffix(strings.ToLower(filename), ".zip")
	isJar := strings.HasSuffix(strings.ToLower(filename), ".jar")

	if !isZip && !isJar {
		jsonError(w, "only .jar and .zip files are allowed", http.StatusBadRequest)
		return
	}

	if isJar {
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
			"success":   true,
			"extracted": []string{filename},
		})
		return
	}

	// .zip: save to temp, extract jars, delete temp
	tmpPath := filepath.Join(modsDir, ".upload_"+filename)
	tmp, err := os.Create(tmpPath)
	if err != nil {
		jsonError(w, "failed to create temp file", http.StatusInternalServerError)
		return
	}
	if _, err := tmp.ReadFrom(file); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		jsonError(w, "failed to save upload", http.StatusInternalServerError)
		return
	}
	_ = tmp.Close()

	extracted, extractErr := extractZipJars(tmpPath, modsDir)
	_ = os.Remove(tmpPath)
	if extractErr != nil {
		jsonError(w, extractErr.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]any{
		"success":   true,
		"extracted": extracted,
	})
}

// GetServerIcon serves the server's icon.png (server-icon.png) from the world volume.
func (h *Handler) GetServerIcon(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}

	iconPath := filepath.Join(s.VolumePath, "server-icon.png")
	f, err := os.Open(iconPath)
	if err != nil {
		if os.IsNotExist(err) {
			jsonError(w, "icon not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to open icon", http.StatusInternalServerError)
		return
	}
	defer func() { _ = f.Close() }()

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, f)
}

// UploadServerIcon accepts an image file and writes it as server-icon.png into the server's volume.
func (h *Handler) UploadServerIcon(w http.ResponseWriter, r *http.Request) {
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

	const maxSize = 1 << 20 // 1MB
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

	file, _, err := r.FormFile("icon")
	if err != nil {
		jsonError(w, "missing icon field", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	iconPath := filepath.Join(s.VolumePath, "server-icon.png")
	if err := os.MkdirAll(s.VolumePath, 0o755); err != nil {
		slog.Error("failed to create volume directory", "path", s.VolumePath, "error", err)
		jsonError(w, fmt.Sprintf("failed to create volume directory: %v", err), http.StatusInternalServerError)
		return
	}
	f, err := os.Create(iconPath)
	if err != nil {
		slog.Error("failed to create icon file", "path", iconPath, "error", err)
		jsonError(w, fmt.Sprintf("failed to create icon file: %v", err), http.StatusInternalServerError)
		return
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(f, file); err != nil {
		slog.Error("failed to write icon file", "path", iconPath, "error", err)
		jsonError(w, fmt.Sprintf("failed to write icon: %v", err), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]bool{"success": true})
}

// HandleDownloadModFromURL downloads a mod from a remote URL and saves it into the server's mods directory.
func (h *Handler) HandleDownloadModFromURL(w http.ResponseWriter, r *http.Request) {
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
		slog.Error("failed to download mod", "url", req.URL, "error", err)
		jsonError(w, fmt.Sprintf("failed to download: %v", err), http.StatusInternalServerError)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		jsonError(w, fmt.Sprintf("upstream returned %d", resp.StatusCode), http.StatusBadGateway)
		return
	}

	modsDir := filepath.Join(s.VolumePath, "mods")
	uid, gid := runner.ContainerUIDGID()
	if err := runner.FixVolumeOwnership(r.Context(), h.Cli, s.VolumePath, uid, gid); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", id, "path", s.VolumePath, "error", err)
	}
	if err := os.MkdirAll(modsDir, 0o775); err != nil {
		jsonError(w, "failed to create mods directory", http.StatusInternalServerError)
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

// HandleCopyAllServerMods copies all .jar files from the server mods directory to the client mods directory.
func (h *Handler) HandleCopyAllServerMods(w http.ResponseWriter, r *http.Request) {
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

	serverDir := filepath.Join(s.VolumePath, "mods")
	clientDir := filepath.Join(s.VolumePath, "mods-client")

	uid, gid := runner.ContainerUIDGID()
	if err := runner.FixVolumeOwnership(r.Context(), h.Cli, s.VolumePath, uid, gid); err != nil {
		slog.Warn("failed to fix volume ownership", "server_id", id, "path", s.VolumePath, "error", err)
	}
	if err := os.MkdirAll(clientDir, 0o775); err != nil {
		jsonError(w, "failed to create client mods directory", http.StatusInternalServerError)
		return
	}

	entries, err := os.ReadDir(serverDir)
	if err != nil {
		if os.IsNotExist(err) {
			jsonResponse(w, map[string]any{"copied": []string{}, "skipped": []string{}, "errors": []string{}})
			return
		}
		jsonError(w, "failed to read server mods directory", http.StatusInternalServerError)
		return
	}

	var copied []string
	var skipped []string
	var errs []string

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

	jsonResponse(w, map[string]any{"copied": copied, "skipped": skipped, "errors": errs})
}
