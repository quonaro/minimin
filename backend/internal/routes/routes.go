package routes

import (
	"net/http"

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
func SetupRoutes(h *handlers.Handler, apiKey string, jwtService *jwt.Service, store *status.Store) http.Handler {
	chiRouter := chi.NewRouter()

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
	hapi := humachi.New(chiRouter, config)

	// Public endpoints
	chiRouter.Post("/api/auth/login", h.LoginHTTP)
	chiRouter.Post("/api/auth/logout", h.LogoutHTTP)

	// WebSocket endpoint for agent statuses (public, no auth required for simplicity)
	chiRouter.Get("/ws/agents", store.ServeHTTP)

	// Protected endpoints
	chiRouter.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtService))

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

		// Generic reverse proxy: any request to /agent/{id}/<anything> is forwarded to the agent's /api/v1/<anything>.
		r.HandleFunc("/agent/{id}/*", h.ProxyAgent)

		// Orchestrator key for frontend (legacy, kept for compatibility)
		huma.Register(hapi, huma.Operation{
			OperationID: "get-orchestrator-key",
			Method:      http.MethodGet,
			Path:        "/agent/key",
			Summary:     "Get orchestrator API key",
		}, h.GetOrchestratorKey)
	})

	return chiRouter
}
