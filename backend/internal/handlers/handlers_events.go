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
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if ok {
		flusher.Flush()
	} else {
		slog.Warn("SSE: http.Flusher not available; events may be buffered")
	}

	clientID, ch := h.EventsHub.Register()
	if ch == nil {
		return
	}
	defer h.EventsHub.Unregister(clientID)

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
				return
			}
			data, err := events.MarshalEvent(ev)
			if err != nil {
				slog.Error("SSE: failed to marshal event", "error", err)
				continue
			}
			if _, writeErr := w.Write(data); writeErr != nil {
				slog.Debug("SSE: client write failed", "error", writeErr)
				return
			}
			if ok {
				flusher.Flush()
			}
		case <-ticker.C:
			if _, err := fmt.Fprintf(w, ":ping\n\n"); err != nil {
				return
			}
			if ok {
				flusher.Flush()
			}
		}
	}
}
