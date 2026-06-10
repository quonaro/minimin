package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"orchestrator/internal/runner"
	"orchestrator/internal/state"

	"github.com/docker/docker/api/types/container"
	"github.com/moby/moby/api/pkg/stdcopy"
)

// handleGetServerConfig returns the raw server.properties file contents.
func (h *Handler) HandleGetServerConfig(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if s.VolumePath == "" {
		jsonResponse(w, map[string]any{"content": "", "initialized": false, "pendingProperties": s.PendingProperties})
		return
	}
	propsPath := filepath.Join(s.VolumePath, "server.properties")
	data, err := os.ReadFile(propsPath)
	if err != nil {
		if os.IsNotExist(err) {
			jsonResponse(w, map[string]any{"content": "", "initialized": false, "pendingProperties": s.PendingProperties})
			return
		}
		jsonError(w, "failed to read config", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]any{"content": string(data), "initialized": true, "pendingProperties": s.PendingProperties})
}

// handleUpdateServerConfig patches keys in server.properties.
func (h *Handler) HandleUpdateServerConfig(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if s.VolumePath == "" {
		jsonError(w, "server volume not initialized", http.StatusConflict)
		return
	}
	var req struct {
		Properties map[string]any `json:"properties"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(s.VolumePath, 0o755); err != nil {
		jsonError(w, "failed to create volume directory", http.StatusInternalServerError)
		return
	}
	propsPath := filepath.Join(s.VolumePath, "server.properties")
	if err := UpdateProperties(propsPath, req.Properties); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pendingProps := make(map[string]string, len(req.Properties))
	for k, v := range req.Properties {
		pendingProps[k] = fmt.Sprintf("%v", v)
	}
	h.Instance.AddPendingProperties(id, pendingProps)
	_ = h.Instance.Save()
	w.WriteHeader(http.StatusNoContent)
}

// UpdateProperties patches a properties file with new key=value pairs.
func UpdateProperties(path string, updates map[string]any) error {
	var lines []string
	data, err := os.ReadFile(path)
	if err == nil {
		lines = strings.Split(string(data), "\n")
	}
	updated := make(map[string]bool)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if val, ok := updates[key]; ok {
			lines[i] = fmt.Sprintf("%s=%v", key, val)
			updated[key] = true
		}
	}
	for k, v := range updates {
		if !updated[k] {
			lines = append(lines, fmt.Sprintf("%s=%v", k, v))
		}
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}

// handleSendRconCommand sends an RCON command to a running server.
func (h *Handler) HandleSendRconCommand(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	var req struct {
		Command string `json:"command"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	addr := fmt.Sprintf("127.0.0.1:%d", s.RconPort)
	rconClient, err := runner.DialRCON(addr, s.RconPassword, 5*time.Second)
	if err != nil {
		jsonError(w, fmt.Sprintf("rcon connection failed: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	defer func() { _ = rconClient.Close() }()
	resp, err := rconClient.Execute(req.Command)
	if err != nil {
		jsonError(w, fmt.Sprintf("rcon execution failed: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"response": resp})
}

// handleGetServerLogs returns the latest container log lines.
func (h *Handler) HandleGetServerLogs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if s.ContainerID == "" {
		jsonError(w, "container not created", http.StatusBadRequest)
		return
	}
	tailLines := 100
	if t := r.URL.Query().Get("tail"); t != "" {
		if v, err := strconv.Atoi(t); err == nil && v > 0 {
			tailLines = v
		}
	}
	if tailLines > 50000 {
		tailLines = 50000
	}
	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
		Tail:       fmt.Sprintf("%d", tailLines),
	}
	out, err := h.Cli.ContainerLogs(r.Context(), s.ContainerID, opts)
	if err != nil {
		jsonError(w, fmt.Sprintf("failed to get logs: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	defer func() { _ = out.Close() }()
	var buf strings.Builder
	_, err = stdcopy.StdCopy(&buf, &buf, out)
	if err != nil {
		jsonError(w, fmt.Sprintf("failed to read logs: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = []string{}
	}
	jsonResponse(w, map[string]any{"lines": lines})
}

var listResponseRe = regexp.MustCompile(`There are (\d+) of a max(?: of)? (\d+) players online[:.]?(.*)`)

func parseListResponse(resp string) (int, int, []string) {
	matches := listResponseRe.FindStringSubmatch(resp)
	if len(matches) < 4 {
		return 0, 0, []string{}
	}
	online, _ := strconv.Atoi(matches[1])
	maxPlayers, _ := strconv.Atoi(matches[2])
	names := strings.Split(matches[3], ",")
	players := make([]string, 0, len(names))
	for _, n := range names {
		if trimmed := strings.TrimSpace(n); trimmed != "" {
			players = append(players, trimmed)
		}
	}
	return online, maxPlayers, players
}

// handleGetPlayers returns the current online players via the RCON list command.
func (h *Handler) HandleGetPlayers(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if s.ServerStatus != "running" {
		jsonError(w, "server is not running", http.StatusServiceUnavailable)
		return
	}
	addr := fmt.Sprintf("127.0.0.1:%d", s.RconPort)
	rconClient, err := runner.DialRCON(addr, s.RconPassword, 5*time.Second)
	if err != nil {
		jsonError(w, fmt.Sprintf("rcon connection failed: %s", err.Error()), http.StatusServiceUnavailable)
		return
	}
	defer func() { _ = rconClient.Close() }()
	resp, err := rconClient.Execute("list")
	if err != nil {
		jsonError(w, fmt.Sprintf("rcon execution failed: %s", err.Error()), http.StatusServiceUnavailable)
		return
	}
	online, maxPlayers, players := parseListResponse(resp)
	jsonResponse(w, map[string]any{"online": online, "max": maxPlayers, "players": players})
}

func readServerJSON(s state.ServerState, filename string) ([]map[string]any, error) {
	if s.VolumePath == "" {
		return []map[string]any{}, nil
	}
	p := filepath.Join(s.VolumePath, filename)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]any{}, nil
		}
		return nil, err
	}
	var out []map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// handleGetBans returns the server's banned-players.json.
func (h *Handler) HandleGetBans(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	bans, err := readServerJSON(s, "banned-players.json")
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]any{"bans": bans})
}

// handleGetOps returns the server's ops.json.
func (h *Handler) HandleGetOps(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	ops, err := readServerJSON(s, "ops.json")
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]any{"ops": ops})
}

// handleGetWhitelist returns the server's whitelist.json.
func (h *Handler) HandleGetWhitelist(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	wl, err := readServerJSON(s, "whitelist.json")
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]any{"whitelist": wl})
}
