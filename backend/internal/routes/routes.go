package routes

import (
	"net/http"

	"orchestrator/internal/handlers"
	"orchestrator/internal/middleware"
)

// SetupRoutes wires net/http.ServeMux with all orchestrator endpoints.
func SetupRoutes(h *handlers.Handler, apiKey string) http.Handler {
	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("POST /api/auth/logout", h.Logout)

	// Versions (no auth required — read-only public data)
	mux.HandleFunc("GET /api/versions", h.HandleGetVersions)

	// WebSocket endpoints (protected)
	mux.HandleFunc("GET /ws/servers/{id}/logs", middleware.WithAuth(apiKey, h.WSLogs))
	mux.HandleFunc("GET /ws/servers/{id}/rcon", middleware.WithAuth(apiKey, h.WSRcon))

	// File upload/download endpoints (protected)
	mux.HandleFunc("GET /api/servers/{id}/icon", middleware.WithAuth(apiKey, h.GetServerIcon))
	mux.HandleFunc("POST /api/servers/{id}/icon", middleware.WithAuth(apiKey, h.UploadServerIcon))
	mux.HandleFunc("GET /api/servers/{id}/files/download", middleware.WithAuth(apiKey, h.DownloadServerFile))
	mux.HandleFunc("POST /api/servers/{id}/files/upload", middleware.WithAuth(apiKey, h.UploadServerFile))
	mux.HandleFunc("POST /api/servers/{id}/mods/upload", middleware.WithAuth(apiKey, h.UploadServerMod))
	mux.HandleFunc("GET /api/servers/{id}/mods/{filename}/icon", middleware.WithAuth(apiKey, h.GetServerModIcon))

	// Server CRUD (protected)
	mux.HandleFunc("POST /api/servers", middleware.WithAuth(apiKey, h.HandleCreateServer))
	mux.HandleFunc("GET /api/servers", middleware.WithAuth(apiKey, h.HandleListServers))
	mux.HandleFunc("GET /api/servers/{id}", middleware.WithAuth(apiKey, h.HandleGetServer))
	mux.HandleFunc("PATCH /api/servers/{id}", middleware.WithAuth(apiKey, h.HandleUpdateServer))
	mux.HandleFunc("POST /api/servers/{id}/start", middleware.WithAuth(apiKey, h.HandleStartServer))
	mux.HandleFunc("POST /api/servers/{id}/stop", middleware.WithAuth(apiKey, h.HandleStopServer))
	mux.HandleFunc("POST /api/servers/{id}/restart", middleware.WithAuth(apiKey, h.HandleRestartServer))
	mux.HandleFunc("DELETE /api/servers/{id}", middleware.WithAuth(apiKey, h.HandleDeleteServer))
	mux.HandleFunc("POST /api/servers/{id}/recreate-world", middleware.WithAuth(apiKey, h.HandleRecreateWorld))

	// Config (protected)
	mux.HandleFunc("GET /api/servers/{id}/config", middleware.WithAuth(apiKey, h.HandleGetServerConfig))
	mux.HandleFunc("PATCH /api/servers/{id}/config", middleware.WithAuth(apiKey, h.HandleUpdateServerConfig))

	// RCON / Logs / Players / Bans / Ops / Whitelist (protected)
	mux.HandleFunc("POST /api/servers/{id}/rcon", middleware.WithAuth(apiKey, h.HandleSendRconCommand))
	mux.HandleFunc("GET /api/servers/{id}/logs", middleware.WithAuth(apiKey, h.HandleGetServerLogs))
	mux.HandleFunc("GET /api/servers/{id}/players", middleware.WithAuth(apiKey, h.HandleGetPlayers))
	mux.HandleFunc("GET /api/servers/{id}/bans", middleware.WithAuth(apiKey, h.HandleGetBans))
	mux.HandleFunc("GET /api/servers/{id}/ops", middleware.WithAuth(apiKey, h.HandleGetOps))
	mux.HandleFunc("GET /api/servers/{id}/whitelist", middleware.WithAuth(apiKey, h.HandleGetWhitelist))

	// Files (protected)
	mux.HandleFunc("GET /api/servers/{id}/files", middleware.WithAuth(apiKey, h.HandleListServerFiles))
	mux.HandleFunc("DELETE /api/servers/{id}/files", middleware.WithAuth(apiKey, h.HandleDeleteServerFile))
	mux.HandleFunc("GET /api/servers/{id}/file", middleware.WithAuth(apiKey, h.HandleReadServerFile))
	mux.HandleFunc("PUT /api/servers/{id}/file", middleware.WithAuth(apiKey, h.HandleWriteServerFile))
	mux.HandleFunc("POST /api/servers/{id}/files/create", middleware.WithAuth(apiKey, h.HandleCreateServerFile))
	mux.HandleFunc("POST /api/servers/{id}/files/mkdir", middleware.WithAuth(apiKey, h.HandleMkdirServerFile))
	mux.HandleFunc("POST /api/servers/{id}/files/move", middleware.WithAuth(apiKey, h.HandleMoveServerFile))

	// Mods (protected)
	mux.HandleFunc("GET /api/servers/{id}/mods", middleware.WithAuth(apiKey, h.HandleListServerMods))
	mux.HandleFunc("POST /api/servers/{id}/mods/download", middleware.WithAuth(apiKey, h.HandleDownloadModFromURL))
	mux.HandleFunc("DELETE /api/servers/{id}/mods/{filename}", middleware.WithAuth(apiKey, h.HandleDeleteServerMod))
	mux.HandleFunc("POST /api/servers/{id}/mods/{filename}/toggle", middleware.WithAuth(apiKey, h.HandleToggleServerMod))

	return middleware.RequestLogger(mux)
}
