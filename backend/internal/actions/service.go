package actions

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"orchestrator/internal/runner"
	"orchestrator/internal/state"
)

// Service provides server lifecycle operations.
type Service struct {
	instance       *state.InstanceFile
	cli            *client.Client
	serversDir     string
	serversHostDir string
	networkName    string
	pullMu         sync.RWMutex
	pullProgress   map[string]*runner.ImagePullProgress
}

// NewService creates a new actions service.
func NewService(instance *state.InstanceFile, cli *client.Client, serversDir, serversHostDir, networkName string) *Service {
	return &Service{
		instance:       instance,
		cli:            cli,
		serversDir:     serversDir,
		serversHostDir: serversHostDir,
		networkName:    networkName,
		pullProgress:   make(map[string]*runner.ImagePullProgress),
	}
}

// GetPullProgress returns the current image pull progress for a server.
func (s *Service) GetPullProgress(id string) *runner.ImagePullProgress {
	s.pullMu.RLock()
	defer s.pullMu.RUnlock()
	return s.pullProgress[id]
}

// Start launches or restarts the server container.
func (s *Service) Start(ctx context.Context, id string, removeExisting bool) {
	srv, _ := s.instance.Get(id)
	prevStatus := srv.Status

	defer func() {
		if r := recover(); r != nil {
			slog.Error("start server panic", "server_id", id, "recover", r)
			s.instance.ClearDesired(id, prevStatus)
			_ = s.instance.Save()
		}
	}()

	if removeExisting && srv.ContainerID != "" {
		if err := s.cli.ContainerRemove(ctx, srv.ContainerID, container.RemoveOptions{Force: true}); err != nil {
			slog.Warn("failed to remove existing container on start", "server_id", id, "container_id", srv.ContainerID, "error", err)
		}
		srv.ContainerID = ""
		srv.ContainerStatus = ""
		srv.ContainerStartedAt = time.Time{}
		s.instance.Set(srv)
	}

	if srv.ContainerID != "" {
		if err := s.cli.ContainerStart(ctx, srv.ContainerID, container.StartOptions{}); err != nil {
			if client.IsErrNotFound(err) {
				slog.Warn("stale container id, recreating", "server_id", id, "container_id", srv.ContainerID)
				srv.ContainerID = ""
			} else {
				slog.Error("failed to start container", "server_id", id, "error", err)
				s.instance.ClearDesired(id, prevStatus)
				_ = s.instance.Save()
				return
			}
		}
	}

	if srv.ContainerID == "" {
		if srv.RconPort == 0 {
			srv.RconPort = srv.GamePort + 10
		}
		portUsed := func(p uint16) bool {
			return s.instance.IsPortUsed(p, id)
		}
		gamePort, err := runner.FindFreePortExcluding("", srv.GamePort, portUsed)
		if err != nil {
			slog.Error("no free game port", "server_id", id, "error", err)
			s.instance.ClearDesired(id, prevStatus)
			_ = s.instance.Save()
			return
		}
		rconHost := "127.0.0.1"
		if srv.PublicRcon {
			rconHost = ""
		}
		rconPort, err := runner.FindFreePortExcluding(rconHost, srv.RconPort, portUsed)
		if err != nil {
			slog.Error("no free rcon port", "server_id", id, "error", err)
			s.instance.ClearDesired(id, prevStatus)
			_ = s.instance.Save()
			return
		}
		if rconPort == gamePort {
			rconPort, err = runner.FindFreePortExcluding(rconHost, 0, portUsed)
			if err != nil {
				slog.Error("no free rcon port", "server_id", id, "error", err)
				s.instance.ClearDesired(id, prevStatus)
				_ = s.instance.Save()
				return
			}
		}
		srv.GamePort = gamePort
		srv.RconPort = rconPort
		if srv.RconPassword == "" {
			pw, _ := runner.GenerateRconPassword()
			srv.RconPassword = pw
		}
		if srv.VolumePath != "" {
			uid, gid := runner.ContainerUIDGID()
			if err := runner.FixVolumeOwnership(ctx, s.cli, srv.VolumePath, uid, gid); err != nil {
				slog.Warn("failed to fix volume ownership", "server_id", id, "path", srv.VolumePath, "error", err)
			}
		}
		srv.ServerStatus = "pulling_image"
		s.instance.Set(srv)

		s.pullMu.Lock()
		s.pullProgress[id] = &runner.ImagePullProgress{}
		s.pullMu.Unlock()
		defer func() {
			s.pullMu.Lock()
			delete(s.pullProgress, id)
			s.pullMu.Unlock()
		}()

		if err := runner.PullImageWithProgress(ctx, s.cli, runner.ImageName, func(current, total int64) {
			s.pullMu.Lock()
			s.pullProgress[id] = &runner.ImagePullProgress{Current: current, Total: total}
			s.pullMu.Unlock()
		}); err != nil {
			slog.Error("failed to pull image", "server_id", id, "error", err)
			s.instance.ClearDesired(id, prevStatus)
			_ = s.instance.Save()
			return
		}

		containerID, volumeID, volumePath, err := runner.StartServerContainer(
			ctx, s.cli, srv.ServerID,
			srv.RamBytes, srv.GamePort,
			srv.EngineType, srv.GameVersion, srv.LoaderVersion,
			s.serversDir, s.serversHostDir,
			srv.RconPort, srv.RconPassword, srv.PublicRcon,
			srv.VolumePath,
			nil,
			srv.RestartPolicy,
			s.networkName,
			srv.ExternalJavaArgs,
		)
		if err != nil {
			slog.Error("failed to start server container", "server_id", id, "error", err)
			s.instance.ClearDesired(id, prevStatus)
			_ = s.instance.Save()
			return
		}
		srv.ContainerID = containerID
		srv.VolumeID = volumeID
		srv.VolumePath = volumePath
		srv.HostPath = runner.HostPathForDocker(volumePath, s.serversDir, s.serversHostDir)
		srv.ContainerPath = "/data"
		srv.ModCount = state.CountMods(srv)
		srv.ImageName = runner.ImageName
	}

	srv.ContainerStatus = "running"
	srv.ContainerStartedAt = time.Now().UTC()
	srv.ServerStatus = "starting"
	srv.ImageName = runner.ImageName
	srv.PendingProperties = nil
	s.instance.Set(srv)
	s.instance.ClearDesired(id, "running")
	_ = s.instance.Save()
	slog.Info("server started", "server_id", id)
}

