//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"filmgap/internal/config"

	_ "github.com/lib/pq"
)

func main() {
	config.LoadEnv(".env")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbUser, dbPass, dbName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Could not connect to DB:", err)
	}

	query := `
		TRUNCATE TABLE 
			pending_ingestion, 
			media, 
			people, 
			genres, 
			characters 
		CASCADE;
	`
	_, err = db.Exec(query)
	if err != nil {
		log.Fatal("Failed to truncate tables:", err)
	}

	fmt.Println("Database successfully truncated. Ready for a fresh ingestion.")
}
