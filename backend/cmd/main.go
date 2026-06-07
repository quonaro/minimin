package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"orchestrator/internal/db"
	"orchestrator/internal/handlers"
	"orchestrator/internal/jwt"
	"orchestrator/internal/routes"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	apiBind := getEnv("ORCHESTRATOR_API_BIND", ":8081")
	apiKey := getEnv("ORCHESTRATOR_API_KEY", "")
	jwtSecret := getEnv("JWT_SECRET", "")
	dbPath := getEnv("DB_PATH", "orchestrator.db")

	if apiKey == "" {
		slog.Error("ORCHESTRATOR_API_KEY must be set")
		os.Exit(1)
	}

	if jwtSecret == "" {
		slog.Error("JWT_SECRET must be set")
		os.Exit(1)
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(log)

	slog.Info("starting orchestrator", "bind", apiBind, "db", dbPath)

	database, err := db.Open(dbPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer func() { _ = database.Close() }()

	jwtService := jwt.NewService(jwtSecret)
	h := handlers.NewHandler(database, apiKey, jwtService)
	router := routes.SetupRoutes(h, apiKey, jwtService)

	slog.Info("orchestrator API listening", "addr", apiBind)
	server := &http.Server{Addr: apiBind, Handler: router}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
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