// Stop gracefully stops a running server container.
func (s *Service) Stop(ctx context.Context, id string) {
	srv, _ := s.instance.Get(id)
	prevStatus := srv.Status

	defer func() {
		if r := recover(); r != nil {
			slog.Error("stop server panic", "server_id", id, "recover", r)
			s.instance.ClearDesired(id, prevStatus)
			_ = s.instance.Save()
		}
	}()

	if srv.ContainerID == "" {
		slog.Warn("stop server: container not created", "server_id", id)
		s.instance.ClearDesired(id, prevStatus)
		_ = s.instance.Save()
		return
	}
	timeout := 30
	if err := s.cli.ContainerStop(ctx, srv.ContainerID, container.StopOptions{Timeout: &timeout}); err != nil {
		if client.IsErrNotFound(err) {
			slog.Warn("container already removed", "server_id", id, "container_id", srv.ContainerID)
		} else {
			slog.Error("failed to stop container", "server_id", id, "error", err)
			s.instance.ClearDesired(id, prevStatus)
			_ = s.instance.Save()
			return
		}
	}

	srv.ContainerStatus = "exited"
	srv.ServerStatus = "stopped"
	srv.ServerStartedAt = time.Time{}
	s.instance.Set(srv)
	s.instance.ClearDesired(id, "exited")
	_ = s.instance.Save()
	slog.Info("server stopped", "server_id", id)
}

// ForceStop kills a running server container immediately.
func (s *Service) ForceStop(ctx context.Context, id string) {
	srv, _ := s.instance.Get(id)
	prevStatus := srv.Status

	defer func() {
		if r := recover(); r != nil {
			slog.Error("force-stop server panic", "server_id", id, "recover", r)
			s.instance.ClearDesired(id, prevStatus)
			_ = s.instance.Save()
		}
	}()

	if srv.ContainerID == "" {
		slog.Warn("force-stop server: container not created", "server_id", id)
		s.instance.ClearDesired(id, prevStatus)
		_ = s.instance.Save()
		return
	}
	if err := s.cli.ContainerKill(ctx, srv.ContainerID, "SIGKILL"); err != nil {
		if client.IsErrNotFound(err) {
			slog.Warn("container already removed", "server_id", id, "container_id", srv.ContainerID)
		} else {
			slog.Error("failed to force-stop container", "server_id", id, "error", err)
			s.instance.ClearDesired(id, prevStatus)
			_ = s.instance.Save()
			return
		}
	}

	srv.ContainerStatus = "exited"
	srv.ServerStatus = "stopped"
	srv.ServerStartedAt = time.Time{}
	s.instance.Set(srv)
	s.instance.ClearDesired(id, "exited")
	_ = s.instance.Save()
	slog.Info("server force-stopped", "server_id", id)
}

