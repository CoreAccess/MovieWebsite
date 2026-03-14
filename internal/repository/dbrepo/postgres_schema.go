package dbrepo

import (
	"log"
)

func (m *PostgresDBRepo) createTables() {
	query := `
	-- Core Entities
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT,
		google_id TEXT,
		facebook_id TEXT,
		avatar TEXT,
		reputation_score INTEGER DEFAULT 0,
		role TEXT DEFAULT 'user',
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		expires_at TIMESTAMPTZ NOT NULL
	);

	CREATE TABLE IF NOT EXISTS password_reset_tokens (
		token TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		expires_at TIMESTAMPTZ NOT NULL
	);

	CREATE TABLE IF NOT EXISTS genres (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		slug TEXT NOT NULL UNIQUE
	);

	CREATE TABLE IF NOT EXISTS languages (
		code TEXT PRIMARY KEY,
		name TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS countries (
		code TEXT PRIMARY KEY,
		name TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS people (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		slug TEXT NOT NULL UNIQUE,
		gender TEXT,
		birth_date TEXT,
		birth_place TEXT,
		death_date TEXT,
		height TEXT,
		description TEXT,
		image TEXT,
		knows_language TEXT,
		nationality_code TEXT,
		known_for_department TEXT,
		popularity_score REAL DEFAULT 0.0,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS characters (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		slug TEXT NOT NULL,
		gender TEXT,
		birth_date TEXT,
		death_date TEXT,
		description TEXT,
		image TEXT,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	-- Schema.org Supertype Migrations
	CREATE TABLE IF NOT EXISTS media (
		id SERIAL PRIMARY KEY,
		media_type TEXT NOT NULL, -- 'Movie' or 'TVSeries'
		name TEXT NOT NULL,
		slug TEXT NOT NULL,
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
		tmdb_id INTEGER,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (slug, media_type)
	);

	CREATE TABLE IF NOT EXISTS movies (
		media_id INTEGER PRIMARY KEY REFERENCES media(id) ON DELETE CASCADE,
		budget TEXT,
		box_office TEXT,
		duration INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS tv_series (
		media_id INTEGER PRIMARY KEY REFERENCES media(id) ON DELETE CASCADE,
		end_date TEXT,
		number_of_seasons INTEGER DEFAULT 0,
		number_of_episodes INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS tv_episodes (
		id SERIAL PRIMARY KEY,
		series_id INTEGER NOT NULL REFERENCES tv_series(media_id) ON DELETE CASCADE,
		season_number INTEGER NOT NULL,
		episode_number INTEGER NOT NULL,
		name TEXT NOT NULL,
		slug TEXT NOT NULL,
		date_published TEXT,
		description TEXT,
		image TEXT,
		duration INTEGER,
		UNIQUE (series_id, season_number, episode_number)
	);

	CREATE TABLE IF NOT EXISTS tv_seasons (
		id SERIAL PRIMARY KEY,
		series_id INTEGER NOT NULL REFERENCES tv_series(media_id) ON DELETE CASCADE,
		season_number INTEGER NOT NULL,
		name TEXT,
		description TEXT,
		image TEXT,
		date_published TEXT,
		episode_count INTEGER,
		aggregate_rating REAL,
		UNIQUE (series_id, season_number)
	);

	-- Unified Relational Tables pointing to media_id
	CREATE TABLE IF NOT EXISTS media_cast (
		id SERIAL PRIMARY KEY,
		media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
		person_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
		character_name TEXT NOT NULL,
		list_order INTEGER DEFAULT 0,
		UNIQUE(media_id, person_id, character_name)
	);

	CREATE TABLE IF NOT EXISTS media_crew (
		id SERIAL PRIMARY KEY,
		media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
		person_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
		job TEXT NOT NULL,
		department TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS media_genres (
		id SERIAL PRIMARY KEY,
		media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
		genre_id INTEGER NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
		UNIQUE(media_id, genre_id)
	);

	CREATE TABLE IF NOT EXISTS reviews (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
		rating REAL NOT NULL,
		title TEXT,
		body TEXT,
		positive_notes TEXT,
		negative_notes TEXT,
		contains_spoilers BOOLEAN DEFAULT FALSE,
		review_type TEXT DEFAULT 'user',
		publication_name TEXT,
		external_review_url TEXT,
		status TEXT DEFAULT 'published',
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, media_id)
	);

	-- Ecosystem & Gamification Tables
	CREATE TABLE IF NOT EXISTS watchlists (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		description TEXT,
		is_public BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS watchlist_items (
		id SERIAL PRIMARY KEY,
		watchlist_id INTEGER NOT NULL REFERENCES watchlists(id) ON DELETE CASCADE,
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		added_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);


	`

	_, err := m.DB.Exec(query)
	if err != nil {
		log.Fatalf("Error creating tables: %v\n", err)
	}
}
