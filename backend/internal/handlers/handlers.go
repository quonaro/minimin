package handlers

import (
	"encoding/json"
	"net/http"

	"orchestrator/external/mm"
	"orchestrator/internal/actions"
	"orchestrator/internal/clientmods"
	"orchestrator/internal/events"
	"orchestrator/internal/mods"
	"orchestrator/internal/runner"
	"orchestrator/internal/state"

	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"
)

// Handler holds dependencies for the orchestrator API.
type Handler struct {
	Cli            *client.Client
	Instance       *state.InstanceFile
	Actions        *actions.Service
	Mods           *mods.Service
	ClientMods     *clientmods.Service
	APIKey         string
	ServersDir     string
	ServersHostDir string
	ModUploadMaxMB int
	ContentSources map[string]mm.ContentSource
	EventsHub      *events.Hub
	NetworkName    string
	SecureCookie   bool
	WSUpgrader     websocket.Upgrader
}

// NewHandler creates a new Handler.
func NewHandler(cli *client.Client, instance *state.InstanceFile, apiKey string) *Handler {
	return &Handler{
		Cli:      cli,
		Instance: instance,
		APIKey:   apiKey,
	}
}

func (h *Handler) isSecure(r *http.Request) bool {
	if h.SecureCookie {
		return true
	}
	if r.TLS != nil {
		return true
	}
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}
	return false
}

// Login validates the static API key and sets an httpOnly cookie.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		APIKey string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.APIKey != h.APIKey {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    h.APIKey,
		Path:     "/",
		MaxAge:   86400 * 365, // 1 year
		HttpOnly: true,
		Secure:   h.isSecure(r),
		SameSite: http.SameSiteStrictMode,
	})
	jsonResponse(w, map[string]bool{"success": true})
}

// Logout clears the httpOnly cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.isSecure(r),
		SameSite: http.SameSiteStrictMode,
	})
	jsonResponse(w, map[string]bool{"success": true})
}

// resolveLegacyFields fills missing fields for legacy servers.
func (h *Handler) resolveLegacyFields(s state.ServerState) state.ServerState {
	if s.HostPath == "" && s.VolumePath != "" {
		s.HostPath = runner.HostPathForDocker(s.VolumePath, h.ServersDir, h.ServersHostDir)
	}
	if s.ContainerPath == "" {
		s.ContainerPath = "/data"
	}
	s.ModCount = state.CountMods(s)
	return s
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}