// Restart stops and recreates a server container.
func (s *Service) Restart(ctx context.Context, id string) {
	srv, _ := s.instance.Get(id)
	prevStatus := srv.Status

	defer func() {
		if r := recover(); r != nil {
			slog.Error("restart server panic", "server_id", id, "recover", r)
			s.instance.ClearDesired(id, prevStatus)
			_ = s.instance.Save()
		}
	}()

	if srv.ContainerID == "" {
		slog.Warn("restart server: container not created", "server_id", id)
		s.instance.ClearDesired(id, prevStatus)
		_ = s.instance.Save()
		return
	}
	timeout := 30
	_ = s.cli.ContainerStop(ctx, srv.ContainerID, container.StopOptions{Timeout: &timeout})
	if err := s.cli.ContainerRemove(ctx, srv.ContainerID, container.RemoveOptions{Force: true}); err != nil {
		slog.Warn("failed to remove container on restart", "server_id", id, "container_id", srv.ContainerID, "error", err)
	}
	srv.ContainerID = ""
	srv.ContainerStatus = ""
	s.instance.Set(srv)
	s.instance.ClearDesired(id, prevStatus)
	_ = s.instance.Save()
	s.Start(ctx, id, false)
}

// RecreateWorld stops the server, deletes world data, and restarts it.
func (s *Service) RecreateWorld(ctx context.Context, id string) {
	srv, _ := s.instance.Get(id)
	prevStatus := srv.Status

	defer func() {
		if r := recover(); r != nil {
			slog.Error("recreate world panic", "server_id", id, "recover", r)
			s.instance.ClearDesired(id, prevStatus)
			_ = s.instance.Save()
		}
	}()

	if srv.ContainerID == "" {
		slog.Warn("recreate world: container not created", "server_id", id)
		s.instance.ClearDesired(id, prevStatus)
		_ = s.instance.Save()
		return
	}

	if srv.ContainerStatus == "running" {
		timeout := 30
		if err := s.cli.ContainerStop(ctx, srv.ContainerID, container.StopOptions{Timeout: &timeout}); err != nil {
			if !client.IsErrNotFound(err) {
				slog.Error("failed to stop container for world recreate", "server_id", id, "error", err)
				s.instance.ClearDesired(id, prevStatus)
				_ = s.instance.Save()
				return
			}
		}
		srv.ContainerStatus = "exited"
		srv.ServerStatus = "stopped"
		srv.ServerStartedAt = time.Time{}
		s.instance.Set(srv)
	}

	if srv.VolumePath != "" {
		dirs := []string{"world", "world_nether", "world_the_end"}
		for _, dir := range dirs {
			path := filepath.Join(srv.VolumePath, dir)
			if err := os.RemoveAll(path); err != nil {
				slog.Warn("failed to remove world dir", "server_id", id, "path", path, "error", err)
			} else {
				slog.Info("removed world dir", "server_id", id, "dir", dir)
			}
		}
	}

	if err := s.cli.ContainerStart(ctx, srv.ContainerID, container.StartOptions{}); err != nil {
		if client.IsErrNotFound(err) {
			slog.Warn("container missing on world recreate, recreating", "server_id", id)
			srv.ContainerID = ""
			s.instance.Set(srv)
			s.instance.ClearDesired(id, prevStatus)
			_ = s.instance.Save()
			s.Start(ctx, id, false)
			return
		}
		slog.Error("failed to start container after world recreate", "server_id", id, "error", err)
		s.instance.ClearDesired(id, prevStatus)
		_ = s.instance.Save()
		return
	}

	srv.ContainerStatus = "running"
	srv.ContainerStartedAt = time.Now().UTC()
	srv.ServerStatus = "starting"
	srv.PendingProperties = nil
	s.instance.Set(srv)
	s.instance.ClearDesired(id, "running")
	_ = s.instance.Save()
	slog.Info("world recreated", "server_id", id)
}

// Delete removes a server container and optionally wipes its data.
func (s *Service) Delete(ctx context.Context, id string, wipe bool) error {
	srv, ok := s.instance.Get(id)
	if !ok {
		return fmt.Errorf("not found")
	}
	if srv.ContainerID != "" {
		if err := s.cli.ContainerRemove(ctx, srv.ContainerID, container.RemoveOptions{Force: true}); err != nil {
			slog.Warn("delete server: failed to remove container", "server_id", id, "error", err)
		} else {
			slog.Info("server container removed", "server_id", id, "container_id", srv.ContainerID[:12])
		}
	}
	if wipe && srv.VolumePath != "" {
		if err := os.RemoveAll(srv.VolumePath); err != nil {
			slog.Warn("delete server: failed to wipe volume", "server_id", id, "path", srv.VolumePath, "error", err)
		}
	}
	s.instance.Delete(id)
	_ = s.instance.Save()
	slog.Info("server deleted", "server_id", id, "wipe", wipe)
	return nil
}
