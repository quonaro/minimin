package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"orchestrator/internal/db"
	"orchestrator/internal/events"
	"orchestrator/internal/handlers"
	"orchestrator/internal/jwt"
	"orchestrator/internal/logger"
	"orchestrator/internal/routes"
	"orchestrator/internal/status"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	apiBind := getEnv("ORCHESTRATOR_API_BIND", ":8081")
	apiKey := getEnv("ORCHESTRATOR_API_KEY", "")
	jwtSecret := getEnv("JWT_SECRET", "")
	dbPath := getEnv("DB_PATH", "orchestrator.db")
	logLevel := getEnv("ORCHESTRATOR_LOG_LEVEL", "info")
	logFormat := getEnv("ORCHESTRATOR_LOG_FORMAT", "pretty")

	if apiKey == "" {
		slog.Error("ORCHESTRATOR_API_KEY must be set")
		os.Exit(1)
	}

	if jwtSecret == "" {
		slog.Error("JWT_SECRET must be set")
		os.Exit(1)
	}

	log := logger.Init(logLevel, logFormat)
	slog.SetDefault(log)

	slog.Info("Starting WebUI", "bind", apiBind, "db", dbPath, "cwd", ".")

	database, err := db.Open(dbPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer func() { _ = database.Close() }()

	statusStore := status.NewStore()
	statusStore.SetCheckInterval(30 * time.Second)

	broadcaster := events.NewBroadcaster()

	jwtService := jwt.NewService(jwtSecret)
	h := handlers.NewHandler(database, apiKey, jwtService, statusStore, broadcaster)
	router := routes.SetupRoutes(h, apiKey, jwtService, statusStore, broadcaster)

	slog.Info("orchestrator API listening", "addr", apiBind)
	server := &http.Server{Addr: apiBind, Handler: router}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	// Run initial health check and start periodic checker every 30 seconds.
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		statusStore.SetLastCheck(time.Now())
		h.CheckAllAgents()
		for range ticker.C {
			statusStore.SetLastCheck(time.Now())
			h.CheckAllAgents()
		}
	}()

	// Poll server statuses from online agents with adaptive speed.
	go func() {
		lastKnown := &sync.Map{}
		slow := 15 * time.Second
		fast := 2 * time.Second
		ticker := time.NewTicker(slow)
		defer ticker.Stop()

		hasPending := h.PollServerStatuses(lastKnown)
		for range ticker.C {
			hasPending = h.PollServerStatuses(lastKnown)
			if hasPending {
				ticker.Reset(fast)
			} else {
				ticker.Reset(slow)
			}
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
