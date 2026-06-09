package routes

import (
	"net/http"

	"orchestrator/internal/events"
	"orchestrator/internal/handlers"
	"orchestrator/internal/jwt"
	"orchestrator/internal/middleware"
	"orchestrator/internal/status"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

func removeErrorsTransformer(ctx huma.Context, status string, v any) (any, error) {
	if len(status) == 0 || (status[0] != '4' && status[0] != '5') {
		return v, nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return v, nil
	}
	delete(m, "errors")
	return m, nil
}

// SetupRoutes wires Huma/Chi router with all orchestrator endpoints.
func SetupRoutes(h *handlers.Handler, apiKey string, jwtService *jwt.Service, store *status.Store, broadcaster *events.Broadcaster) http.Handler {
	chiRouter := chi.NewRouter()

	chiRouter.Use(middleware.RequestLogger)

	// CORS middleware
	chiRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = "http://localhost:3000"
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// REST API uses Huma over chi
	config := huma.DefaultConfig("Minecraft Orchestrator API", "1.0")
	config.CreateHooks = nil
	config.SchemasPath = ""
	config.Transformers = append(config.Transformers, removeErrorsTransformer)

	protectedRouter := chi.NewRouter()
	protectedRouter.Use(middleware.AuthMiddleware(jwtService))
	hapi := humachi.New(protectedRouter, config)

	// Public endpoints
	chiRouter.Post("/api/auth/login", h.LoginHTTP)
	chiRouter.Post("/api/auth/logout", h.LogoutHTTP)

	// SSE endpoint for notifications (public for simplicity; could be protected)
	chiRouter.Get("/api/events", broadcaster.ServeHTTP)

	// Protected endpoints
	protectedRouter.Group(func(r chi.Router) {
		// Agent status endpoint
		r.Get("/agents/status", h.GetAgentStatuses)

		// Agent endpoints
		huma.Register(hapi, huma.Operation{
			OperationID:   "create-agent",
			Method:        http.MethodPost,
			Path:          "/agents",
			Summary:       "Register a new agent",
			DefaultStatus: http.StatusCreated,
		}, h.CreateAgent)

		huma.Register(hapi, huma.Operation{
			OperationID: "check-agent",
			Method:      http.MethodPost,
			Path:        "/agents/check",
			Summary:     "Check agent connectivity",
		}, h.CheckAgent)

		huma.Register(hapi, huma.Operation{
			OperationID: "list-agents",
			Method:      http.MethodGet,
			Path:        "/agents",
			Summary:     "List all agents",
		}, h.ListAgents)

		huma.Register(hapi, huma.Operation{
			OperationID: "get-agent",
			Method:      http.MethodGet,
			Path:        "/agent/{id}",
			Summary:     "Get agent by ID",
		}, h.GetAgent)

		huma.Register(hapi, huma.Operation{
			OperationID:   "delete-agent",
			Method:        http.MethodDelete,
			Path:          "/agents/{id}",
			Summary:       "Delete an agent",
			DefaultStatus: http.StatusNoContent,
		}, h.DeleteAgent)

		// Proxy endpoints to agents
		huma.Register(hapi, huma.Operation{
			OperationID: "proxy-server-create",
			Method:      http.MethodPost,
			Path:        "/agents/{agent_id}/servers",
			Summary:     "Proxy POST /api/v1/servers to agent",
		}, h.ProxyServer)

		huma.Register(hapi, huma.Operation{
			OperationID: "proxy-server-get",
			Method:      http.MethodGet,
			Path:        "/agents/{agent_id}/servers/{server_id}",
			Summary:     "Proxy GET /api/v1/servers/{id} to agent",
		}, h.ProxyServer)

		huma.Register(hapi, huma.Operation{
			OperationID: "proxy-server-action",
			Method:      http.MethodPost,
			Path:        "/agents/{agent_id}/servers/{server_id}/{action}",
			Summary:     "Proxy server actions (start/stop/restart/logs/players/rcon) to agent",
		}, h.ProxyServer)

		// WebSocket proxy for agent server logs.
		r.Get("/ws/agent/{id}/servers/{server_id}/logs", h.WSAgentLogs)

		// WebSocket proxy for agent server RCON console.
		r.Get("/ws/agent/{id}/servers/{server_id}/rcon", h.WSAgentRcon)

		// Generic reverse proxy: any request to /agent/{id}/<anything> is forwarded to the agent's /api/v1/<anything>.
		r.HandleFunc("/agent/{id}/*", h.ProxyAgent)

		// Version proxy endpoints
		huma.Register(hapi, huma.Operation{
			OperationID: "get-all-versions",
			Method:      http.MethodGet,
			Path:        "/versions/all",
			Summary:     "Get all Minecraft versions (vanilla, paper, fabric, forge)",
		}, h.GetAllVersions)

		huma.Register(hapi, huma.Operation{
			OperationID: "get-vanilla-versions",
			Method:      http.MethodGet,
			Path:        "/versions/vanilla",
			Summary:     "Get vanilla Minecraft versions",
		}, h.GetVanillaVersions)

		huma.Register(hapi, huma.Operation{
			OperationID: "get-paper-versions",
			Method:      http.MethodGet,
			Path:        "/versions/paper",
			Summary:     "Get Paper versions",
		}, h.GetPaperVersions)

		huma.Register(hapi, huma.Operation{
			OperationID: "get-fabric-versions",
			Method:      http.MethodGet,
			Path:        "/versions/fabric",
			Summary:     "Get Fabric versions",
		}, h.GetFabricVersions)

		huma.Register(hapi, huma.Operation{
			OperationID: "get-forge-versions",
			Method:      http.MethodGet,
			Path:        "/versions/forge",
			Summary:     "Get Forge versions",
		}, h.GetForgeVersions)

		// Modrinth endpoints
		huma.Register(hapi, huma.Operation{
			OperationID: "mod-search",
			Method:      http.MethodGet,
			Path:        "/mods/search",
			Summary:     "Search mods on Modrinth",
		}, h.ModSearch)

		huma.Register(hapi, huma.Operation{
			OperationID: "mod-versions",
			Method:      http.MethodGet,
			Path:        "/mods/versions/{project_id}",
			Summary:     "Get versions for a Modrinth project",
		}, h.ModVersions)

		huma.Register(hapi, huma.Operation{
			OperationID: "mod-install",
			Method:      http.MethodPost,
			Path:        "/mods/install",
			Summary:     "Install a mod from Modrinth",
		}, h.ModInstall)

		huma.Register(hapi, huma.Operation{
			OperationID: "mod-download",
			Method:      http.MethodPost,
			Path:        "/mods/download",
			Summary:     "Download a mod from a URL to the agent",
		}, h.ModDownload)

		huma.Register(hapi, huma.Operation{
			OperationID: "mod-bulk-install",
			Method:      http.MethodPost,
			Path:        "/mods/bulk",
			Summary:     "Bulk install mods from Modrinth",
		}, h.ModBulkInstall)

		huma.Register(hapi, huma.Operation{
			OperationID: "mod-bulk-job",
			Method:      http.MethodGet,
			Path:        "/mods/jobs/{job_id}",
			Summary:     "Get bulk install job status",
		}, h.ModBulkJob)

		// Orchestrator key for frontend (legacy, kept for compatibility)
		huma.Register(hapi, huma.Operation{
			OperationID: "get-orchestrator-key",
			Method:      http.MethodGet,
			Path:        "/agent/key",
			Summary:     "Get orchestrator API key",
		}, h.GetOrchestratorKey)
	})

	chiRouter.Mount("/", protectedRouter)

	return chiRouter
}
