package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"filmgap/internal/config"
	"filmgap/internal/repository/dbrepo"
	"filmgap/internal/service"
)

func main() {
	// Initialize slog JSON logger as per AGENTS.md requirements
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Load environment variables from .env file if it exists
	config.LoadEnv(".env")

	// Initialize the PostgreSQL database
	tmdbKey := os.Getenv("TMDB_ACCESS_TOKEN")
	if tmdbKey == "" {
		tmdbKey = os.Getenv("TMDB_API_KEY")
	}

	if tmdbKey == "" {
		logger.Warn("TMDB API key not found — database seeding will be skipped",
			"env_vars", "TMDB_ACCESS_TOKEN or TMDB_API_KEY",
		)
	}

	// PostgreSQL connection parameters
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

	// Create the PostgreSQL repository
	pgRepo := &dbrepo.PostgresDBRepo{}
	db, err := pgRepo.InitDB(dsn, tmdbKey)
	if err != nil {
		logger.Error("failed to initialize PostgreSQL database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create the Service layer — encapsulates business logic and is Vector DB ready
	appService := service.NewAppService(pgRepo, tmdbKey)
	appService.SeedHomepageData()

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error("failed to build template cache", "error", err)
		os.Exit(1)
	}

	app := &application{
		logger:        logger,
		templateCache: templateCache,
		Service:       appService,
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("starting web server", "port", port, "addr", ":"+port)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
