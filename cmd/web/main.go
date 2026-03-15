package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"movieweb/internal/config"
	"movieweb/internal/repository/dbrepo"
	"movieweb/internal/service"
)

func main() {
	// Initialize loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Load environment variables from .env file if it exists
	config.LoadEnv(".env")

	// Initialize the PostgreSQL database and seed it with data from TMDB if empty.
	tmdbKey := os.Getenv("TMDB_ACCESS_TOKEN")
	if tmdbKey == "" {
		tmdbKey = os.Getenv("TMDB_API_KEY")
	}

	if tmdbKey == "" {
		errorLog.Println("TMDB API keys and Access Token not found in environment variables (TMDB_ACCESS_TOKEN or TMDB_API_KEY). Database seeding will be skipped.")
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

	// Create the PostgreSQL repository mapping
	pgRepo := &dbrepo.PostgresDBRepo{}
	db, err := pgRepo.InitDB(dsn, tmdbKey)
	if err != nil {
		errorLog.Fatalf("Failed to initialize PostgreSQL database: %v\n", err)
	}
	defer db.Close()

	// Create the Service layer to encapsulate business logic and Vector integration readiness
	appService := service.NewAppService(pgRepo)

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// Initialize slog logger as per AGENTS.md requirements
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		logger:        logger,
		templateCache: templateCache,
		Service:       appService,
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	infoLog.Printf("Starting web server on port :%s\n", port)

	err = http.ListenAndServe(":"+port, app.routes())
	errorLog.Fatal(err)
}
