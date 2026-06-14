package handlers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"orchestrator/internal/runner"
	"orchestrator/internal/state"
)

func offlinePlayerUUID(name string) string {
	data := []byte("OfflinePlayer:" + strings.ToLower(name))
	hash := md5.Sum(data)
	hash[6] = (hash[6] & 0x0f) | 0x30
	hash[8] = (hash[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		hash[0:4], hash[4:6], hash[6:8], hash[8:10], hash[10:16])
}

func writeServerJSON(s state.ServerState, filename string, data []map[string]any) error {
	if s.VolumePath == "" {
		return fmt.Errorf("server volume not initialized")
	}
	p := filepath.Join(s.VolumePath, filename)
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}

func sendRcon(s state.ServerState, id string, command string) (string, error) {
	addr := fmt.Sprintf("mc-srv-%s:25575", id)
	client, err := runner.DialRCON(addr, s.RconPassword, 5*time.Second)
	if err != nil {
		return "", err
	}
	defer func() { _ = client.Close() }()
	return client.Execute(command)
}

// HandleOfflineAction performs player list mutations directly on disk
// when offline=true, bypassing Mojang UUID resolution.
func (h *Handler) HandleOfflineAction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}

	var req struct {
		Action  string `json:"action"`
		Name    string `json:"name"`
		Offline bool   `json:"offline"`
		Reason  string `json:"reason"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	// If online mode (or legacy), just proxy through RCON.
	if !req.Offline {
		var cmd string
		switch req.Action {
		case "ban":
			if req.Reason != "" {
				cmd = fmt.Sprintf("ban %s %s", req.Name, req.Reason)
			} else {
				cmd = fmt.Sprintf("ban %s", req.Name)
			}
		case "op":
			cmd = fmt.Sprintf("op %s", req.Name)
		case "whitelist":
			cmd = fmt.Sprintf("whitelist add %s", req.Name)
		case "unban":
			cmd = fmt.Sprintf("pardon %s", req.Name)
		case "deop":
			cmd = fmt.Sprintf("deop %s", req.Name)
		case "whitelist-remove":
			cmd = fmt.Sprintf("whitelist remove %s", req.Name)
		default:
			jsonError(w, "unknown action", http.StatusBadRequest)
			return
		}
		resp, err := sendRcon(s, id, cmd)
		if err != nil {
			jsonError(w, fmt.Sprintf("rcon failed: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, map[string]string{"response": resp})
		h.BroadcastPlayerDataAsync(id)
		return
	}

	uuid := offlinePlayerUUID(req.Name)
	target := strings.ToLower(req.Name)

	switch req.Action {
	case "whitelist":
		list, err := state.ReadServerJSON(s, "whitelist.json")
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		found := false
		for _, e := range list {
			if n, _ := e["name"].(string); strings.ToLower(n) == target {
				found = true
				break
			}
		}
		if !found {
			list = append(list, map[string]any{"name": req.Name, "uuid": uuid})
			if err := writeServerJSON(s, "whitelist.json", list); err != nil {
				jsonError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		go func() { _, _ = sendRcon(s, id, "whitelist reload") }()
		jsonResponse(w, map[string]any{"ok": true, "uuid": uuid})

	case "op":
		list, err := state.ReadServerJSON(s, "ops.json")
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		found := false
		for _, e := range list {
			if n, _ := e["name"].(string); strings.ToLower(n) == target {
				found = true
				break
			}
		}
		if !found {
			list = append(list, map[string]any{
				"name":                req.Name,
				"uuid":                uuid,
				"level":               4,
				"bypassesPlayerLimit": false,
			})
			if err := writeServerJSON(s, "ops.json", list); err != nil {
				jsonError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		jsonResponse(w, map[string]any{"ok": true, "uuid": uuid, "note": "ops.json requires server restart to take effect in vanilla"})

	case "ban":
		list, err := state.ReadServerJSON(s, "banned-players.json")
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		found := false
		for _, e := range list {
			if n, _ := e["name"].(string); strings.ToLower(n) == target {
				found = true
				break
			}
		}
		if !found {
			reason := req.Reason
			if reason == "" {
				reason = "Banned by an operator."
			}
			list = append(list, map[string]any{
				"uuid":    uuid,
				"name":    req.Name,
				"created": time.Now().UTC().Format("2006-01-02 15:04:05 +0000"),
				"source":  "Server",
				"expires": "forever",
				"reason":  reason,
			})
			if err := writeServerJSON(s, "banned-players.json", list); err != nil {
				jsonError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		go func() { _, _ = sendRcon(s, id, "banlist reload") }()
		jsonResponse(w, map[string]any{"ok": true, "uuid": uuid})

	case "unban":
		list, err := state.ReadServerJSON(s, "banned-players.json")
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		filtered := make([]map[string]any, 0, len(list))
		for _, e := range list {
			if n, _ := e["name"].(string); strings.ToLower(n) != target {
				filtered = append(filtered, e)
			}
		}
		if err := writeServerJSON(s, "banned-players.json", filtered); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		go func() { _, _ = sendRcon(s, id, "banlist reload") }()
		jsonResponse(w, map[string]any{"ok": true})

	case "deop":
		list, err := state.ReadServerJSON(s, "ops.json")
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		filtered := make([]map[string]any, 0, len(list))
		for _, e := range list {
			if n, _ := e["name"].(string); strings.ToLower(n) != target {
				filtered = append(filtered, e)
			}
		}
		if err := writeServerJSON(s, "ops.json", filtered); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, map[string]any{"ok": true, "note": "ops.json requires server restart to take effect in vanilla"})

	case "whitelist-remove":
		list, err := state.ReadServerJSON(s, "whitelist.json")
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		filtered := make([]map[string]any, 0, len(list))
		for _, e := range list {
			if n, _ := e["name"].(string); strings.ToLower(n) != target {
				filtered = append(filtered, e)
			}
		}
		if err := writeServerJSON(s, "whitelist.json", filtered); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		go func() { _, _ = sendRcon(s, id, "whitelist reload") }()
		jsonResponse(w, map[string]any{"ok": true})

	default:
		jsonError(w, "unknown action", http.StatusBadRequest)
	}

	h.BroadcastPlayerDataAsync(id)
}
