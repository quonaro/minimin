package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"orchestrator/internal/db"
	"orchestrator/internal/jwt"
	"orchestrator/internal/status"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// Handler holds dependencies for the orchestrator API.
type Handler struct {
	DB     *db.DB
	APIKey string
	JWT    *jwt.Service
	Status *status.Store
}

// NewHandler creates a new Handler.
func NewHandler(database *db.DB, apiKey string, jwtService *jwt.Service, store *status.Store) *Handler {
	return &Handler{DB: database, APIKey: apiKey, JWT: jwtService, Status: store}
}

// CreateAgentInput is the input for POST /agents.
type CreateAgentInput struct {
	Body struct {
		Name   string `json:"name" doc:"Agent name"`
		Host   string `json:"host" doc:"Agent host (e.g., http://localhost:8080)"`
		APIKey string `json:"api_key" doc:"Agent API key for authentication"`
	}
}

// CreateAgentOutput is the output for POST /agents.
type CreateAgentOutput struct {
	Body db.Agent
}

// CreateAgent registers a new agent.
func (h *Handler) CreateAgent(ctx context.Context, input *CreateAgentInput) (*CreateAgentOutput, error) {
	id := uuid.New().String()
	agent := db.Agent{
		ID:        id,
		Name:      input.Body.Name,
		Host:      input.Body.Host,
		APIKey:    input.Body.APIKey,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.DB.CreateAgent(agent); err != nil {
		return nil, huma.Error500InternalServerError("failed to create agent", err)
	}

	return &CreateAgentOutput{Body: agent}, nil
}

// ListAgentsInput is the input for GET /agents.
type ListAgentsInput struct{}

// ListAgentsOutput is the output for GET /agents.
type ListAgentsOutput struct {
	Body []db.Agent
}

// ListAgents returns all registered agents.
func (h *Handler) ListAgents(ctx context.Context, input *ListAgentsInput) (*ListAgentsOutput, error) {
	agents, err := h.DB.ListAgents()
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list agents", err)
	}
	for i := range agents {
		agents[i].APIKey = ""
	}
	return &ListAgentsOutput{Body: agents}, nil
}

// GetAgentInput is the input for GET /agents/{id}.
type GetAgentInput struct {
	ID string `path:"id" doc:"Agent ID"`
}

// GetAgentOutput is the output for GET /agents/{id}.
type GetAgentOutput struct {
	Body db.Agent
}

// GetAgent returns a single agent by ID.
func (h *Handler) GetAgent(ctx context.Context, input *GetAgentInput) (*GetAgentOutput, error) {
	slog.Info("GetAgent called", "id", input.ID)
	agent, ok := h.DB.GetAgent(input.ID)
	if !ok {
		slog.Warn("agent not found in DB", "id", input.ID)
		return nil, huma.Error404NotFound("agent not found", nil)
	}
	agent.APIKey = ""
	return &GetAgentOutput{Body: agent}, nil
}

// DeleteAgentInput is the input for DELETE /agents/{id}.
type DeleteAgentInput struct {
	ID string `path:"id" doc:"Agent ID"`
}

// DeleteAgentOutput is the output for DELETE /agents/{id}.
type DeleteAgentOutput struct{}

// DeleteAgent removes an agent by ID.
func (h *Handler) DeleteAgent(ctx context.Context, input *DeleteAgentInput) (*DeleteAgentOutput, error) {
	if err := h.DB.DeleteAgent(input.ID); err != nil {
		if err.Error() == "agent not found" {
			return nil, huma.Error404NotFound("agent not found", nil)
		}
		return nil, huma.Error500InternalServerError("failed to delete agent", err)
	}
	return &DeleteAgentOutput{}, nil
}

// CheckAgentInput is the input for POST /agents/check.
type CheckAgentInput struct {
	Body struct {
		Host   string `json:"host" doc:"Agent host (e.g., http://localhost:8080)"`
		APIKey string `json:"api_key" doc:"Agent API key for authentication"`
	}
}

// CheckAgentOutputBody is the body of CheckAgentOutput.
type CheckAgentOutputBody struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

// CheckAgentOutput is the output for POST /agents/check.
type CheckAgentOutput struct {
	Body CheckAgentOutputBody
}

// CheckAgent verifies connectivity to an agent by calling its /api/v1/agent/key endpoint.
func (h *Handler) CheckAgent(ctx context.Context, input *CheckAgentInput) (*CheckAgentOutput, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := strings.TrimSuffix(input.Body.Host, "/") + "/api/v1/agent/key"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create request", err)
	}
	req.Header.Set("Authorization", "Bearer "+input.Body.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return &CheckAgentOutput{Body: CheckAgentOutputBody{Valid: false, Error: "agent unreachable"}}, nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return &CheckAgentOutput{Body: CheckAgentOutputBody{Valid: false, Error: "invalid response from agent"}}, nil
	}

	return &CheckAgentOutput{Body: CheckAgentOutputBody{Valid: true}}, nil
}

