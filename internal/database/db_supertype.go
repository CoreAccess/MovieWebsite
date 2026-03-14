package database

import "log"

// ExecuteSchemaOrgMigrations applies the new supertype table structure for Phase 1.
// Note: In a production scenario with live data, we would write complex INSERT INTO ... SELECT
// data migrations. Since this is an architectural refactor, we are altering the initialization
// DDL to define the strict Schema.org-compliant table hierarchy.
func ExecuteSchemaOrgMigrations() {
    queries := []string{
        // Drop legacy specific tables if they exist to force the new schema
        "DROP TABLE IF EXISTS movies;",
        "DROP TABLE IF EXISTS tv_series;",
        "DROP TABLE IF EXISTS movie_cast;",
        "DROP TABLE IF EXISTS tv_cast;",

        // The Supertype Table: Media
        `CREATE TABLE IF NOT EXISTS media (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            media_type TEXT NOT NULL, -- 'Movie' or 'TVSeries'
            name TEXT NOT NULL,
            slug TEXT UNIQUE NOT NULL,
            description TEXT,
            image TEXT,
            date_published TEXT, -- Generic for release_date or first_air_date
            aggregate_rating REAL DEFAULT 0.0,
            tmdb_id INTEGER UNIQUE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`,

        // Subtype Table: Movies (References media.id)
        `CREATE TABLE IF NOT EXISTS movies (
            media_id INTEGER PRIMARY KEY REFERENCES media(id) ON DELETE CASCADE,
            runtime INTEGER DEFAULT 0
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
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
            person_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
            character_name TEXT NOT NULL,
            list_order INTEGER DEFAULT 0,
            UNIQUE(media_id, person_id, character_name)
        );`,

        `CREATE TABLE IF NOT EXISTS media_crew (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
            person_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
            job TEXT NOT NULL,
            department TEXT NOT NULL
        );`,

        `CREATE TABLE IF NOT EXISTS media_genres (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
            genre_id INTEGER NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
            UNIQUE(media_id, genre_id)
        );`,

        `CREATE TABLE IF NOT EXISTS reviews (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
            rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 10),
            content TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(user_id, media_id)
        );`,
    }

    for _, q := range queries {
        if _, err := DB.Exec(q); err != nil {
            log.Printf("Error executing Supertype schema query: %s\nError: %v\n", q, err)
        }
    }
}
// Ensure seed data works with the supertype
