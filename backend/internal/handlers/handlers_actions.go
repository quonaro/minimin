package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"orchestrator/internal/runner"
	"orchestrator/internal/state"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func (h *Handler) doStart(id string) {
	s, _ := h.Instance.Get(id)
	prevStatus := s.Status

	defer func() {
		if r := recover(); r != nil {
			slog.Error("start server panic", "server_id", id, "recover", r)
			h.Instance.ClearDesired(id, prevStatus)
			_ = h.Instance.Save()
		}
	}()

	if s.ContainerID != "" {
		if err := h.Cli.ContainerStart(context.Background(), s.ContainerID, container.StartOptions{}); err != nil {
			if client.IsErrNotFound(err) {
				slog.Warn("stale container id, recreating", "server_id", id, "container_id", s.ContainerID)
				s.ContainerID = ""
			} else {
				slog.Error("failed to start container", "server_id", id, "error", err)
				h.Instance.ClearDesired(id, prevStatus)
				_ = h.Instance.Save()
				return
			}
		}
	}

	if s.ContainerID == "" {
		if s.RconPort == 0 {
			s.RconPort = s.GamePort + 10
		}
		portUsed := func(p uint16) bool {
			return h.Instance.IsPortUsed(p, id)
		}
		gamePort, err := runner.FindFreePortExcluding("", s.GamePort, portUsed)
		if err != nil {
			slog.Error("no free game port", "server_id", id, "error", err)
			h.Instance.ClearDesired(id, prevStatus)
			_ = h.Instance.Save()
			return
		}
		rconHost := "127.0.0.1"
		if s.PublicRcon {
			rconHost = ""
		}
		rconPort, err := runner.FindFreePortExcluding(rconHost, s.RconPort, portUsed)
		if err != nil {
			slog.Error("no free rcon port", "server_id", id, "error", err)
			h.Instance.ClearDesired(id, prevStatus)
			_ = h.Instance.Save()
			return
		}
		if rconPort == gamePort {
			rconPort, err = runner.FindFreePortExcluding(rconHost, 0, portUsed)
			if err != nil {
				slog.Error("no free rcon port", "server_id", id, "error", err)
				h.Instance.ClearDesired(id, prevStatus)
				_ = h.Instance.Save()
				return
			}
		}
		s.GamePort = gamePort
		s.RconPort = rconPort
		if s.RconPassword == "" {
			pw, _ := runner.GenerateRconPassword()
			s.RconPassword = pw
		}
		if s.VolumePath != "" {
			uid, gid := runner.ContainerUIDGID()
			if err := runner.FixVolumeOwnership(context.Background(), h.Cli, s.VolumePath, uid, gid); err != nil {
				slog.Warn("failed to fix volume ownership", "server_id", id, "path", s.VolumePath, "error", err)
			}
		}
		containerID, volumeID, volumePath, err := runner.StartServerContainer(
			context.Background(), h.Cli, s.ServerID,
			s.RamBytes, s.GamePort,
			s.EngineType, s.GameVersion, s.LoaderVersion,
			h.ServersDir, h.ServersHostDir,
			s.RconPort, s.RconPassword, s.PublicRcon,
			s.VolumePath,
			nil,
			s.RestartPolicy,
			h.NetworkName,
			s.ExternalJavaArgs,
		)
		if err != nil {
			slog.Error("failed to start server container", "server_id", id, "error", err)
			h.Instance.ClearDesired(id, prevStatus)
			_ = h.Instance.Save()
			return
		}
		s.ContainerID = containerID
		s.VolumeID = volumeID
		s.VolumePath = volumePath
		s.HostPath = runner.HostPathForDocker(volumePath, h.ServersDir, h.ServersHostDir)
		s.ContainerPath = "/data"
		s.ModCount = state.CountMods(s)
	}

	s.ContainerStatus = "running"
	s.ContainerStartedAt = time.Now().UTC()
	s.ServerStatus = "starting"
	s.PendingProperties = nil
	h.Instance.Set(s)
	h.Instance.ClearDesired(id, "running")
	_ = h.Instance.Save()
	slog.Info("server started", "server_id", id)
}

// handleStartServer starts an existing server container asynchronously.
func (h *Handler) HandleStartServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	s, ok := h.Instance.TrySetDesired(id, "running")
	if !ok {
		jsonError(w, "operation in progress", http.StatusConflict)
		return
	}
	_ = h.Instance.Save()
	go h.doStart(id)
	jsonResponse(w, s)
}

