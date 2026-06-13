package health

import (
	"context"
	"log/slog"
	"time"

	"orchestrator/internal/runner"
	"orchestrator/internal/state"
)

// Checker periodically pings running Minecraft servers and updates their status.
type Checker struct {
	Instance         *state.InstanceFile
	BroadcastPlayerData func(serverID string)
}

// NewChecker creates a Checker.
func NewChecker(instance *state.InstanceFile, broadcastFn func(string)) *Checker {
	return &Checker{
		Instance:         instance,
		BroadcastPlayerData: broadcastFn,
	}
}

// Start runs the health-check loop until ctx is cancelled.
func (c *Checker) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	failures := make(map[string]int)
	broadcastTick := 0
	for {
		select {
		case <-ticker.C:
			broadcastTick++
			active := make(map[string]bool, len(c.Instance.All()))
			for _, s := range c.Instance.All() {
				active[s.ServerID] = true
				if s.ContainerStatus != "running" {
					continue
				}
				port := s.GamePort
				if port == 0 {
					port = 25565
				}
				ok, pingErr := runner.TryPingServer(s.ServerID, port, 5*time.Second)
				if ok {
					failures[s.ServerID] = 0
					if s.ServerStatus != "running" {
						s.ServerStatus = "running"
						s.ServerStartedAt = time.Now().UTC()
						c.Instance.Set(s)
						_ = c.Instance.Save()
						slog.Info("server ready", "server_id", s.ServerID, "status", s.ServerStatus, "volume_path", s.VolumePath)
					}
				} else {
					failures[s.ServerID]++
					if pingErr != nil {
						slog.Debug("server ping failed", "server_id", s.ServerID, "error", pingErr, "consecutive_failures", failures[s.ServerID])
					}
					if s.ServerStatus == "running" && failures[s.ServerID] >= 3 {
						s.ServerStatus = "starting"
						s.ServerStartedAt = time.Time{}
						c.Instance.Set(s)
						_ = c.Instance.Save()
						slog.Info("server not ready", "server_id", s.ServerID, "status", s.ServerStatus, "consecutive_failures", failures[s.ServerID])
					}
				}
				// Push player data every 30 seconds.
				if broadcastTick%3 == 0 && c.BroadcastPlayerData != nil {
					c.BroadcastPlayerData(s.ServerID)
				}
			}
			for id := range failures {
				if !active[id] {
					delete(failures, id)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
