// Package scheduler provides task scheduling and automation for servers.
package scheduler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/robfig/cron/v3"

	"orchestrator/internal/actionlog"
	"orchestrator/internal/actions"
	"orchestrator/internal/backup"
	"orchestrator/internal/events"
	"orchestrator/internal/runner"
	"orchestrator/internal/state"
)

// Service orchestrates scheduled and event-driven tasks.
type Service struct {
	Store       *Store
	instance    *state.InstanceFile
	cli         *client.Client
	actions     *actions.Service
	backup      *backup.Service
	hub         *events.Hub
	logStore    *actionlog.Store
	rconTimeout time.Duration

	cron      *cron.Cron
	tickers   map[string]*time.Ticker
	stopChans map[string]chan struct{}
	mu        sync.Mutex
}

// NewService creates a scheduler service.
func NewService(
	store *Store,
	instance *state.InstanceFile,
	cli *client.Client,
	act *actions.Service,
	bk *backup.Service,
	hub *events.Hub,
	logStore *actionlog.Store,
) *Service {
	return &Service{
		Store:       store,
		instance:    instance,
		cli:         cli,
		actions:     act,
		backup:      bk,
		hub:         hub,
		logStore:    logStore,
		rconTimeout: 10 * time.Second,
		cron:        cron.New(),
		tickers:     make(map[string]*time.Ticker),
		stopChans:   make(map[string]chan struct{}),
	}
}

// TasksForServer returns tasks filtered by server ID.
func (s *Service) TasksForServer(serverID string) []Task {
	return s.Store.AllForServer(serverID)
}

// Start loads enabled tasks and begins scheduling.
func (s *Service) Start(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.Store.All() {
		if !t.Enabled {
			continue
		}
		s.scheduleLocked(t)
	}

	s.cron.Start()

	// Event-based tasks
	if s.hub != nil {
		go s.eventListener(ctx)
	}

	// Wait for shutdown
	go func() {
		<-ctx.Done()
		s.Stop()
	}()
}

// Stop halts all tickers and cron jobs.
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := s.cron.Stop()
	<-ctx.Done()

	for id, ch := range s.stopChans {
		close(ch)
		delete(s.stopChans, id)
	}
	for id, t := range s.tickers {
		t.Stop()
		delete(s.tickers, id)
	}
}

// Reload refreshes the schedule after external changes to the store.
func (s *Service) Reload() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop existing tickers for this store
	for id, ch := range s.stopChans {
		close(ch)
		delete(s.stopChans, id)
	}
	for id, t := range s.tickers {
		t.Stop()
		delete(s.tickers, id)
	}

	for _, t := range s.Store.All() {
		if !t.Enabled {
			continue
		}
		s.scheduleLocked(t)
	}
}

// scheduleLocked registers a task based on its trigger type.
// Caller must hold s.mu.
func (s *Service) scheduleLocked(t Task) {
	switch t.Trigger.Type {
	case TriggerInterval:
		d, err := time.ParseDuration(t.Trigger.Interval)
		if err != nil {
			slog.Warn("invalid interval, skipping task", "task_id", t.ID, "interval", t.Trigger.Interval, "error", err)
			return
		}
		stopCh := make(chan struct{})
		s.stopChans[t.ID] = stopCh
		s.tickers[t.ID] = time.NewTicker(d)
		go s.runInterval(t.ID, s.tickers[t.ID], stopCh)

	case TriggerCron:
		_, err := s.cron.AddFunc(t.Trigger.Cron, func() {
			s.execute(t)
		})
		if err != nil {
			slog.Warn("invalid cron, skipping task", "task_id", t.ID, "cron", t.Trigger.Cron, "error", err)
		}

	case TriggerEvent:
		// Event tasks are handled by eventListener.
	}
}

func (s *Service) runInterval(taskID string, ticker *time.Ticker, stop <-chan struct{}) {
	for {
		select {
		case <-ticker.C:
			t, ok := s.Store.Get(taskID)
			if !ok || !t.Enabled {
				return
			}
			s.execute(t)
		case <-stop:
			return
		}
	}
}

func (s *Service) eventListener(ctx context.Context) {
	id, ch := s.hub.Register()
	defer s.hub.Unregister(id)
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				return
			}
			s.handleEvent(ev)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) handleEvent(ev events.Event) {
	if ev.Type != "server" {
		return
	}
	payload, ok := ev.Payload.(state.ServerState)
	if !ok {
		return
	}
	for _, t := range s.Store.AllForServer(payload.ServerID) {
		if !t.Enabled || t.Trigger.Type != TriggerEvent {
			continue
		}
		// Map server state changes to event names
		matched := false
		switch t.Trigger.Event {
		case "server_start":
			if payload.ServerStatus == "running" && payload.ContainerStatus == "running" {
				matched = true
			}
		case "server_stop":
			if payload.ContainerStatus == "exited" {
				matched = true
			}
		case "server_restart":
			// Restart is detected by DesiredStatus transitions; handled separately if needed.
		}
		if matched {
			s.execute(t)
		}
	}
}