func (h *Handler) doStop(id string) {
	s, _ := h.Instance.Get(id)
	prevStatus := s.Status

	defer func() {
		if r := recover(); r != nil {
			slog.Error("stop server panic", "server_id", id, "recover", r)
			h.Instance.ClearDesired(id, prevStatus)
			_ = h.Instance.Save()
		}
	}()

	if s.ContainerID == "" {
		slog.Warn("stop server: container not created", "server_id", id)
		h.Instance.ClearDesired(id, prevStatus)
		_ = h.Instance.Save()
		return
	}
	timeout := 30
	if err := h.Cli.ContainerStop(context.Background(), s.ContainerID, container.StopOptions{Timeout: &timeout}); err != nil {
		if client.IsErrNotFound(err) {
			slog.Warn("container already removed", "server_id", id, "container_id", s.ContainerID)
		} else {
			slog.Error("failed to stop container", "server_id", id, "error", err)
			h.Instance.ClearDesired(id, prevStatus)
			_ = h.Instance.Save()
			return
		}
	}

	s.ContainerStatus = "exited"
	s.ServerStatus = "stopped"
	s.ServerStartedAt = time.Time{}
	h.Instance.Set(s)
	h.Instance.ClearDesired(id, "exited")
	_ = h.Instance.Save()
	slog.Info("server stopped", "server_id", id)
}

// handleStopServer stops a running server container asynchronously.
func (h *Handler) HandleStopServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	s, ok := h.Instance.TrySetDesired(id, "exited")
	if !ok {
		jsonError(w, "operation in progress", http.StatusConflict)
		return
	}
	_ = h.Instance.Save()
	go h.doStop(id)
	jsonResponse(w, s)
}

func (h *Handler) doForceStop(id string) {
	s, _ := h.Instance.Get(id)
	prevStatus := s.Status

	defer func() {
		if r := recover(); r != nil {
			slog.Error("force-stop server panic", "server_id", id, "recover", r)
			h.Instance.ClearDesired(id, prevStatus)
			_ = h.Instance.Save()
		}
	}()

	if s.ContainerID == "" {
		slog.Warn("force-stop server: container not created", "server_id", id)
		h.Instance.ClearDesired(id, prevStatus)
		_ = h.Instance.Save()
		return
	}
	if err := h.Cli.ContainerKill(context.Background(), s.ContainerID, "SIGKILL"); err != nil {
		if client.IsErrNotFound(err) {
			slog.Warn("container already removed", "server_id", id, "container_id", s.ContainerID)
		} else {
			slog.Error("failed to force-stop container", "server_id", id, "error", err)
			h.Instance.ClearDesired(id, prevStatus)
			_ = h.Instance.Save()
			return
		}
	}

	s.ContainerStatus = "exited"
	s.ServerStatus = "stopped"
	s.ServerStartedAt = time.Time{}
	h.Instance.Set(s)
	h.Instance.ClearDesired(id, "exited")
	_ = h.Instance.Save()
	slog.Info("server force-stopped", "server_id", id)
}

// HandleForceStopServer force-stops a running server container asynchronously.
func (h *Handler) HandleForceStopServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	s, ok := h.Instance.TrySetDesired(id, "exited")
	if !ok {
		jsonError(w, "operation in progress", http.StatusConflict)
		return
	}
	_ = h.Instance.Save()
	go h.doForceStop(id)
	jsonResponse(w, s)
}

func (h *Handler) doRestart(id string) {
	s, _ := h.Instance.Get(id)
	prevStatus := s.Status

	defer func() {
		if r := recover(); r != nil {
			slog.Error("restart server panic", "server_id", id, "recover", r)
			h.Instance.ClearDesired(id, prevStatus)
			_ = h.Instance.Save()
		}
	}()

	if s.ContainerID == "" {
		slog.Warn("restart server: container not created", "server_id", id)
		h.Instance.ClearDesired(id, prevStatus)
		_ = h.Instance.Save()
		return
	}
	timeout := 30
	_ = h.Cli.ContainerStop(context.Background(), s.ContainerID, container.StopOptions{Timeout: &timeout})
	if err := h.Cli.ContainerRemove(context.Background(), s.ContainerID, container.RemoveOptions{Force: true}); err != nil {
		slog.Warn("failed to remove container on restart", "server_id", id, "container_id", s.ContainerID, "error", err)
	}
	s.ContainerID = ""
	s.ContainerStatus = ""
	h.Instance.Set(s)
	h.Instance.ClearDesired(id, prevStatus)
	_ = h.Instance.Save()
	h.doStart(id)
}

