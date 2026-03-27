package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"fmt"

	"filmgap/internal/config"
	"filmgap/internal/ingestor"
	"filmgap/internal/repository/dbrepo"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load .env if present (for local development).
	config.LoadEnv(".env")

	tmdbKey := os.Getenv("TMDB_ACCESS_TOKEN")
	if tmdbKey == "" {
		tmdbKey = os.Getenv("TMDB_API_KEY")
	}
	if tmdbKey == "" {
		logger.Error("TMDB_ACCESS_TOKEN or TMDB_API_KEY must be set")
		os.Exit(1)
	}

	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	pgRepo := &dbrepo.PostgresDBRepo{}
	db, err := pgRepo.InitDB(dsn, tmdbKey)
	if err != nil {
		logger.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("ingestor connected to database")

	// Graceful shutdown via SIGINT / SIGTERM (Ctrl+C or GCE VM termination).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := ingestor.DefaultConfig()
	ing := ingestor.New(pgRepo, tmdbKey, cfg, logger)
	ing.Run(ctx)
}
