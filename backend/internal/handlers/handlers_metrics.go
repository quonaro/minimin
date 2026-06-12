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
	rconCache map[string]*runner.RCONClient      // keyed by serverID
	mu        sync.Mutex
}

// NewMetricsPoller creates a poller attached to the given handler.
func NewMetricsPoller(h *Handler) *MetricsPoller {
	return &MetricsPoller{
		h:         h,
		interval:  2 * time.Second,
		prevStats: make(map[string]container.StatsResponse),
		rconCache: make(map[string]*runner.RCONClient),
	}
}

// Start runs the polling loop until ctx is cancelled.
func (p *MetricsPoller) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	defer p.closeAllRCON()
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
	runningServerIDs := make(map[string]bool, len(servers))

	for _, s := range servers {
		if s.ContainerID == "" || s.ContainerStatus != "running" {
			continue
		}
		runningIDs[s.ContainerID] = true
		runningServerIDs[s.ServerID] = true

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
			resp, rconErr := p.executeOrReconnect(s.ServerID, addr, s.RconPassword, "list")
			if rconErr != nil {
				slog.Warn("metrics poll: rcon list failed", "server", s.ServerID, "error", rconErr)
			} else {
				online, maxPlayers, _ = parseListResponse(resp)
				if maxPlayers == 0 {
					slog.Warn("metrics poll: list response unparsed", "server", s.ServerID, "response", resp)
				}
			}
			tpsResp, tpsErr := p.executeOrReconnect(s.ServerID, addr, s.RconPassword, "tps")
			if tpsErr != nil {
				slog.Warn("metrics poll: rcon tps failed", "server", s.ServerID, "error", tpsErr)
			} else {
				tps = parseTPSOutput(tpsResp)
			}
		}

		payload := events.MetricsPayload{
			ServerID:  s.ServerID,
			RAMUsage:  ramUsage,
			RAMLimit:  ramLimit,
			CPU:       cpuPercent,
			CPUs:      s.CPUs,
			Online:    online,
			Max:       maxPlayers,
			TPS:       tps,
			Timestamp: time.Now().UTC(),
		}

		if p.h.EventsHub != nil {
			p.h.EventsHub.BroadcastJSON("metrics", payload)
		}
	}

	// Clean up stale entries for containers/servers that are no longer running.
	p.mu.Lock()
	for cid := range p.prevStats {
		if !runningIDs[cid] {
			delete(p.prevStats, cid)
		}
	}
	for sid, client := range p.rconCache {
		if !runningServerIDs[sid] {
			_ = client.Close()
			delete(p.rconCache, sid)
		}
	}
	p.mu.Unlock()
}

func (p *MetricsPoller) closeAllRCON() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, client := range p.rconCache {
		_ = client.Close()
	}
	p.rconCache = make(map[string]*runner.RCONClient)
}

func (p *MetricsPoller) getRCONClient(serverID, addr, password string) (*runner.RCONClient, error) {
	p.mu.Lock()
	if client, ok := p.rconCache[serverID]; ok {
		p.mu.Unlock()
		return client, nil
	}
	p.mu.Unlock()

	client, err := runner.DialRCON(addr, password, 3*time.Second)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	if existing, ok := p.rconCache[serverID]; ok {
		p.mu.Unlock()
		_ = client.Close()
		return existing, nil
	}
	p.rconCache[serverID] = client
	p.mu.Unlock()
	return client, nil
}

func (p *MetricsPoller) executeOrReconnect(serverID, addr, password, cmd string) (string, error) {
	client, err := p.getRCONClient(serverID, addr, password)
	if err != nil {
		return "", err
	}
	resp, err := client.Execute(cmd)
	if err != nil {
		_ = client.Close()
		p.mu.Lock()
		delete(p.rconCache, serverID)
		p.mu.Unlock()

		client, err = runner.DialRCON(addr, password, 3*time.Second)
		if err != nil {
			return "", err
		}
		p.mu.Lock()
		p.rconCache[serverID] = client
		p.mu.Unlock()
		return client.Execute(cmd)
	}
	return resp, nil
}