// GetOrchestratorKeyInput is the input for GET /agent/key.
type GetOrchestratorKeyInput struct {
	Authorization string `header:"Authorization" doc:"Bearer token for authentication"`
}

// GetOrchestratorKeyOutput is the output for GET /agent/key.
type GetOrchestratorKeyOutput struct {
	Body GetOrchestratorKeyOutputBody
}

// GetOrchestratorKeyOutputBody is the body of GetOrchestratorKeyOutput.
type GetOrchestratorKeyOutputBody struct {
	Valid bool `json:"valid"`
}

// GetOrchestratorKey validates the orchestrator API key.
func (h *Handler) GetOrchestratorKey(ctx context.Context, input *GetOrchestratorKeyInput) (*GetOrchestratorKeyOutput, error) {
	expected := "Bearer " + h.APIKey
	if input.Authorization != expected {
		return nil, huma.Error401Unauthorized("invalid API key", nil)
	}
	return &GetOrchestratorKeyOutput{Body: GetOrchestratorKeyOutputBody{Valid: true}}, nil
}

// ProxyServerInput is the input for proxying to agent servers.
type ProxyServerInput struct {
	AgentID  string `path:"agent_id" doc:"Agent ID"`
	ServerID string `path:"server_id" doc:"Server ID"`
	Action   string `path:"action" doc:"Action (start, stop, restart, logs, players, rcon)"`
	Body     struct {
		Command string `json:"command,omitempty" doc:"RCON command (for rcon action)"`
	}
}

// ProxyServerOutput is the output for proxying to agent servers.
type ProxyServerOutput struct {
	Body interface{}
}

// ProxyServer proxies requests to agent servers.
func (h *Handler) ProxyServer(ctx context.Context, input *ProxyServerInput) (*ProxyServerOutput, error) {
	agent, ok := h.DB.GetAgent(input.AgentID)
	if !ok {
		return nil, huma.Error404NotFound("agent not found", nil)
	}

	// TODO: Implement HTTP proxy to agent
	// For now, return placeholder
	return &ProxyServerOutput{Body: map[string]string{
		"message": fmt.Sprintf("Proxy to %s for server %s action %s not yet implemented", agent.Host, input.ServerID, input.Action),
	}}, nil
}

// ProxyAgent is a raw HTTP reverse proxy to an agent. Any path under /agent/{id}/* is forwarded to the agent's /api/v1/*.
func (h *Handler) ProxyAgent(w http.ResponseWriter, r *http.Request) {
	parts := strings.SplitN(r.URL.Path, "/", 4)
	if len(parts) < 4 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid proxy path"})
		return
	}

	agentID := parts[2]
	suffix := parts[3]

	agent, ok := h.DB.GetAgent(agentID)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "agent not found"})
		return
	}

	targetURL := strings.TrimSuffix(agent.Host, "/") + "/api/v1/" + suffix
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to create proxy request"})
		return
	}

	for k, vv := range r.Header {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	req.Header.Set("Authorization", "Bearer "+agent.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "agent unreachable"})
		return
	}
	defer func() { _ = resp.Body.Close() }()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

// LoginInput is the input for POST /auth/login.
type LoginInput struct {
	Body struct {
		APIKey string `json:"api_key" doc:"Orchestrator API key"`
	}
}

// LoginOutputBody is the body of LoginOutput.
type LoginOutputBody struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
}

// LoginOutput is the output for POST /auth/login.
type LoginOutput struct {
	Body LoginOutputBody
}

// Login validates the API key and returns JWT token.
func (h *Handler) Login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	if input.Body.APIKey != h.APIKey {
		return nil, huma.Error401Unauthorized("invalid API key", nil)
	}

	token, err := h.JWT.GenerateToken()
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to generate token", err)
	}

	return &LoginOutput{Body: LoginOutputBody{Success: true, Token: token}}, nil
}

// LoginHTTP is a regular HTTP handler for login that sets httpOnly cookie.
func (h *Handler) LoginHTTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		APIKey string `json:"api_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.APIKey != h.APIKey {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token, err := h.JWT.GenerateToken()
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set httpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// LogoutHTTP clears the httpOnly cookie.
func (h *Handler) LogoutHTTP(w http.ResponseWriter, r *http.Request) {
	// Clear cookie by setting MaxAge to -1
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// CheckAllAgents pings every registered agent and updates the in-memory status store.
func (h *Handler) CheckAllAgents() {
	agents, err := h.DB.ListAgents()
	if err != nil {
		slog.Error("failed to list agents for health check", "error", err)
		return
	}

	var wg sync.WaitGroup
	for _, agent := range agents {
		wg.Add(1)
		go func(a db.Agent) {
			defer wg.Done()
			online := h.pingAgent(a.Host, a.APIKey)
			h.Status.Set(a.ID, online)
		}(agent)
	}
	wg.Wait()
	h.Status.Broadcast()
}

func (h *Handler) pingAgent(host, apiKey string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	url := strings.TrimSuffix(host, "/") + "/api/v1/agent/key"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}
