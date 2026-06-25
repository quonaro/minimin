package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"orchestrator/internal/mods"
)

// HandleListServerMods returns metadata for every .jar in the server's mods directory.
func (h *Handler) HandleListServerMods(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	m, err := h.Mods.ListServerMods(id)
	if err != nil {
		if err == mods.ErrNotFound {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		if err == mods.ErrVolumeNotInitialized {
			jsonError(w, "server volume not initialized", http.StatusConflict)
			return
		}
		jsonError(w, "failed to read mods directory", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]any{"mods": m})
}

// HandleDeleteServerMod removes a single mod .jar from the server's mods directory.
func (h *Handler) HandleDeleteServerMod(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")
	if err := h.Mods.DeleteServerMod(id, filename); err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
		}
		if err == mods.ErrInvalidFilename {
			status = http.StatusBadRequest
		}
		jsonError(w, err.Error(), status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleToggleServerMod renames a mod file between .jar and .jar.deactivated.
func (h *Handler) HandleToggleServerMod(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")
	newFilename, enabled, err := h.Mods.ToggleServerMod(id, filename)
	if err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
		}
		if err == mods.ErrInvalidFilename {
			status = http.StatusBadRequest
		}
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]any{"filename": newFilename, "enabled": enabled})
}

// GetServerModIcon serves the icon image embedded in a mod .jar.
func (h *Handler) GetServerModIcon(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")
	rc, contentType, err := h.Mods.GetServerModIcon(id, filename)
	if err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
		}
		if err == mods.ErrInvalidFilename {
			status = http.StatusBadRequest
		}
		jsonError(w, err.Error(), status)
		return
	}
	defer func() { _ = rc.Close() }()
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, rc)
}

// GetServerModIconsBatch returns base64-encoded icons for multiple mods.
func (h *Handler) GetServerModIconsBatch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Filenames []string `json:"filenames"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	icons, err := h.Mods.GetServerModIconsBatch(id, req.Filenames)
	if err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
		}
		if err == mods.ErrVolumeNotInitialized {
			status = http.StatusConflict
		}
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]any{"icons": icons})
}

// UploadServerMod uploads a .jar or .zip into the server's mods directory.
func (h *Handler) UploadServerMod(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	maxMB := h.ModUploadMaxMB
	if maxMB <= 0 {
		maxMB = 1024
	}
	maxSize := int64(maxMB) << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := r.ParseMultipartForm(maxSize); err != nil {
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

	extracted, err := h.Mods.UploadServerMod(id, file, header.Filename, header.Size)
	if err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
		}
		if err == mods.ErrVolumeNotInitialized {
			status = http.StatusConflict
		}
		if strings.Contains(err.Error(), "only .jar and .zip") {
			status = http.StatusBadRequest
		}
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]any{
		"success":   true,
		"extracted": extracted,
	})
}

// GetServerIcon serves the server's icon.png from the world volume.
func (h *Handler) GetServerIcon(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	f, err := h.Mods.GetServerIcon(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
		}
		jsonError(w, err.Error(), status)
		return
	}
	defer func() { _ = f.Close() }()
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, f)
}

// UploadServerIcon accepts an image file and writes it as server-icon.png.
func (h *Handler) UploadServerIcon(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
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

	if err := h.Mods.UploadServerIcon(id, file); err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
		}
		if err == mods.ErrVolumeNotInitialized {
			status = http.StatusConflict
		}
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]bool{"success": true})
}

// HandleDownloadModFromURL downloads a mod from a remote URL.
func (h *Handler) HandleDownloadModFromURL(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
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

	filename, err := h.Mods.DownloadModFromURL(id, req.URL, req.Filename)
	if err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
		}
		if err == mods.ErrVolumeNotInitialized {
			status = http.StatusConflict
		}
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]string{"success": "true", "filename": filename})
}

// HandleCopyAllServerMods copies all .jar files from server mods to client mods directory.
func (h *Handler) HandleCopyAllServerMods(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	copied, skipped, errs, err := h.Mods.CopyAllServerMods(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
		}
		if err == mods.ErrVolumeNotInitialized {
			status = http.StatusConflict
		}
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]any{"copied": copied, "skipped": skipped, "errors": errs})
}
