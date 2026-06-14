package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"orchestrator/internal/events"
)

// SSEEvents streams server events as SSE.
// The client receives events: server, players, ops, bans, whitelist.
func (h *Handler) SSEEvents(w http.ResponseWriter, r *http.Request) {
	slog.Info("SSE: client connected", "remote", r.RemoteAddr, "origin", r.Header.Get("Origin"))

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if ok {
		flusher.Flush()
	} else {
		slog.Warn("SSE: http.Flusher not available; events may be buffered")
	}

	clientID, ch := h.EventsHub.Register()
	if ch == nil {
		slog.Warn("SSE: hub closed, rejecting connection")
		return
	}
	slog.Info("SSE: client registered", "client_id", clientID)
	defer func() {
		h.EventsHub.Unregister(clientID)
		slog.Info("SSE: client disconnected", "client_id", clientID)
	}()

	// Send cached last metrics so the client sees data immediately.
	h.EventsHub.RangeMetrics(func(serverID string, payload events.MetricsPayload) bool {
		data, err := events.MarshalEvent(events.Event{Type: "metrics", Payload: payload})
		if err != nil {
			slog.Error("SSE: failed to marshal cached metrics", "error", err)
			return true
		}
		if _, writeErr := w.Write(data); writeErr != nil {
			slog.Error("SSE: client write failed for cached metrics", "client_id", clientID, "error", writeErr)
			return false
		}
		if ok {
			flusher.Flush()
		}
		return true
	})

	// Send an initial ping so the browser knows the connection is alive.
	_, _ = fmt.Fprintf(w, ":ping\n\n")
	if ok {
		flusher.Flush()
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Keep-alive ticker.
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok2 := <-ch:
			if !ok2 {
				slog.Info("SSE: hub channel closed", "client_id", clientID)
				return
			}
			data, err := events.MarshalEvent(ev)
			if err != nil {
				slog.Error("SSE: failed to marshal event", "error", err)
				continue
			}
			if _, writeErr := w.Write(data); writeErr != nil {
				slog.Error("SSE: client write failed", "client_id", clientID, "error", writeErr)
				return
			}
			if ok {
				flusher.Flush()
			}
			slog.Debug("SSE: event sent", "client_id", clientID, "type", ev.Type)
		case <-ticker.C:
			if _, err := fmt.Fprintf(w, ":ping\n\n"); err != nil {
				slog.Error("SSE: ping write failed", "client_id", clientID, "error", err)
				return
			}
			if ok {
				flusher.Flush()
			}
			slog.Debug("SSE: ping sent", "client_id", clientID)
		}
	}
}
