package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"orchestrator/internal/events"
	"orchestrator/internal/runner"

	"github.com/docker/docker/api/types/container"
)

var tpsFloatRe = regexp.MustCompile(`\d+\.\d+`)

func parseTPSOutput(resp string) *float64 {
	if !strings.Contains(strings.ToLower(resp), "tps") {
		return nil
	}
	m := tpsFloatRe.FindString(resp)
	if m == "" {
		return nil
	}
	v, err := strconv.ParseFloat(m, 64)
	if err != nil {
		return nil
	}
	return &v
}

// MetricsPoller collects container stats and RCON data and broadcasts metrics events.
type MetricsPoller struct {
	h         *Handler
	interval  time.Duration
	prevStats map[string]container.StatsResponse // keyed by containerID
	mu        sync.Mutex
}

// NewMetricsPoller creates a poller attached to the given handler.
func NewMetricsPoller(h *Handler) *MetricsPoller {
	return &MetricsPoller{
		h:         h,
		interval:  5 * time.Second,
		prevStats: make(map[string]container.StatsResponse),
	}
}

// Start runs the polling loop until ctx is cancelled.
func (p *MetricsPoller) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

func (p *MetricsPoller) poll(ctx context.Context) {
	servers := p.h.Instance.All()
	runningIDs := make(map[string]bool, len(servers))

	for _, s := range servers {
		if s.ContainerID == "" || s.ContainerStatus != "running" {
			continue
		}
		runningIDs[s.ContainerID] = true

		statsReader, err := p.h.Cli.ContainerStats(ctx, s.ContainerID, false)
		if err != nil {
			slog.Debug("metrics poll: container stats failed", "server", s.ServerID, "error", err)
			continue
		}
		var stats container.StatsResponse
		if decErr := json.NewDecoder(statsReader.Body).Decode(&stats); decErr != nil {
			_ = statsReader.Body.Close()
			slog.Debug("metrics poll: decode stats failed", "server", s.ServerID, "error", decErr)
			continue
		}
		_ = statsReader.Body.Close()

		ramUsage := stats.MemoryStats.Usage
		ramLimit := stats.MemoryStats.Limit
		if ramLimit == 0 {
			ramLimit = 1
		}

		cpuPercent := 0.0
		p.mu.Lock()
		prev, hasPrev := p.prevStats[s.ContainerID]
		p.prevStats[s.ContainerID] = stats
		p.mu.Unlock()

		if hasPrev {
			cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(prev.CPUStats.CPUUsage.TotalUsage)
			systemDelta := float64(stats.CPUStats.SystemUsage) - float64(prev.CPUStats.SystemUsage)
			if systemDelta > 0 && cpuDelta > 0 {
				onlineCPUs := uint32(len(stats.CPUStats.CPUUsage.PercpuUsage))
				if onlineCPUs == 0 {
					onlineCPUs = stats.CPUStats.OnlineCPUs
				}
				if onlineCPUs == 0 {
					onlineCPUs = 1
				}
				cpuPercent = (cpuDelta / systemDelta) * float64(onlineCPUs) * 100.0
			}
		}

		online := 0
		maxPlayers := 0
		var tps *float64

		if s.ServerStatus == "running" {
			addr := fmt.Sprintf("mc-srv-%s:25575", s.ServerID)
			rconClient, err := runner.DialRCON(addr, s.RconPassword, 3*time.Second)
			if err != nil {
				slog.Warn("metrics poll: rcon dial failed", "server", s.ServerID, "error", err)
			} else {
				resp, rconErr := rconClient.Execute("list")
				if rconErr != nil {
					slog.Warn("metrics poll: rcon list failed", "server", s.ServerID, "error", rconErr)
				} else {
					online, maxPlayers, _ = parseListResponse(resp)
					if maxPlayers == 0 {
						slog.Warn("metrics poll: list response unparsed", "server", s.ServerID, "response", resp)
					}
				}
				tpsResp, tpsErr := rconClient.Execute("tps")
				if tpsErr != nil {
					slog.Warn("metrics poll: rcon tps failed", "server", s.ServerID, "error", tpsErr)
				} else {
					tps = parseTPSOutput(tpsResp)
				}
				_ = rconClient.Close()
			}
		}

		payload := events.MetricsPayload{
			ServerID:  s.ServerID,
			RAMUsage:  ramUsage,
			RAMLimit:  ramLimit,
			CPU:       cpuPercent,
			Online:    online,
			Max:       maxPlayers,
			TPS:       tps,
			Timestamp: time.Now().UTC(),
		}

		if p.h.EventsHub != nil {
			p.h.EventsHub.BroadcastJSON("metrics", payload)
		}
	}

	// Clean up stale entries for containers that are no longer running.
	p.mu.Lock()
	for cid := range p.prevStats {
		if !runningIDs[cid] {
			delete(p.prevStats, cid)
		}
	}
	p.mu.Unlock()
}
