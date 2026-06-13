package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"orchestrator/internal/events"
	"orchestrator/internal/handlers"
	"orchestrator/internal/health"
	"orchestrator/internal/metrics"
	"orchestrator/internal/routes"
	"orchestrator/internal/runner"
	"orchestrator/internal/state"
	"orchestrator/internal/static"

	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	apiBind := getEnv("ORCHESTRATOR_API_BIND", ":8081")
	apiKey := getEnv("ORCHESTRATOR_API_KEY", "")
	serversDir := getEnv("MC_SERVERS_DIR", "./servers")
	serversHostDir := os.Getenv("MC_SERVERS_HOST_DIR")
	if serversHostDir == "" {
		slog.Error("MC_SERVERS_HOST_DIR must be set")
		os.Exit(1)
	}
	instanceFile := getEnv("MC_INSTANCE_FILE", "./instance.yml")
	modUploadMaxMB := 1024
	if v := os.Getenv("MC_MOD_UPLOAD_MAX_MB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			modUploadMaxMB = n
		}
	}
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

	networkName, err := runner.ValidateDockerEnvironment(ctx, cli)
	if err != nil {
		slog.Error("docker environment validation failed", "error", err)
		os.Exit(1)
	}
	slog.Info("using docker network", "name", networkName)

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

	hub := events.NewHub()

	instance.Broadcast = func(serverID string, s state.ServerState) {
		if s.HostPath == "" && s.VolumePath != "" {
			s.HostPath = runner.HostPathForDocker(s.VolumePath, serversDir, serversHostDir)
		}
		if s.ContainerPath == "" {
			s.ContainerPath = "/data"
		}
		s.ModCount = state.CountMods(s)
		hub.BroadcastJSON("server", s)
	}

	modrinthCustomURL := os.Getenv("MODRINTH_CUSTOM_URL")
	secureCookie := os.Getenv("ORCHESTRATOR_SECURE_COOKIE") == "true"
	wsOrigin := os.Getenv("ORCHESTRATOR_WS_ORIGIN")
	if wsOrigin == "" {
		wsOrigin = "*"
	}
	var allowedOrigins []string
	if wsOrigin != "*" {
		allowedOrigins = strings.Split(wsOrigin, ",")
		for i := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
		}
	}

	h := handlers.NewHandler(cli, instance, apiKey, modrinthCustomURL)
	h.ServersDir = serversDir
	h.ServersHostDir = serversHostDir
	h.ModUploadMaxMB = modUploadMaxMB
	h.EventsHub = hub
	h.NetworkName = networkName
	h.SecureCookie = secureCookie
	h.WSUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if len(allowedOrigins) == 0 || (len(allowedOrigins) == 1 && allowedOrigins[0] == "*") {
				return true
			}
			origin := r.Header.Get("Origin")
			for _, allowed := range allowedOrigins {
				if allowed == origin {
					return true
				}
			}
			return false
		},
	}
	h.InitArchives()
	router := routes.SetupRoutes(h, apiKey)

	// Serve embedded SPA in production; nil in dev where web/ is empty.
	if spa := static.SPAHandler(); spa != nil {
		combined := http.NewServeMux()
		combined.Handle("/api/", router)
		combined.Handle("/ws/", router)
		combined.Handle("/healthz", router)
		combined.Handle("/", spa)
		router = combined
	}

	// Background archive cleanup
	go h.StartArchiveCleanup(ctx)

	// Background metrics poller
	go metrics.NewPoller(cli, instance, hub).Start(ctx)

	// Background health-checker
	go health.NewChecker(instance, h.BroadcastPlayerDataAsync).Start(ctx)

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
