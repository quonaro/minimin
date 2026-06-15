package handlers

import (
	"context"
	"fmt"
	"net/http"

	"orchestrator/internal/archive"
)

func (h *Handler) archiveService() *archive.Service {
	return archive.NewService(h.Instance)
}

// HandleCreateClientArchive generates .zip and/or .mrpack from client mods.
func (h *Handler) HandleCreateClientArchive(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req archive.CreateRequest
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	a, err := h.archiveService().Create(id, req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusConflict)
		return
	}

	jsonResponse(w, map[string]any{
		"token":      a.Token,
		"expiresAt":  a.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		"serverName": a.ServerName,
		"formats":    []string{"zip", "mrpack", "curseforge", "prism"},
	})
}

// HandleDownloadClientArchive serves a generated archive by token (public, no auth).
func (h *Handler) HandleDownloadClientArchive(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "zip"
	}

	srv := h.archiveService()
	arc, err := srv.Get(token)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	var contentType, ext string
	switch format {
	case "zip", "curseforge", "prism":
		contentType = "application/zip"
		ext = "zip"
	case "mrpack":
		contentType = "application/octet-stream"
		ext = "mrpack"
	default:
		jsonError(w, "invalid format", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s-%s.%s\"", arc.ServerName, arc.CreatedAt.Format("20060102"), ext))
	w.WriteHeader(http.StatusOK)

	baseURL := fmt.Sprintf("%s://%s", r.URL.Scheme, r.Host)
	_ = srv.Download(token, format, baseURL, w)
}

// HandleGetClientArchiveInfo returns metadata about an archive (public, no auth).
func (h *Handler) HandleGetClientArchiveInfo(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	arc, err := h.archiveService().Get(token)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	jsonResponse(w, map[string]any{
		"token":          token,
		"serverName":     arc.ServerName,
		"expiresAt":      arc.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		"createdAt":      arc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"formats":        []string{"zip", "mrpack", "curseforge", "prism"},
		"downloadCounts": arc.DownloadCounts,
		"totalDownloads": arc.TotalDownloads,
	})
}

// HandleGetClientArchiveManifest returns a manifest of all client files with sha256 hashes.
func (h *Handler) HandleGetClientArchiveManifest(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	manifest, err := h.archiveService().Manifest(token)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, manifest)
}

// HandleDownloadClientArchiveFile serves a single file from an archive by token (public, no auth).
func (h *Handler) HandleDownloadClientArchiveFile(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	filePath := r.PathValue("path")
	if err := h.archiveService().DownloadFile(token, filePath, w); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "archive not found or expired" || err.Error() == "file not found" {
			status = http.StatusNotFound
		} else if err.Error() == "invalid file path" {
			status = http.StatusBadRequest
		}
		jsonError(w, err.Error(), status)
		return
	}
}

// HandleListServerArchives returns all active archive tokens for a server.
func (h *Handler) HandleListServerArchives(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	results := h.archiveService().List(id)
	jsonResponse(w, results)
}

// HandleDeleteServerArchive removes an archive token.
func (h *Handler) HandleDeleteServerArchive(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	token := r.PathValue("token")
	if err := h.archiveService().Delete(id, token); err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, map[string]any{"deleted": true})
}

// InitArchives loads archive metadata from all server volumes on startup.
func (h *Handler) InitArchives() {
	archive.InitArchives(h.Instance)
}

// StartArchiveCleanup runs a background ticker that removes expired archives.
func (h *Handler) StartArchiveCleanup(ctx context.Context) {
	archive.StartCleanup(ctx, h.Instance)
}
