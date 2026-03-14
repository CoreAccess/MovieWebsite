//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "./streamline.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, name, slug FROM movies")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("--- Movies List ---")
	for rows.Next() {
		var id int
		var name, slug string
		if err := rows.Scan(&id, &name, &slug); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %d | Name: %s | Slug: %s\n", id, name, slug)
	}

	rows2, err := db.Query("SELECT pos.media_type, pos.media_id, p.name, pos.job_title FROM media_crew pos JOIN people p ON pos.person_id = p.id WHERE pos.media_id = 1")
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	fmt.Println("\n--- Crew for Movie ID 1 ---")
	for rows2.Next() {
		var mediaType, personName, jobTitle string
		var mediaId int
		if err := rows2.Scan(&mediaType, &mediaId, &personName, &jobTitle); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("[%s %d] %s: %s\n", mediaType, mediaId, jobTitle, personName)
	}
}
