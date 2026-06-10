package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"orchestrator/internal/modrinth"
)

// HandleModrinthSearch proxies search requests to Modrinth.
func (h *Handler) HandleModrinthSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("query")
	facetsRaw := r.URL.Query().Get("facets")
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	offset := 0
	if v, err := strconv.Atoi(offsetStr); err == nil {
		offset = v
	}
	limit := 20
	if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
		limit = v
	}

	client := modrinth.NewClient()
	result, err := client.Search(modrinth.SearchParams{
		Query:  q,
		Facets: facetsRaw,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		jsonError(w, fmt.Sprintf("search failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// HandleModrinthGetProject proxies project info from Modrinth.
func (h *Handler) HandleModrinthGetProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	client := modrinth.NewClient()
	project, err := client.GetProject(id)
	if err != nil {
		jsonError(w, fmt.Sprintf("project fetch failed: %v", err), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, project)
}

// HandleModrinthGetVersions proxies version list from Modrinth.
func (h *Handler) HandleModrinthGetVersions(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	loaders := r.URL.Query()["loaders"]
	gameVersions := r.URL.Query()["game_versions"]

	client := modrinth.NewClient()
	versions, err := client.GetVersions(id, modrinth.VersionParams{
		Loaders:      loaders,
		GameVersions: gameVersions,
	})
	if err != nil {
		jsonError(w, fmt.Sprintf("versions fetch failed: %v", err), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, versions)
}

// HandleModrinthGetVersion proxies a single version from Modrinth.
func (h *Handler) HandleModrinthGetVersion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	client := modrinth.NewClient()
	version, err := client.GetVersion(id)
	if err != nil {
		jsonError(w, fmt.Sprintf("version fetch failed: %v", err), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, version)
}