// handleRestartServer restarts a server container asynchronously.
func (h *Handler) HandleRestartServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	s, ok := h.Instance.TrySetDesired(id, "running")
	if !ok {
		jsonError(w, "operation in progress", http.StatusConflict)
		return
	}
	_ = h.Instance.Save()
	go h.doRestart(id)
	jsonResponse(w, s)
}

// handleDeleteServer removes a server container and optionally wipes its data.
func (h *Handler) HandleDeleteServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	wipe := r.URL.Query().Get("wipe") == "true"
	s, ok := h.Instance.Get(id)
	if !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if s.ContainerID != "" {
		if err := h.Cli.ContainerRemove(r.Context(), s.ContainerID, container.RemoveOptions{Force: true}); err != nil {
			slog.Warn("delete server: failed to remove container", "server_id", id, "error", err)
		} else {
			slog.Info("server container removed", "server_id", id, "container_id", s.ContainerID[:12])
		}
	}
	if wipe && s.VolumePath != "" {
		if err := os.RemoveAll(s.VolumePath); err != nil {
			slog.Warn("delete server: failed to wipe volume", "server_id", id, "path", s.VolumePath, "error", err)
		}
	}
	h.Instance.Delete(id)
	_ = h.Instance.Save()
	slog.Info("server deleted", "server_id", id, "wipe", wipe)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) doRecreateWorld(id string) {
	s, _ := h.Instance.Get(id)
	prevStatus := s.Status

	defer func() {
		if r := recover(); r != nil {
			slog.Error("recreate world panic", "server_id", id, "recover", r)
			h.Instance.ClearDesired(id, prevStatus)
			_ = h.Instance.Save()
		}
	}()

	if s.ContainerID == "" {
		slog.Warn("recreate world: container not created", "server_id", id)
		h.Instance.ClearDesired(id, prevStatus)
		_ = h.Instance.Save()
		return
	}

	if s.ContainerStatus == "running" {
		timeout := 30
		if err := h.Cli.ContainerStop(context.Background(), s.ContainerID, container.StopOptions{Timeout: &timeout}); err != nil {
			if !client.IsErrNotFound(err) {
				slog.Error("failed to stop container for world recreate", "server_id", id, "error", err)
				h.Instance.ClearDesired(id, prevStatus)
				_ = h.Instance.Save()
				return
			}
		}
		s.ContainerStatus = "exited"
		s.ServerStatus = "stopped"
		s.ServerStartedAt = time.Time{}
		h.Instance.Set(s)
	}

	if s.VolumePath != "" {
		dirs := []string{"world", "world_nether", "world_the_end"}
		for _, dir := range dirs {
			path := filepath.Join(s.VolumePath, dir)
			if err := os.RemoveAll(path); err != nil {
				slog.Warn("failed to remove world dir", "server_id", id, "path", path, "error", err)
			} else {
				slog.Info("removed world dir", "server_id", id, "dir", dir)
			}
		}
	}

	if err := h.Cli.ContainerStart(context.Background(), s.ContainerID, container.StartOptions{}); err != nil {
		if client.IsErrNotFound(err) {
			slog.Warn("container missing on world recreate, recreating", "server_id", id)
			s.ContainerID = ""
			h.Instance.Set(s)
			h.Instance.ClearDesired(id, prevStatus)
			_ = h.Instance.Save()
			h.doStart(id)
			return
		}
		slog.Error("failed to start container after world recreate", "server_id", id, "error", err)
		h.Instance.ClearDesired(id, prevStatus)
		_ = h.Instance.Save()
		return
	}

	s.ContainerStatus = "running"
	s.ContainerStartedAt = time.Now().UTC()
	s.ServerStatus = "starting"
	s.PendingProperties = nil
	h.Instance.Set(s)
	h.Instance.ClearDesired(id, "running")
	_ = h.Instance.Save()
	slog.Info("world recreated", "server_id", id)
}

// handleRecreateWorld stops the server, deletes world data, and restarts it.
func (h *Handler) HandleRecreateWorld(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Instance.Get(id); !ok {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	_, ok := h.Instance.TrySetDesired(id, "running")
	if !ok {
		jsonError(w, "operation in progress", http.StatusConflict)
		return
	}
	_ = h.Instance.Save()
	go h.doRecreateWorld(id)
	w.WriteHeader(http.StatusNoContent)
}
