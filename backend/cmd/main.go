package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"orchestrator/internal/handlers"
	"orchestrator/internal/runner"
	"orchestrator/internal/routes"
	"orchestrator/internal/state"

	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	apiBind := getEnv("ORCHESTRATOR_API_BIND", ":8081")
	apiKey := getEnv("ORCHESTRATOR_API_KEY", "")
	serversDir := getEnv("MC_SERVERS_DIR", "./servers")
	instanceFile := getEnv("MC_INSTANCE_FILE", "./instance.yml")
	logLevel := getEnv("ORCHESTRATOR_LOG_LEVEL", "info")

	if apiKey == "" {
		slog.Error("ORCHESTRATOR_API_KEY must be set")
		os.Exit(1)
	}

	level, err := parseLogLevel(logLevel)
	if err != nil {
		slog.Warn("invalid log level, using info", "error", err)
		level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	instance, err := state.Load(instanceFile)
	if err != nil {
		slog.Error("failed to load state", "error", err)
		os.Exit(1)
	}

	// Ensure an API key exists in instance state
	if instance.APIKey == "" {
		instance.APIKey = apiKey
		if saveErr := instance.Save(); saveErr != nil {
			slog.Error("failed to save initial state", "error", saveErr)
			os.Exit(1)
		}
	}

	slog.Info("initializing docker client")
	cli, err := runner.GetClient()
	if err != nil {
		slog.Error("initialization failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = cli.Close() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Scan existing containers and import into instance.yml
	slog.Info("scanning for managed containers")
	scanned, scanErr := runner.ScanManagedContainers(ctx, cli, serversDir)
	if scanErr != nil {
		slog.Warn("scan warning", "error", scanErr)
	} else {
		imported := 0
		for _, s := range scanned {
			existing, ok := instance.Get(s.ServerID)
			if !ok {
				instance.Set(s)
				imported++
				slog.Info("imported external container",
					"server_id", s.ServerID,
					"container", s.ContainerID[:12],
					"volume", s.VolumeID)
			} else if existing.ContainerID != s.ContainerID || existing.Status != s.Status {
				existing.ContainerID = s.ContainerID
				existing.ContainerStatus = s.ContainerStatus
				existing.ServerStatus = s.ServerStatus
				existing.ContainerStartedAt = s.ContainerStartedAt
				existing.ServerStartedAt = s.ServerStartedAt
				if s.VolumePath != "" {
					existing.VolumePath = s.VolumePath
					existing.VolumeID = s.VolumeID
				}
				instance.Set(existing)
				imported++
				slog.Info("updated container record",
					"server_id", s.ServerID,
					"container", s.ContainerID[:12],
					"status", s.Status)
			}
		}
		if imported == 0 {
			slog.Info("no new or changed containers found")
		}
		if err := instance.Save(); err != nil {
			slog.Warn("failed to save state after scan", "error", err)
		}
	}

	// Reconcile stale container IDs
	for _, existing := range instance.All() {
		if existing.ContainerID == "" {
			continue
		}
		_, inspectErr := cli.ContainerInspect(ctx, existing.ContainerID)
		if client.IsErrNotFound(inspectErr) {
			existing.ContainerID = ""
			existing.ContainerStatus = "exited"
			existing.ServerStatus = "stopped"
			existing.ServerStartedAt = time.Time{}
			instance.Set(existing)
			slog.Info("cleared stale container record", "server_id", existing.ServerID)
		}
	}
	if err := instance.Save(); err != nil {
		slog.Warn("failed to save state after reconciliation", "error", err)
	}

	h := handlers.NewHandler(cli, instance, apiKey)
	h.ServersDir = serversDir
	router := routes.SetupRoutes(h, apiKey)

	// Background health-checker
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for _, s := range instance.All() {
					if s.ContainerStatus != "running" {
						continue
					}
					ok, _ := runner.PingServer("127.0.0.1", s.GamePort, 5*time.Second)
					if ok && s.ServerStatus != "running" {
						s.ServerStatus = "running"
						s.ServerStartedAt = time.Now().UTC()
						instance.Set(s)
						_ = instance.Save()
						slog.Info("server ready", "server_id", s.ServerID, "status", s.ServerStatus)
					} else if !ok && s.ServerStatus == "running" {
						s.ServerStatus = "starting"
						s.ServerStartedAt = time.Time{}
						instance.Set(s)
						_ = instance.Save()
						slog.Info("server not ready", "server_id", s.ServerID, "status", s.ServerStatus)
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	slog.Info("orchestrator API listening", "addr", apiBind)
	server := &http.Server{Addr: apiBind, Handler: router}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseLogLevel(s string) (slog.Level, error) {
	switch s {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	}
	return slog.LevelInfo, fmt.Errorf("unknown log level: %s", s)
}
