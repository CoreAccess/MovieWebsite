package main

import (
	"log"
	"net/http"
	"os"

	"movieweb/internal/database"
)

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	tmdbKey := "eyJhbGciOiJIUzI1NiJ9.eyJhdWQiOiJhOWJkZTc1NTdkZTNmNTBiN2FiNzRhODU2MGU0YTc2NCIsIm5iZiI6MTY4ODY3NDU1OC4zOTIsInN1YiI6IjY0YTcyMGZlZjkyNTMyMDE0ZTljNmE4NCIsInNjb3BlcyI6WyJhcGpfcmVhZCJdLCJ2ZXJzaW9uIjoxfQ.8dDf7xLb6lSf1n6TwUgxV3loKu3ieuB0yQw0J4MXCg4"
	if _, err := database.InitDB("./streamline.db", tmdbKey); err != nil {
		errorLog.Fatalf("Failed to initialize database: %v\n", err)
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		templateCache: templateCache,
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	infoLog.Printf("Starting web server on port :%s\n", port)

	err = http.ListenAndServe(":"+port, app.routes())
	errorLog.Fatal(err)
}
