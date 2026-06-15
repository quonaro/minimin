package handlers

import (
	"fmt"
	"io"
	"net/http"
	"orchestrator/internal/mods"
	"strings"
)

// HandleDownloadClientModFromURL downloads a mod from a remote URL.
func (h *Handler) HandleDownloadClientModFromURL(w http.ResponseWriter, r *http.Request) {
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

	filename, err := h.ClientMods.DownloadClientModFromURL(id, req.URL, req.Filename)
	if err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
		}
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]string{"success": "true", "filename": filename})
}

// HandleListClientMods returns metadata for every .jar in mods-client.
func (h *Handler) HandleListClientMods(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	m, err := h.ClientMods.ListClientMods(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
		}
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]any{"mods": m})
}

// HandleDeleteClientMod removes a single mod .jar from mods-client.
func (h *Handler) HandleDeleteClientMod(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")
	if err := h.ClientMods.DeleteClientMod(id, filename); err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
		}
		jsonError(w, err.Error(), status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleToggleClientMod renames a mod file in mods-client.
func (h *Handler) HandleToggleClientMod(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")
	newFilename, enabled, err := h.ClientMods.ToggleClientMod(id, filename)
	if err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
		}
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]any{"filename": newFilename, "enabled": enabled})
}

// HandleMoveMod moves a mod file between mods/ and mods-client/.
func (h *Handler) HandleMoveMod(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Filename string `json:"filename"`
		Target   string `json:"target"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Filename == "" || req.Target == "" {
		jsonError(w, "filename and target are required", http.StatusBadRequest)
		return
	}

	if err := h.ClientMods.MoveMod(id, req); err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
		}
		if err == mods.ErrVolumeNotInitialized {
			status = http.StatusConflict
		}
		if err == mods.ErrInvalidFilename {
			status = http.StatusBadRequest
		}
		if strings.Contains(err.Error(), "target must be") {
			status = http.StatusBadRequest
		}
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]bool{"success": true})
}

// HandleCopyMod copies a mod file between mods/ and mods-client/.
func (h *Handler) HandleCopyMod(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Filename string `json:"filename"`
		Source   string `json:"source"`
		Target   string `json:"target"`
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

	if err := h.ClientMods.CopyMod(id, req); err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
		}
		if err == mods.ErrVolumeNotInitialized {
			status = http.StatusConflict
		}
		if err == mods.ErrInvalidFilename {
			status = http.StatusBadRequest
		}
		if strings.Contains(err.Error(), "must be 'server' or 'client'") {
			status = http.StatusBadRequest
		}
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]bool{"success": true})
}

// HandleUploadClientMod uploads a .jar or .zip into mods-client.
func (h *Handler) HandleUploadClientMod(w http.ResponseWriter, r *http.Request) {
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

	extracted, err := h.ClientMods.UploadClientMod(id, file, header.Filename, header.Size)
	if err != nil {
		status := http.StatusInternalServerError
		if err == mods.ErrNotFound {
			status = http.StatusNotFound
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

// GetClientModIcon serves the icon image embedded in a client mod .jar.
func (h *Handler) GetClientModIcon(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")
	rc, contentType, err := h.ClientMods.GetClientModIcon(id, filename)
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
