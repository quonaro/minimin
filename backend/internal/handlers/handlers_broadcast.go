package handlers

import (
	"fmt"
	"log/slog"
	"time"

	"orchestrator/internal/events"
	"orchestrator/internal/runner"
)

// broadcastPlayerData reads the current player-related JSON files and
// RCON list output, then broadcasts them via SSE.
func (h *Handler) broadcastPlayerData(id string) {
	if h.EventsHub == nil {
		return
	}
	s, ok := h.Instance.Get(id)
	if !ok {
		return
	}

	// Online players via RCON
	if s.ServerStatus == "running" {
		addr := fmt.Sprintf("mc-srv-%s:25575", id)
		client, err := runner.DialRCON(addr, s.RconPassword, 5*time.Second)
		if err == nil {
			resp, rconErr := client.Execute("list")
			_ = client.Close()
			if rconErr == nil {
				online, maxPlayers, players := parseListResponse(resp)
				_ = online
				h.EventsHub.BroadcastJSON("players", events.PlayerListPayload{
					ServerID: id,
					Players:  players,
					Max:      maxPlayers,
				})
			}
		} else {
			slog.Debug("broadcastPlayerData: rcon not available", "server_id", id, "error", err)
		}
	}

	// Ops
	ops, _ := readServerJSON(s, "ops.json")
	h.EventsHub.BroadcastJSON("ops", events.OpsPayload{ServerID: id, Ops: ops})

	// Bans
	bans, _ := readServerJSON(s, "banned-players.json")
	h.EventsHub.BroadcastJSON("bans", events.BansPayload{ServerID: id, Bans: bans})

	// Whitelist
	wl, _ := readServerJSON(s, "whitelist.json")
	h.EventsHub.BroadcastJSON("whitelist", events.WhitelistPayload{ServerID: id, Whitelist: wl})
}

// BroadcastPlayerDataAsync wraps broadcastPlayerData in a goroutine.
func (h *Handler) BroadcastPlayerDataAsync(id string) {
	go h.broadcastPlayerData(id)
}
