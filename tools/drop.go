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
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" { dbPort = "5432" }

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
	db, err := sql.Open("postgres", dsn)
	if err != nil { log.Fatal(err) }
	defer db.Close()

	_, err = db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if err != nil {
		log.Println("Error:", err)
	} else {
		log.Println("Schema wiped successfully.")
	}
}

