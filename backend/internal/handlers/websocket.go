package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"orchestrator/internal/runner"

	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"
)

// WSLogs upgrades the connection and streams container stdout/stderr as text messages.
func (h *Handler) WSLogs(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("id")

	s, ok := h.Instance.Get(serverID)
	if !ok {
		slog.Warn("WSLogs: server not found", "id", serverID)
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if s.ContainerID == "" {
		slog.Warn("WSLogs: container not created", "id", serverID)
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

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WSLogs: upgrade failed", "error", err)
		return
	}
	defer func() {
		slog.Info("WSLogs: closing websocket", "server", serverID)
		_ = conn.Close()
	}()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go func() {
		defer cancel()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	ww := newWSBatchWriter(conn, 50*time.Millisecond, 100)
	defer func() { _ = ww.Close() }()

	slog.Info("WSLogs: sending tail logs", "server", serverID, "container", s.ContainerID, "tail", tailLines)
	if err := runner.StreamContainerLogs(ctx, h.Cli, s.ContainerID, ww, ww, tailLines, false); err != nil {
		if isClientDisconnect(err) {
			slog.Info("WSLogs: client disconnected during tail logs", "server", serverID)
		} else if client.IsErrNotFound(err) {
			slog.Warn("WSLogs: container not found", "server", serverID, "container", s.ContainerID)
			_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"container not found"}`))
		} else {
			slog.Error("WSLogs: tail logs failed", "server", serverID, "error", err)
		}
		return
	}

	retries := 0
	for {
		select {
		case <-ctx.Done():
			slog.Info("WSLogs: context cancelled, exiting", "server", serverID)
			return
		default:
		}

		inspectCtx, cancelInspect := context.WithTimeout(context.Background(), 5*time.Second)
		info, err := h.Cli.ContainerInspect(inspectCtx, s.ContainerID)
		cancelInspect()
		if err != nil {
			slog.Error("WSLogs: inspect container failed", "server", serverID, "container", s.ContainerID, "error", err)
			return
		}
		if !info.State.Running {
			slog.Debug("WSLogs: container not running, waiting", "server", serverID, "status", info.State.Status, "retries", retries)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				retries++
				continue
			}
		}

		slog.Info("WSLogs: starting live stream", "server", serverID, "container", s.ContainerID)
		retries = 0
		if err := runner.StreamContainerLogs(ctx, h.Cli, s.ContainerID, ww, ww, 0, true); err != nil {
			if isClientDisconnect(err) {
				slog.Info("WSLogs: client disconnected during live stream", "server", serverID)
			} else if client.IsErrNotFound(err) {
				slog.Warn("WSLogs: container not found during live stream", "server", serverID, "container", s.ContainerID)
				_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"container not found"}`))
			} else {
				slog.Error("WSLogs: live stream ended with error", "server", serverID, "error", err)
			}
			return
		}

		slog.Debug("WSLogs: live stream ended cleanly, retrying in 2s", "server", serverID)
		select {
		case <-ctx.Done():
			slog.Info("WSLogs: context cancelled during retry wait", "server", serverID)
			return
		case <-time.After(2 * time.Second):
		}
	}
}

// WSRcon upgrades the connection and proxies JSON commands to the Minecraft RCON port.
func (h *Handler) WSRcon(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("id")
	s, ok := h.Instance.Get(serverID)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if s.ContainerID == "" {
		jsonError(w, "container not created", http.StatusBadRequest)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	addr := fmt.Sprintf("127.0.0.1:%d", s.RconPort)
	rconClient, err := runner.DialRCON(addr, s.RconPassword, 5*time.Second)
	if err != nil {
		_ = conn.WriteJSON(map[string]string{"error": err.Error()})
		return
	}
	defer func() { _ = rconClient.Close() }()

	for {
		var req struct {
			Command string `json:"command"`
		}
		if err := conn.ReadJSON(&req); err != nil {
			return
		}
		resp, err := rconClient.Execute(req.Command)
		if err != nil {
			if writeErr := conn.WriteJSON(map[string]string{"error": err.Error()}); writeErr != nil {
				return
			}
			continue
		}
		if writeErr := conn.WriteJSON(map[string]string{"response": resp}); writeErr != nil {
			return
		}
	}
}

// isClientDisconnect reports whether an error is caused by the client
// closing the WebSocket (broken pipe, connection reset, etc.).
func isClientDisconnect(err error) bool {
	if err == nil {
		return false
	}
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
		return true
	}
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		if errors.Is(netErr.Err, syscall.EPIPE) || errors.Is(netErr.Err, syscall.ECONNRESET) {
			return true
		}
	}
	if errors.Is(err, net.ErrClosed) {
		return true
	}
	return false
}

// wsBatchWriter batches log lines and flushes them as WebSocket text messages.
type wsBatchWriter struct {
	conn          *websocket.Conn
	mu            sync.Mutex
	buf           bytes.Buffer
	timer         *time.Timer
	flushInterval time.Duration
	maxLines      int
}

func newWSBatchWriter(conn *websocket.Conn, flushInterval time.Duration, maxLines int) *wsBatchWriter {
	return &wsBatchWriter{
		conn:          conn,
		flushInterval: flushInterval,
		maxLines:      maxLines,
	}
}

func (w *wsBatchWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, writeErr := w.buf.Write(p); writeErr != nil {
		return 0, writeErr
	}

	lines := strings.Count(w.buf.String(), "\n")
	if lines >= w.maxLines {
		return len(p), w.flushLocked()
	}

	if w.timer == nil {
		w.timer = time.AfterFunc(w.flushInterval, w.flush)
	}
	return len(p), nil
}

func (w *wsBatchWriter) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	_ = w.flushLocked()
}

func (w *wsBatchWriter) flushLocked() error {
	if w.timer != nil {
		w.timer.Stop()
		w.timer = nil
	}
	if w.buf.Len() == 0 {
		return nil
	}
	data := w.buf.Bytes()
	w.buf.Reset()
	return w.conn.WriteMessage(websocket.TextMessage, data)
}

func (w *wsBatchWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.flushLocked()
}
