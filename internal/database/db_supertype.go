package database

import (
	"database/sql"
	"log"
)

// ExecuteSchemaOrgMigrations applies the new supertype table structure for Phase 1.
func ExecuteSchemaOrgMigrations(db *sql.DB) {
    queries := []string{
        // Drop legacy specific tables if they exist to force the new schema
        "DROP TABLE IF EXISTS reviews CASCADE;",
        "DROP TABLE IF EXISTS media_genres CASCADE;",
        "DROP TABLE IF EXISTS media_crew CASCADE;",
        "DROP TABLE IF EXISTS media_cast CASCADE;",
        "DROP TABLE IF EXISTS tv_episodes CASCADE;",
        "DROP TABLE IF EXISTS tv_series CASCADE;",
        "DROP TABLE IF EXISTS movies CASCADE;",
        "DROP TABLE IF EXISTS person_aliases CASCADE;",
        "DROP TABLE IF EXISTS characters CASCADE;",
        "DROP TABLE IF EXISTS people CASCADE;",
        "DROP TABLE IF EXISTS languages CASCADE;",
        "DROP TABLE IF EXISTS countries CASCADE;",
        "DROP TABLE IF EXISTS genres CASCADE;",
        "DROP TABLE IF EXISTS keywords CASCADE;",
        "DROP TABLE IF EXISTS media CASCADE;",

        // The Supertype Table: Media
        `CREATE TABLE IF NOT EXISTS media (
            id SERIAL PRIMARY KEY,
            media_type TEXT NOT NULL, -- 'Movie' or 'TVSeries'
            name TEXT NOT NULL,
            slug TEXT UNIQUE NOT NULL,
            description TEXT,
            image TEXT,
            date_published TEXT, -- Generic for release_date or first_air_date
            content_rating TEXT,
            aggregate_rating REAL DEFAULT 0.0,
            rating_count INTEGER DEFAULT 0,
            review_count INTEGER DEFAULT 0,
            best_rating REAL DEFAULT 10.0,
            worst_rating REAL DEFAULT 1.0,
            is_family_friendly BOOLEAN DEFAULT TRUE,
            tagline TEXT,
            subtitle TEXT,
            language_code TEXT,
            country_code TEXT,
            tmdb_id INTEGER UNIQUE,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
        );`,

        // Subtype Table: Movies (References media.id)
        `CREATE TABLE IF NOT EXISTS movies (
            media_id INTEGER PRIMARY KEY REFERENCES media(id) ON DELETE CASCADE,
            budget TEXT,
            box_office TEXT,
            duration INTEGER DEFAULT 0
        );`,

        // Subtype Table: TV Series (References media.id)
        `CREATE TABLE IF NOT EXISTS tv_series (
            media_id INTEGER PRIMARY KEY REFERENCES media(id) ON DELETE CASCADE,
            end_date TEXT,
            number_of_seasons INTEGER DEFAULT 0,
            number_of_episodes INTEGER DEFAULT 0
        );`,

        // Unified Relational Tables pointing to media_id
        `CREATE TABLE IF NOT EXISTS media_cast (
            id SERIAL PRIMARY KEY,
            media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
            person_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
            character_name TEXT NOT NULL,
            list_order INTEGER DEFAULT 0,
            UNIQUE(media_id, person_id, character_name)
        );`,

        `CREATE TABLE IF NOT EXISTS media_crew (
            id SERIAL PRIMARY KEY,
            media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
            person_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
            job TEXT NOT NULL,
            department TEXT NOT NULL
        );`,

        `CREATE TABLE IF NOT EXISTS media_genres (
            id SERIAL PRIMARY KEY,
            media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
            genre_id INTEGER NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
            UNIQUE(media_id, genre_id)
        );`,

        `CREATE TABLE IF NOT EXISTS reviews (
            id SERIAL PRIMARY KEY,
            user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
            rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 10),
            content TEXT,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(user_id, media_id)
        );`,
    }

    for _, q := range queries {
        if _, err := db.Exec(q); err != nil {
            log.Printf("Error executing Supertype schema query: %s\nError: %v\n", q, err)
        }
    }
}
// Ensure seed data works with the supertype
