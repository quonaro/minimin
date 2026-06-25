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

	// Health check (no auth required)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Versions (no auth required — read-only public data)
	mux.HandleFunc("GET /api/versions", h.HandleGetVersions)

	// SSE events endpoint (protected)
	mux.HandleFunc("GET /api/events", middleware.WithAuth(apiKey, h.SSEEvents))

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
	mux.HandleFunc("POST /api/servers/{id}/mods/icons", middleware.WithAuth(apiKey, h.GetServerModIconsBatch))

	// Server CRUD (protected)
	mux.HandleFunc("POST /api/servers", middleware.WithAuth(apiKey, h.HandleCreateServer))
	mux.HandleFunc("POST /api/servers/prepare-instance", middleware.WithAuth(apiKey, h.HandlePrepareInstance))
	mux.HandleFunc("POST /api/servers/from-instance", middleware.WithAuth(apiKey, h.HandleCreateServerFromInstance))
	mux.HandleFunc("GET /api/servers", middleware.WithAuth(apiKey, h.HandleListServers))
	mux.HandleFunc("GET /api/servers/{id}", middleware.WithAuth(apiKey, h.HandleGetServer))
	mux.HandleFunc("PATCH /api/servers/{id}", middleware.WithAuth(apiKey, h.HandleUpdateServer))
	mux.HandleFunc("POST /api/servers/{id}/start", middleware.WithAuth(apiKey, h.HandleStartServer))
	mux.HandleFunc("POST /api/servers/{id}/stop", middleware.WithAuth(apiKey, h.HandleStopServer))
	mux.HandleFunc("POST /api/servers/{id}/restart", middleware.WithAuth(apiKey, h.HandleRestartServer))
	mux.HandleFunc("POST /api/servers/{id}/force-stop", middleware.WithAuth(apiKey, h.HandleForceStopServer))
	mux.HandleFunc("DELETE /api/servers/{id}", middleware.WithAuth(apiKey, h.HandleDeleteServer))
	mux.HandleFunc("POST /api/servers/{id}/recreate-world", middleware.WithAuth(apiKey, h.HandleRecreateWorld))
	mux.HandleFunc("POST /api/servers/{id}/reassign-ports", middleware.WithAuth(apiKey, h.HandleReassignPorts))
	mux.HandleFunc("GET /api/servers/{id}/pull-progress", middleware.WithAuth(apiKey, h.HandleGetPullProgress))

	// Config (protected)
	mux.HandleFunc("GET /api/servers/{id}/config", middleware.WithAuth(apiKey, h.HandleGetServerConfig))
	mux.HandleFunc("PATCH /api/servers/{id}/config", middleware.WithAuth(apiKey, h.HandleUpdateServerConfig))

	// Disk usage (protected)
	mux.HandleFunc("GET /api/servers/{id}/disk", middleware.WithAuth(apiKey, h.HandleGetServerDisk))

	// RCON / Logs / Players / Bans / Ops / Whitelist (protected)
	mux.HandleFunc("POST /api/servers/{id}/rcon", middleware.WithAuth(apiKey, h.HandleSendRconCommand))
	mux.HandleFunc("POST /api/servers/{id}/offline-action", middleware.WithAuth(apiKey, h.HandleOfflineAction))
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

	// Client mods (protected)
	mux.HandleFunc("GET /api/servers/{id}/client-mods", middleware.WithAuth(apiKey, h.HandleListClientMods))
	mux.HandleFunc("POST /api/servers/{id}/client-mods/upload", middleware.WithAuth(apiKey, h.HandleUploadClientMod))
	mux.HandleFunc("POST /api/servers/{id}/client-mods/download", middleware.WithAuth(apiKey, h.HandleDownloadClientModFromURL))
	mux.HandleFunc("DELETE /api/servers/{id}/client-mods/{filename}", middleware.WithAuth(apiKey, h.HandleDeleteClientMod))
	mux.HandleFunc("POST /api/servers/{id}/client-mods/{filename}/toggle", middleware.WithAuth(apiKey, h.HandleToggleClientMod))
	mux.HandleFunc("POST /api/servers/{id}/client-mods/move", middleware.WithAuth(apiKey, h.HandleMoveMod))
	mux.HandleFunc("POST /api/servers/{id}/client-mods/archive", middleware.WithAuth(apiKey, h.HandleCreateClientArchive))
	mux.HandleFunc("GET /api/servers/{id}/client-mods/archives", middleware.WithAuth(apiKey, h.HandleListServerArchives))
	mux.HandleFunc("DELETE /api/servers/{id}/client-mods/archives/{token}", middleware.WithAuth(apiKey, h.HandleDeleteServerArchive))
	mux.HandleFunc("POST /api/servers/{id}/mods/copy", middleware.WithAuth(apiKey, h.HandleCopyMod))
	mux.HandleFunc("POST /api/servers/{id}/mods/copy-all", middleware.WithAuth(apiKey, h.HandleCopyAllServerMods))
	mux.HandleFunc("GET /api/servers/{id}/client-mods/{filename}/icon", middleware.WithAuth(apiKey, h.GetClientModIcon))
	mux.HandleFunc("POST /api/servers/{id}/client-mods/icons", middleware.WithAuth(apiKey, h.GetClientModIconsBatch))

	// Crash Reports (protected)
	mux.HandleFunc("GET /api/servers/{id}/crash-reports", middleware.WithAuth(apiKey, h.HandleListCrashReports))
	mux.HandleFunc("GET /api/servers/{id}/crash-reports/{filename}", middleware.WithAuth(apiKey, h.HandleReadCrashReport))
	mux.HandleFunc("DELETE /api/servers/{id}/crash-reports/{filename}", middleware.WithAuth(apiKey, h.HandleDeleteCrashReport))

	// Client assets (resourcepacks / shaderpacks) (protected)
	mux.HandleFunc("GET /api/servers/{id}/client-assets", middleware.WithAuth(apiKey, h.HandleListClientAssets))
	mux.HandleFunc("POST /api/servers/{id}/client-assets/upload", middleware.WithAuth(apiKey, h.HandleUploadClientAsset))
	mux.HandleFunc("POST /api/servers/{id}/client-assets/download", middleware.WithAuth(apiKey, h.HandleDownloadClientAssetFromURL))
	mux.HandleFunc("DELETE /api/servers/{id}/client-assets/{filename}", middleware.WithAuth(apiKey, h.HandleDeleteClientAsset))
	mux.HandleFunc("POST /api/servers/{id}/client-assets/{filename}/toggle", middleware.WithAuth(apiKey, h.HandleToggleClientAsset))

	// Client archive (public)
	mux.HandleFunc("GET /api/client-archive/{token}", h.HandleDownloadClientArchive)
	mux.HandleFunc("GET /api/client-archive/{token}/info", h.HandleGetClientArchiveInfo)
	mux.HandleFunc("GET /api/client-archive/{token}/manifest", h.HandleGetClientArchiveManifest)
	mux.HandleFunc("GET /api/client-archive/{token}/file/{path...}", h.HandleDownloadClientArchiveFile)

	// Content source proxy (protected)
	mux.HandleFunc("GET /api/mm/sources", middleware.WithAuth(apiKey, h.HandleMMSources))
	mux.HandleFunc("GET /api/mm/{source}/search", middleware.WithAuth(apiKey, h.HandleMMSearch))
	mux.HandleFunc("GET /api/mm/{source}/content/{id}", middleware.WithAuth(apiKey, h.HandleMMGetContent))
	mux.HandleFunc("GET /api/mm/{source}/content/{id}/versions", middleware.WithAuth(apiKey, h.HandleMMGetVersions))
	mux.HandleFunc("GET /api/mm/{source}/version/{id}", middleware.WithAuth(apiKey, h.HandleMMGetVersion))
	mux.HandleFunc("GET /api/mm/{source}/version/{id}/download", middleware.WithAuth(apiKey, h.HandleMMDownload))

	return middleware.RequestLogger(mux)
}
