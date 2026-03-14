package main

import (
	"log"
	"net/http"
	"os"

	"movieweb/internal/config"
	"movieweb/internal/database"

	"movieweb/internal/repository/dbrepo"
	"movieweb/internal/service"
)

func main() {
	// Initialize loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Load environment variables from .env file if it exists
	config.LoadEnv(".env")

	// Initialize the SQLite database and seed it with data from TMDB if empty.
	// We check for TMDB_ACCESS_TOKEN (v4) first, then fallback to TMDB_API_KEY (v3).
	tmdbKey := os.Getenv("TMDB_ACCESS_TOKEN")
	if tmdbKey == "" {
		tmdbKey = os.Getenv("TMDB_API_KEY")
	}

	if tmdbKey == "" {
		tmdbKey = "eyJhbGciOiJIUzI1NiJ9.eyJhdWQiOiJhOWJkZTc1NTdkZTNmNTBiN2FiNzRhODU2MGU0YTc2NCIsIm5iZiI6MTY4ODY3NDU1OC4zOTIsInN1YiI6IjY0YTcyMGZlZjkyNTMyMDE0ZTljNmE4NCIsInNjb3BlcyI6WyJhcGpfcmVhZCJdLCJ2ZXJzaW9uIjoxfQ.8dDf7xLb6lSf1n6TwUgxV3loKu3ieuB0yQw0J4MXCg4"
		infoLog.Println("Note: Using hardcoded TMDB API key. If you see 401 Unauthorized errors, please set the TMDB_ACCESS_TOKEN or TMDB_API_KEY environment variable.")
	}
	
		// LEGACY INIT: Keep the old global DB connection alive for un-migrated handlers
	if _, err := database.InitDB("./streamline.db", tmdbKey); err != nil {
		errorLog.Fatalf("Failed to initialize database: %v\n", err)
	}

	// Create the SQLite repository mapping (Phase 3 N-Tier Layering)
	sqliteRepo := &dbrepo.SqliteDBRepo{DB: database.DB} // Use the same connection for now

	// Create the Service layer to encapsulate business logic and Vector integration readiness
	appService := service.NewAppService(sqliteRepo)


	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
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
