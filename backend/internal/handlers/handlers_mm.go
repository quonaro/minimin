package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"orchestrator/external/mm"
)

// HandleMMSources lists registered content sources and their capabilities.
func (h *Handler) HandleMMSources(w http.ResponseWriter, r *http.Request) {
	result := make(map[string][]string)
	for name, src := range h.ContentSources {
		caps := make([]string, len(src.Capabilities()))
		for i, c := range src.Capabilities() {
			caps[i] = string(c)
		}
		result[name] = caps
	}
	jsonResponse(w, result)
}

// HandleMMSearch proxies search to the requested source.
func (h *Handler) HandleMMSearch(w http.ResponseWriter, r *http.Request) {
	sourceName := r.PathValue("source")
	src, ok := h.ContentSources[sourceName]
	if !ok {
		jsonError(w, "unknown source", http.StatusNotFound)
		return
	}

	contentType := mm.ContentType(r.URL.Query().Get("type"))
	if contentType == "" {
		contentType = mm.ContentMod
	}
	if !mm.IsValidContentType(string(contentType)) {
		jsonError(w, "invalid content type", http.StatusBadRequest)
		return
	}

	// Verify source supports this content type
	supported := false
	for _, c := range src.Capabilities() {
		if c == contentType {
			supported = true
			break
		}
	}
	if !supported {
		jsonError(w, mm.ErrUnsupportedContentType.Error(), http.StatusBadRequest)
		return
	}

	q := r.URL.Query().Get("query")
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")
	gameVersion := r.URL.Query().Get("game_version")
	loader := r.URL.Query().Get("loader")

	offset := 0
	if v, err := strconv.Atoi(offsetStr); err == nil {
		offset = v
	}
	limit := 20
	if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
		limit = v
	}

	items, total, err := src.Search(r.Context(), mm.SearchRequest{
		ContentType: contentType,
		Query:       q,
		GameVersion: gameVersion,
		Loader:      loader,
		Offset:      offset,
		Limit:       limit,
	})
	if err != nil {
		slog.Error("mm search failed", "source", sourceName, "error", err)
		jsonError(w, fmt.Sprintf("search failed: %v", err), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]any{
		"items": items,
		"total": total,
	})
}

// HandleMMGetContent fetches a single content item.
func (h *Handler) HandleMMGetContent(w http.ResponseWriter, r *http.Request) {
	sourceName := r.PathValue("source")
	id := r.PathValue("id")
	src, ok := h.ContentSources[sourceName]
	if !ok {
		jsonError(w, "unknown source", http.StatusNotFound)
		return
	}

	content, err := src.GetContent(r.Context(), id)
	if err != nil {
		slog.Error("mm content fetch failed", "source", sourceName, "id", id, "error", err)
		jsonError(w, fmt.Sprintf("content fetch failed: %v", err), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, content)
}

// HandleMMGetVersions lists versions for a content item.
func (h *Handler) HandleMMGetVersions(w http.ResponseWriter, r *http.Request) {
	sourceName := r.PathValue("source")
	id := r.PathValue("id")
	src, ok := h.ContentSources[sourceName]
	if !ok {
		jsonError(w, "unknown source", http.StatusNotFound)
		return
	}

	loaders := r.URL.Query()["loaders"]
	gameVersions := r.URL.Query()["game_versions"]

	versions, err := src.GetVersions(r.Context(), id, mm.VersionFilter{
		Loaders:      loaders,
		GameVersions: gameVersions,
	})
	if err != nil {
		slog.Error("mm versions fetch failed", "source", sourceName, "id", id, "error", err)
		jsonError(w, fmt.Sprintf("versions fetch failed: %v", err), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, versions)
}

// HandleMMGetVersion fetches a single version.
func (h *Handler) HandleMMGetVersion(w http.ResponseWriter, r *http.Request) {
	sourceName := r.PathValue("source")
	id := r.PathValue("id")
	src, ok := h.ContentSources[sourceName]
	if !ok {
		jsonError(w, "unknown source", http.StatusNotFound)
		return
	}

	version, err := src.GetVersion(r.Context(), id)
	if err != nil {
		slog.Error("mm version fetch failed", "source", sourceName, "version_id", id, "error", err)
		jsonError(w, fmt.Sprintf("version fetch failed: %v", err), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, version)
}

// HandleMMDownload returns a download URL for a version.
func (h *Handler) HandleMMDownload(w http.ResponseWriter, r *http.Request) {
	sourceName := r.PathValue("source")
	id := r.PathValue("id")
	src, ok := h.ContentSources[sourceName]
	if !ok {
		jsonError(w, "unknown source", http.StatusNotFound)
		return
	}

	url, err := src.GetDownloadURL(r.Context(), id)
	if err != nil {
		slog.Error("mm download url fetch failed", "source", sourceName, "version_id", id, "error", err)
		jsonError(w, fmt.Sprintf("download url fetch failed: %v", err), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"url": url})
}
