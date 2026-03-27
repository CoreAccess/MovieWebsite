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
		avatar TEXT DEFAULT '/static/img/default_avatar.webp',
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
		UNIQUE (slug, media_type),
		UNIQUE (tmdb_id)
	);

	CREATE TABLE IF NOT EXISTS movies (
		media_id INTEGER PRIMARY KEY REFERENCES media(id) ON DELETE CASCADE,
		budget TEXT,
		box_office TEXT,
		duration INTEGER DEFAULT 0,
		tagline TEXT
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

	CREATE TABLE IF NOT EXISTS watch_providers (
		id SERIAL PRIMARY KEY,
		media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
		country_code TEXT NOT NULL,
		provider_type TEXT NOT NULL,
		provider_id INTEGER NOT NULL,
		provider_name TEXT NOT NULL,
		logo_url TEXT,
		display_priority INTEGER DEFAULT 0,
		deep_link_url TEXT,
		source TEXT DEFAULT 'tmdb',
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(media_id, country_code, provider_type, provider_id)
	);
	CREATE INDEX IF NOT EXISTS idx_watch_providers_media_country ON watch_providers(media_id, country_code);

	-- Ecosystem & Gamification Tables
	CREATE TABLE IF NOT EXISTS hero_features (
		id SERIAL PRIMARY KEY,
		media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
		selected_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS pending_ingestion (
		id SERIAL PRIMARY KEY,
		tmdb_id INTEGER NOT NULL,
		media_type VARCHAR(10) NOT NULL CHECK (media_type IN ('Movie', 'TV')),
		status VARCHAR(20) NOT NULL DEFAULT 'QUEUED' CHECK (status IN ('QUEUED', 'PROCESSING', 'COMPLETED', 'FAILED')),
		attempts INTEGER NOT NULL DEFAULT 0,
		last_attempt TIMESTAMPTZ,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(tmdb_id, media_type)
	);

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
		added_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(watchlist_id, media_id)
	);

	CREATE TABLE IF NOT EXISTS lists (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		slug TEXT NOT NULL,
		description TEXT,
		is_ranked BOOLEAN DEFAULT FALSE,
		is_collaborative BOOLEAN DEFAULT FALSE,
		visibility TEXT DEFAULT 'public',
		like_count INTEGER DEFAULT 0,
		follower_count INTEGER DEFAULT 0,
		item_count INTEGER DEFAULT 0,
		is_featured BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, slug)
	);
	CREATE INDEX IF NOT EXISTS idx_lists_user ON lists(user_id);
	CREATE INDEX IF NOT EXISTS idx_lists_public ON lists(created_at DESC) WHERE visibility = 'public';

	CREATE TABLE IF NOT EXISTS list_items (
		id SERIAL PRIMARY KEY,
		list_id INTEGER NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
		media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
		rank INTEGER,
		note TEXT,
		added_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		added_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(list_id, media_id)
	);
	CREATE INDEX IF NOT EXISTS idx_list_items_list ON list_items(list_id);

	CREATE TABLE IF NOT EXISTS list_collaborators (
		id SERIAL PRIMARY KEY,
		list_id INTEGER NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		role TEXT DEFAULT 'contributor',
		joined_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(list_id, user_id)
	);

	-- New Homepage Expansion Tables
	CREATE TABLE IF NOT EXISTS blog_posts (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		slug TEXT NOT NULL UNIQUE,
		content TEXT NOT NULL,
		image TEXT,
		author_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
		is_featured BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS activities (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		activity_type TEXT NOT NULL,
		target_id INTEGER,
		target_type TEXT,
		metadata JSONB,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_activities_user ON activities(user_id);
	CREATE INDEX IF NOT EXISTS idx_activities_created ON activities(created_at DESC);

	CREATE TABLE IF NOT EXISTS franchises (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		slug TEXT NOT NULL UNIQUE,
		description TEXT,
		image TEXT,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS media_franchises (
		media_id INTEGER REFERENCES media(id) ON DELETE CASCADE,
		franchise_id INTEGER REFERENCES franchises(id) ON DELETE CASCADE,
		list_order INTEGER DEFAULT 0,
		PRIMARY KEY (media_id, franchise_id)
	);

	CREATE TABLE IF NOT EXISTS photos (
		id SERIAL PRIMARY KEY,
		media_id INTEGER REFERENCES media(id) ON DELETE CASCADE,
		person_id INTEGER REFERENCES people(id) ON DELETE CASCADE,
		image_url TEXT NOT NULL,
		caption TEXT,
		is_popular BOOLEAN DEFAULT FALSE,
		view_count INTEGER DEFAULT 0,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS user_follows (
		id SERIAL PRIMARY KEY,
		follower_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		followed_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		followed_person_id INTEGER REFERENCES people(id) ON DELETE CASCADE,
		followed_list_id INTEGER REFERENCES lists(id) ON DELETE CASCADE,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		-- Ensure a follow is unique across its target types
		CONSTRAINT unique_user_follow UNIQUE (follower_id, followed_user_id),
		CONSTRAINT unique_person_follow UNIQUE (follower_id, followed_person_id),
		CONSTRAINT unique_list_follow UNIQUE (follower_id, followed_list_id),
		-- Ensure at least one target is specified
		CONSTRAINT at_least_one_target CHECK (
			(followed_user_id IS NOT NULL)::INT + 
			(followed_person_id IS NOT NULL)::INT + 
			(followed_list_id IS NOT NULL)::INT = 1
		)
	);
	CREATE INDEX IF NOT EXISTS idx_user_follows_follower ON user_follows(follower_id);
	`

	_, err := m.DB.Exec(query)
	if err != nil {
		log.Fatalf("Error creating tables: %v\n", err)
	}
}