// RunNow triggers a task immediately regardless of schedule.
func (s *Service) RunNow(id string) error {
	t, ok := s.Store.Get(id)
	if !ok {
		return fmt.Errorf("task not found")
	}
	go s.execute(t)
	return nil
}

// execute runs the task action with preconditions.
func (s *Service) execute(t Task) {
	server, ok := s.instance.Get(t.ServerID)
	if !ok {
		slog.Warn("task server not found, skipping", "task_id", t.ID, "server_id", t.ServerID)
		s.log(t, "skipped", "server not found")
		return
	}

	slog.Info("executing task", "task_id", t.ID, "server_id", t.ServerID, "action_type", t.Action.Type)

	err := s.runAction(server, t.Action)
	if err != nil {
		slog.Error("task execution failed", "task_id", t.ID, "error", err)
		s.log(t, "error", err.Error())
		return
	}

	s.log(t, "success", "")

	now := time.Now().UTC()
	t.LastRun = now
	// Recalculate NextRun if interval/cron
	switch t.Trigger.Type {
	case TriggerInterval:
		d, _ := time.ParseDuration(t.Trigger.Interval)
		t.NextRun = now.Add(d)
	case TriggerCron:
		// NextRun for cron is best-effort; cron lib handles actual scheduling.
		// We could parse cron spec to compute next occurrence, but keep it simple.
	}
	s.Store.Put(t)
	if saveErr := s.Store.Save(); saveErr != nil {
		slog.Error("failed to save actions file after task run", "task_id", t.ID, "error", saveErr)
	}
}

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Service) log(t Task, status, msg string) {
	if s.logStore == nil {
		return
	}
	detail := ""
	switch t.Action.Type {
	case ActionRCON:
		detail = t.Action.Command
	case ActionContainerExec:
		detail = t.Action.Command
	case ActionLifecycle:
		detail = t.Action.Lifecycle
	case ActionBackup:
		detail = "backup world"
	}
	_ = s.logStore.Append(actionlog.Entry{
		ID:        generateUUID(),
		Timestamp: time.Now().UTC(),
		ServerID:  t.ServerID,
		TaskID:    t.ID,
		Source:    string(t.Trigger.Type),
		Action:    string(t.Action.Type),
		Detail:    detail,
		Status:    status,
		Message:   msg,
	})
}

func (s *Service) runAction(server state.ServerState, action Action) error {
	switch action.Type {
	case ActionRCON:
		if server.ServerStatus != "running" || server.RconPort == 0 {
			return fmt.Errorf("rcon unavailable: server not running or no rcon port")
		}
		addr := fmt.Sprintf("127.0.0.1:%d", server.RconPort)
		if server.PublicRcon {
			addr = fmt.Sprintf("0.0.0.0:%d", server.RconPort)
		}
		client, err := runner.DialRCON(addr, server.RconPassword, s.rconTimeout)
		if err != nil {
			return fmt.Errorf("rcon dial: %w", err)
		}
		defer func() { _ = client.Close() }()
		_, err = client.Execute(action.Command)
		if err != nil {
			return fmt.Errorf("rcon execute: %w", err)
		}
		return nil

	case ActionLifecycle:
		ctx := context.Background()
		switch action.Lifecycle {
		case "start":
			if server.ContainerStatus == "running" {
				return fmt.Errorf("server already running")
			}
			s.actions.Start(ctx, server.ServerID, false)
		case "stop":
			if server.ContainerStatus != "running" {
				return fmt.Errorf("server not running")
			}
			s.actions.Stop(ctx, server.ServerID)
		case "restart":
			s.actions.Restart(ctx, server.ServerID)
		default:
			return fmt.Errorf("unknown lifecycle action: %s", action.Lifecycle)
		}
		return nil

	case ActionContainerExec:
		if server.ContainerID == "" || server.ContainerStatus != "running" {
			return fmt.Errorf("container not running")
		}
		ctx := context.Background()
		cmd := []string{"/bin/sh", "-c", action.Command}
		resp, err := s.cli.ContainerExecCreate(ctx, server.ContainerID, container.ExecOptions{
			Cmd:          cmd,
			AttachStdout: true,
			AttachStderr: true,
		})
		if err != nil {
			return fmt.Errorf("container exec create: %w", err)
		}
		attach, err := s.cli.ContainerExecAttach(ctx, resp.ID, container.ExecStartOptions{})
		if err != nil {
			return fmt.Errorf("container exec attach: %w", err)
		}
		defer attach.Close()
		// Drain output to prevent blocking
		_, _ = io.Copy(io.Discard, attach.Reader)
		return nil

	case ActionBackup:
		if s.backup == nil {
			return fmt.Errorf("backup service not configured")
		}
		_, err := s.backup.Create(server.ServerID, action.KeepLast, action.KeepDays)
		return err

	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}
