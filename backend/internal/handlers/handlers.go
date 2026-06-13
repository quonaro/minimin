package handlers

import (
	"encoding/json"
	"net/http"

	"orchestrator/internal/events"
	"orchestrator/internal/modrinth"
	"orchestrator/internal/runner"
	"orchestrator/internal/state"

	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"
)

// Handler holds dependencies for the orchestrator API.
type Handler struct {
	Cli            *client.Client
	Instance       *state.InstanceFile
	APIKey         string
	ServersDir     string
	ServersHostDir string
	ModUploadMaxMB int
	ModrinthClient *modrinth.Client
	EventsHub      *events.Hub
	NetworkName    string
	SecureCookie   bool
	WSUpgrader     websocket.Upgrader
}

// NewHandler creates a new Handler.
func NewHandler(cli *client.Client, instance *state.InstanceFile, apiKey string, modrinthBaseURL string) *Handler {
	return &Handler{
		Cli:            cli,
		Instance:       instance,
		APIKey:         apiKey,
		ModrinthClient: modrinth.NewClient(modrinthBaseURL),
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
