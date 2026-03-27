//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	
	"filmgap/internal/config"
	"filmgap/internal/repository/dbrepo"
)

func main() {
	config.LoadEnv(".env")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" { dbPort = "5432" }

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", 
		dbHost, dbPort, dbUser, dbPass, dbName)

	pgRepo := &dbrepo.PostgresDBRepo{}
	_, err := pgRepo.InitDB(dsn, "")
	if err != nil {
		log.Fatalf("Failed to init: %v", err)
	}

	movies, err := pgRepo.GetPopularMovies(10)
	fmt.Printf("GetPopularMovies: %v, Err: %v\n", len(movies), err)

	shows, err := pgRepo.GetPopularShows(10)
	fmt.Printf("GetPopularShows: %v, Err: %v\n", len(shows), err)

	people, err := pgRepo.GetAllPeople(10, 0, "")
	fmt.Printf("GetAllPeople: %v, Err: %v\n", len(people), err)

	cast, err := pgRepo.GetCastForMedia(3)
	fmt.Printf("GetCastForMedia(3): %v, Err: %v\n", len(cast), err)
}

