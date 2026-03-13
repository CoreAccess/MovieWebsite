package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"movieweb/internal/models"
	"movieweb/internal/tmdb"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// DB holds the global database connection pool.
var DB *sql.DB

// InitDB sets up the database connection, creates tables if they don't exist,
// and populates it with seed data from TMDB if the database is initially empty.
// dataSourceName specifies the path to the SQLite file (e.g., "./streamline.db").
func InitDB(dataSourceName string, tmdbAPIKey string) (*sql.DB, error) {
	var err error
	// sql.Open initializes a connection pool for a specific driver ("sqlite" here).
	// It doesn't actually connect to the database yet.
	DB, err = sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Ping actually establishes a connection and verifies that the database is reachable.
	if err = DB.Ping(); err != nil {
		return nil, err
	}

	log.Println("Connected to SQLite Database")
	// Run the schema creation query to ensure all tables exist before proceeding.
	createTables()
	// Populate initial movie/show/actor data if the `movies` table has 0 rows.
	seedDataIfEmpty(tmdbAPIKey)
	return DB, nil
}

func createTables() {
	query := `
	-- Core Entities
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT,
		google_id TEXT,
		facebook_id TEXT,
		avatar TEXT,
		reputation_score INTEGER DEFAULT 0,
		role TEXT DEFAULT 'user',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS password_reset_tokens (
		token TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS organizations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		description TEXT,
		image TEXT,
		logo TEXT,
		url TEXT,
		founding_date TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS people (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		slug TEXT NOT NULL,
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
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(nationality_code) REFERENCES countries(code)
	);

	CREATE TABLE IF NOT EXISTS person_aliases (
		person_id INTEGER NOT NULL,
		alias TEXT NOT NULL,
		FOREIGN KEY(person_id) REFERENCES people(id)
	);

	CREATE TABLE IF NOT EXISTS characters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		slug TEXT NOT NULL,
		gender TEXT,
		birth_date TEXT,
		death_date TEXT,
		description TEXT,
		image TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS genres (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		slug TEXT NOT NULL UNIQUE
	);

	CREATE TABLE IF NOT EXISTS keywords (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
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

	CREATE TABLE IF NOT EXISTS external_ids (
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		source TEXT NOT NULL,
		external_id TEXT NOT NULL,
		UNIQUE (media_type, media_id, source)
	);

	CREATE TABLE IF NOT EXISTS movies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		slug TEXT NOT NULL,
		date_published TEXT,
		description TEXT,
		image TEXT,
		trailer TEXT,
		video TEXT,
		content_rating TEXT,
		duration INTEGER,
		aggregate_rating REAL,
		budget TEXT,
		box_office TEXT,
		language_code TEXT,
		country_code TEXT,
		tagline TEXT,
		rating_count INTEGER DEFAULT 0,
		review_count INTEGER DEFAULT 0,
		best_rating REAL DEFAULT 10.0,
		worst_rating REAL DEFAULT 1.0,
		is_family_friendly BOOLEAN DEFAULT 0,
		subtitle TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(name, date_published),
		FOREIGN KEY(language_code) REFERENCES languages(code),
		FOREIGN KEY(country_code) REFERENCES countries(code)
	);

	CREATE TABLE IF NOT EXISTS movie_genres (
		movie_id INTEGER NOT NULL,
		genre_id INTEGER NOT NULL,
		PRIMARY KEY(movie_id, genre_id),
		FOREIGN KEY(movie_id) REFERENCES movies(id),
		FOREIGN KEY(genre_id) REFERENCES genres(id)
	);

	CREATE TABLE IF NOT EXISTS movie_keywords (
		movie_id INTEGER NOT NULL,
		keyword_id INTEGER NOT NULL,
		PRIMARY KEY(movie_id, keyword_id),
		FOREIGN KEY(movie_id) REFERENCES movies(id),
		FOREIGN KEY(keyword_id) REFERENCES keywords(id)
	);

	CREATE TABLE IF NOT EXISTS tv_series (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		slug TEXT NOT NULL,
		start_date TEXT,
		end_date TEXT,
		description TEXT,
		image TEXT,
		content_rating TEXT,
		aggregate_rating REAL,
		number_of_seasons INTEGER,
		number_of_episodes INTEGER,
		trailer TEXT,
		language_code TEXT,
		country_code TEXT,
		tagline TEXT,
		rating_count INTEGER DEFAULT 0,
		review_count INTEGER DEFAULT 0,
		best_rating REAL DEFAULT 10.0,
		worst_rating REAL DEFAULT 1.0,
		subtitle TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(name, start_date),
		FOREIGN KEY(language_code) REFERENCES languages(code),
		FOREIGN KEY(country_code) REFERENCES countries(code)
	);

	CREATE TABLE IF NOT EXISTS tv_genres (
		series_id INTEGER NOT NULL,
		genre_id INTEGER NOT NULL,
		PRIMARY KEY(series_id, genre_id),
		FOREIGN KEY(series_id) REFERENCES tv_series(id),
		FOREIGN KEY(genre_id) REFERENCES genres(id)
	);

	CREATE TABLE IF NOT EXISTS tv_keywords (
		series_id INTEGER NOT NULL,
		keyword_id INTEGER NOT NULL,
		PRIMARY KEY(series_id, keyword_id),
		FOREIGN KEY(series_id) REFERENCES tv_series(id),
		FOREIGN KEY(keyword_id) REFERENCES keywords(id)
	);

	CREATE TABLE IF NOT EXISTS tv_episodes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		series_id INTEGER NOT NULL,
		season_number INTEGER NOT NULL,
		episode_number INTEGER NOT NULL,
		name TEXT NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		date_published TEXT,
		description TEXT,
		image TEXT,
		duration INTEGER,
		FOREIGN KEY(series_id) REFERENCES tv_series(id),
		UNIQUE(series_id, season_number, episode_number)
	);

	-- Bridging / Relational Tables
	CREATE TABLE IF NOT EXISTS movie_cast (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		movie_id INTEGER NOT NULL,
		person_id INTEGER NOT NULL,
		character_id INTEGER NOT NULL,
		billing_order INTEGER,
		FOREIGN KEY(movie_id) REFERENCES movies(id),
		FOREIGN KEY(person_id) REFERENCES people(id),
		FOREIGN KEY(character_id) REFERENCES characters(id),
		UNIQUE(movie_id, person_id, character_id)
	);

	CREATE TABLE IF NOT EXISTS tv_cast (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		series_id INTEGER NOT NULL,
		person_id INTEGER NOT NULL,
		character_id INTEGER NOT NULL,
		billing_order INTEGER,
		FOREIGN KEY(series_id) REFERENCES tv_series(id),
		FOREIGN KEY(person_id) REFERENCES people(id),
		FOREIGN KEY(character_id) REFERENCES characters(id),
		UNIQUE(series_id, person_id, character_id)
	);

	CREATE TABLE IF NOT EXISTS media_crew (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_type TEXT NOT NULL, -- 'movie', 'tv_series', 'tv_episode'
		media_id INTEGER NOT NULL,
		person_id INTEGER NOT NULL,
		job_title TEXT NOT NULL,
		FOREIGN KEY(person_id) REFERENCES people(id)
	);

	CREATE TABLE IF NOT EXISTS production_companies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		organization_id INTEGER NOT NULL,
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		role TEXT,
		FOREIGN KEY(organization_id) REFERENCES organizations(id)
	);

	CREATE TABLE IF NOT EXISTS person_relationships (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		person_id INTEGER NOT NULL,
		related_person_id INTEGER NOT NULL,
		relationship_type TEXT NOT NULL, -- 'parent', 'child', 'sibling', 'spouse'
		start_date TEXT,
		end_date TEXT,
		FOREIGN KEY(person_id) REFERENCES people(id),
		FOREIGN KEY(related_person_id) REFERENCES people(id)
	);

	-- Ecosystem & Gamification Tables
	CREATE TABLE IF NOT EXISTS media_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		event_name TEXT NOT NULL,
		in_universe_date TEXT,
		description TEXT
	);

	CREATE TABLE IF NOT EXISTS media_trivia (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target_type TEXT NOT NULL,
		target_id INTEGER NOT NULL,
		trivia_type TEXT NOT NULL, -- 'easter_egg', 'goof', 'quote'
		content TEXT NOT NULL,
		submitted_by INTEGER NOT NULL,
		status TEXT DEFAULT 'pending',
		FOREIGN KEY(submitted_by) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS watchlists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		is_public BOOLEAN DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS watchlist_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		watchlist_id INTEGER NOT NULL,
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(watchlist_id) REFERENCES watchlists(id)
	);

	CREATE TABLE IF NOT EXISTS achievements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		points INTEGER DEFAULT 0,
		badge_image TEXT
	);

	CREATE TABLE IF NOT EXISTS user_achievements (
		user_id INTEGER NOT NULL,
		achievement_id INTEGER NOT NULL,
		earned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY(user_id, achievement_id),
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(achievement_id) REFERENCES achievements(id)
	);

	CREATE TABLE IF NOT EXISTS user_notification_settings (
		user_id INTEGER PRIMARY KEY,
		email_alerts BOOLEAN DEFAULT 1,
		site_alerts BOOLEAN DEFAULT 1,
		mentions BOOLEAN DEFAULT 1,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS edit_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		entity_type TEXT NOT NULL,
		entity_id INTEGER NOT NULL,
		field TEXT NOT NULL,
		old_value TEXT,
		new_value TEXT,
		approved BOOLEAN DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS edit_suggestions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		entity_type TEXT NOT NULL,
		entity_id INTEGER NOT NULL,
		suggested_data TEXT NOT NULL,
		status TEXT DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS ad_campaigns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		company_id INTEGER NOT NULL,
		budget REAL DEFAULT 0.0,
		impressions INTEGER DEFAULT 0,
		clicks INTEGER DEFAULT 0,
		start_date TIMESTAMP,
		end_date TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS campaign_targets (
		campaign_id INTEGER NOT NULL,
		page_slug TEXT NOT NULL,
		PRIMARY KEY(campaign_id, page_slug),
		FOREIGN KEY(campaign_id) REFERENCES ad_campaigns(id)
	);

	CREATE TABLE IF NOT EXISTS advertisements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		campaign_id INTEGER NOT NULL,
		image TEXT,
		url TEXT,
		title TEXT,
		description TEXT,
		FOREIGN KEY(campaign_id) REFERENCES ad_campaigns(id)
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		media_type TEXT,
		media_id INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(post_id) REFERENCES posts(id),
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS likes (
		user_id INTEGER NOT NULL,
		post_id INTEGER NOT NULL,
		PRIMARY KEY(user_id, post_id),
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(post_id) REFERENCES posts(id)
	);

	CREATE TABLE IF NOT EXISTS polls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		question TEXT NOT NULL,
		FOREIGN KEY(post_id) REFERENCES posts(id)
	);

	CREATE TABLE IF NOT EXISTS poll_options (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		poll_id INTEGER NOT NULL,
		option_text TEXT NOT NULL,
		FOREIGN KEY(poll_id) REFERENCES polls(id)
	);

	CREATE TABLE IF NOT EXISTS poll_votes (
		user_id INTEGER NOT NULL,
		option_id INTEGER NOT NULL,
		voted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY(user_id, option_id),
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(option_id) REFERENCES poll_options(id)
	);

	-- Tier 2 - Structural Improvements (New Tables)
	CREATE TABLE IF NOT EXISTS reviews (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		rating REAL NOT NULL,
		title TEXT,
		body TEXT,
		positive_notes TEXT,
		negative_notes TEXT,
		contains_spoilers BOOLEAN DEFAULT 0,
		review_type TEXT DEFAULT 'user',
		publication_name TEXT,
		external_review_url TEXT,
		status TEXT DEFAULT 'published',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (user_id, media_type, media_id),
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS rating_demographics (
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		age_group TEXT NOT NULL,
		avg_rating REAL,
		vote_count INTEGER,
		UNIQUE (media_type, media_id, age_group)
	);

	CREATE TABLE IF NOT EXISTS media_images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		image_type TEXT NOT NULL,
		url TEXT NOT NULL,
		is_primary BOOLEAN DEFAULT 0,
		source TEXT,
		language_code TEXT
	);

	CREATE TABLE IF NOT EXISTS source_material (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		source_type TEXT NOT NULL,
		author_id INTEGER,
		year INTEGER,
		FOREIGN KEY(author_id) REFERENCES people(id)
	);

	CREATE TABLE IF NOT EXISTS movie_source_material (
		movie_id INTEGER NOT NULL,
		source_id INTEGER NOT NULL,
		PRIMARY KEY(movie_id, source_id),
		FOREIGN KEY(movie_id) REFERENCES movies(id),
		FOREIGN KEY(source_id) REFERENCES source_material(id)
	);

	CREATE TABLE IF NOT EXISTS tv_source_material (
		series_id INTEGER NOT NULL,
		source_id INTEGER NOT NULL,
		PRIMARY KEY(series_id, source_id),
		FOREIGN KEY(series_id) REFERENCES tv_series(id),
		FOREIGN KEY(source_id) REFERENCES source_material(id)
	);

	CREATE TABLE IF NOT EXISTS award_bodies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		slug TEXT NOT NULL UNIQUE,
		website_url TEXT
	);

	CREATE TABLE IF NOT EXISTS award_ceremonies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		body_id INTEGER NOT NULL,
		year INTEGER,
		ceremony_number INTEGER,
		date_held DATE,
		FOREIGN KEY(body_id) REFERENCES award_bodies(id)
	);

	CREATE TABLE IF NOT EXISTS award_categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ceremony_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		department TEXT,
		FOREIGN KEY(ceremony_id) REFERENCES award_ceremonies(id)
	);

	CREATE TABLE IF NOT EXISTS award_nominations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category_id INTEGER NOT NULL,
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		person_id INTEGER,
		won BOOLEAN DEFAULT 0,
		nominee_note TEXT,
		FOREIGN KEY(category_id) REFERENCES award_categories(id),
		FOREIGN KEY(person_id) REFERENCES people(id)
	);

	-- Tier 3 - New Entities
	CREATE TABLE IF NOT EXISTS tv_seasons (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		series_id INTEGER NOT NULL,
		season_number INTEGER NOT NULL,
		name TEXT,
		description TEXT,
		image TEXT,
		date_published TEXT,
		episode_count INTEGER,
		aggregate_rating REAL,
		UNIQUE (series_id, season_number),
		FOREIGN KEY(series_id) REFERENCES tv_series(id)
	);

	CREATE TABLE IF NOT EXISTS movie_series (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		description TEXT,
		image TEXT
	);

	CREATE TABLE IF NOT EXISTS movie_series_entries (
		series_id INTEGER NOT NULL,
		movie_id INTEGER NOT NULL,
		position INTEGER,
		PRIMARY KEY(series_id, movie_id),
		FOREIGN KEY(series_id) REFERENCES movie_series(id),
		FOREIGN KEY(movie_id) REFERENCES movies(id)
	);

	CREATE TABLE IF NOT EXISTS quotations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		person_id INTEGER,
		character_id INTEGER,
		quote_text TEXT NOT NULL,
		scene_context TEXT,
		submitted_by INTEGER NOT NULL,
		status TEXT DEFAULT 'published',
		FOREIGN KEY(person_id) REFERENCES people(id),
		FOREIGN KEY(character_id) REFERENCES characters(id),
		FOREIGN KEY(submitted_by) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS screening_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		event_type TEXT,
		event_name TEXT,
		location TEXT,
		event_date DATE,
		description TEXT
	);

	CREATE TABLE IF NOT EXISTS networks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		network_type TEXT,
		country_code TEXT,
		logo_url TEXT,
		website_url TEXT,
		FOREIGN KEY(country_code) REFERENCES countries(code)
	);

	CREATE TABLE IF NOT EXISTS tv_networks (
		series_id INTEGER NOT NULL,
		network_id INTEGER NOT NULL,
		PRIMARY KEY(series_id, network_id),
		FOREIGN KEY(series_id) REFERENCES tv_series(id),
		FOREIGN KEY(network_id) REFERENCES networks(id)
	);

	-- Tier 4 - Competitive Features
	CREATE TABLE IF NOT EXISTS streaming_providers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		logo_url TEXT,
		website_url TEXT,
		affiliate_url TEXT,
		provider_type TEXT,
		has_affiliate_program BOOLEAN DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS media_availability (
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		provider_id INTEGER NOT NULL,
		country_code TEXT NOT NULL,
		availability_type TEXT,
		available_from DATE,
		available_until DATE,
		UNIQUE (media_type, media_id, provider_id, country_code),
		FOREIGN KEY(provider_id) REFERENCES streaming_providers(id),
		FOREIGN KEY(country_code) REFERENCES countries(code)
	);

	CREATE TABLE IF NOT EXISTS release_dates (
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		country_code TEXT NOT NULL,
		release_date DATE NOT NULL,
		release_type TEXT,
		certification TEXT,
		notes TEXT,
		UNIQUE (media_type, media_id, country_code, release_type),
		FOREIGN KEY(country_code) REFERENCES countries(code)
	);

	CREATE TABLE IF NOT EXISTS filming_locations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		location_name TEXT NOT NULL,
		country_code TEXT,
		latitude REAL,
		longitude REAL,
		description TEXT,
		is_real_world BOOLEAN DEFAULT 1,
		FOREIGN KEY(country_code) REFERENCES countries(code)
	);

	CREATE TABLE IF NOT EXISTS technical_specs (
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		color_type TEXT,
		aspect_ratio TEXT,
		sound_mix TEXT,
		negative_format TEXT,
		camera TEXT,
		runtime_minutes INTEGER,
		UNIQUE (media_type, media_id)
	);

	CREATE TABLE IF NOT EXISTS content_advisory (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		category TEXT,
		severity_level TEXT,
		notes TEXT,
		submitted_by INTEGER NOT NULL,
		status TEXT DEFAULT 'published',
		FOREIGN KEY(submitted_by) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS social_links (
		entity_type TEXT NOT NULL,
		entity_id INTEGER NOT NULL,
		platform TEXT NOT NULL,
		url TEXT NOT NULL,
		username TEXT,
		UNIQUE (entity_type, entity_id, platform)
	);

	CREATE TABLE IF NOT EXISTS popularity_snapshots (
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		snapshot_date DATE NOT NULL,
		popularity_score REAL,
		rank_position INTEGER,
		UNIQUE (media_type, media_id, snapshot_date)
	);

	CREATE TABLE IF NOT EXISTS episode_cast (
		episode_id INTEGER NOT NULL,
		person_id INTEGER NOT NULL,
		character_id INTEGER,
		billing_order INTEGER,
		credit_type TEXT DEFAULT 'regular',
		PRIMARY KEY (episode_id, person_id, character_id),
		FOREIGN KEY(episode_id) REFERENCES tv_episodes(id),
		FOREIGN KEY(person_id) REFERENCES people(id),
		FOREIGN KEY(character_id) REFERENCES characters(id)
	);

	CREATE TABLE IF NOT EXISTS user_lists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		is_ranked BOOLEAN DEFAULT 0,
		is_public BOOLEAN DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS user_list_items (
		list_id INTEGER NOT NULL,
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		position INTEGER,
		note TEXT,
		added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (list_id, media_type, media_id),
		FOREIGN KEY(list_id) REFERENCES user_lists(id)
	);

	CREATE TABLE IF NOT EXISTS watch_history (
		user_id INTEGER NOT NULL,
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		watched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		rewatch_count INTEGER DEFAULT 0,
		quick_rating REAL,
		UNIQUE (user_id, media_type, media_id),
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS episode_watch_history (
		user_id INTEGER NOT NULL,
		episode_id INTEGER NOT NULL,
		watched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (user_id, episode_id),
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(episode_id) REFERENCES tv_episodes(id)
	);

	CREATE TABLE IF NOT EXISTS user_follows (
		follower_id INTEGER NOT NULL,
		followed_id INTEGER NOT NULL,
		followed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (follower_id, followed_id),
		FOREIGN KEY(follower_id) REFERENCES users(id),
		FOREIGN KEY(followed_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS moods (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		emoji TEXT,
		description TEXT
	);

	CREATE TABLE IF NOT EXISTS media_moods (
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		mood_id INTEGER NOT NULL,
		vote_count INTEGER DEFAULT 0,
		PRIMARY KEY (media_type, media_id, mood_id),
		FOREIGN KEY(mood_id) REFERENCES moods(id)
	);

	CREATE TABLE IF NOT EXISTS user_mood_votes (
		user_id INTEGER NOT NULL,
		media_type TEXT NOT NULL,
		media_id INTEGER NOT NULL,
		mood_id INTEGER NOT NULL,
		PRIMARY KEY (user_id, media_type, media_id, mood_id),
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(mood_id) REFERENCES moods(id)
	);

	CREATE TABLE IF NOT EXISTS on_this_day_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		month INTEGER NOT NULL,
		day INTEGER NOT NULL,
		entity_type TEXT NOT NULL,
		entity_id INTEGER NOT NULL,
		event_type TEXT NOT NULL,
		year INTEGER,
		description TEXT
	);

	CREATE TABLE IF NOT EXISTS notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		type TEXT NOT NULL,
		actor_id INTEGER,
		entity_type TEXT,
		entity_id INTEGER,
		message TEXT,
		link TEXT,
		is_read BOOLEAN DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(actor_id) REFERENCES users(id)
	);
	`

	_, err := DB.Exec(query)
	if err != nil {
		log.Fatalf("Error creating tables: %v\n", err)
	}
}

func seedDataIfEmpty(tmdbAPIKey string) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM movies").Scan(&count)
	if err != nil {
		log.Fatalf("Error checking movies table: %v\n", err)
	}

	if count == 0 {
		log.Println("Seeding TMDB data into new Schema.org database...")

		// Clear everything so we can insert cleanly since movies is empty.
		createTables()

		// 1. Seed Users
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
		users := []models.User{
			{Username: "adamd", Email: "adam@example.com", Avatar: "/static/img/avatar1.png", ReputationScore: 50, Role: "admin"},
			{Username: "sarah_k", Email: "sarah@example.com", Avatar: "/static/img/avatar2.png", ReputationScore: 10, Role: "user"},
			{Username: "moviebuff99", Email: "buff99@example.com", Avatar: "/static/img/avatar3.png", ReputationScore: 100, Role: "moderator"},
		}

		// Use a single bulk insert to avoid N+1 queries during database initialization
		query := "INSERT INTO users (username, email, password_hash, avatar, reputation_score, role) VALUES "
		var args []interface{}

		for i, u := range users {
			query += "(?, ?, ?, ?, ?, ?)"
			if i < len(users)-1 {
				query += ", "
			}
			args = append(args, u.Username, u.Email, string(hashedPassword), u.Avatar, u.ReputationScore, u.Role)
		}

		_, err = DB.Exec(query, args...)
		if err != nil {
			log.Println("Error inserting users in bulk:", err)
		}

		client := tmdb.NewClient(tmdbAPIKey)

		// Pre-seed genres
		mGenres, err := client.FetchMovieGenres()
		if err == nil {
			for _, g := range mGenres {
				_, _ = DB.Exec("INSERT OR IGNORE INTO genres (id, name, slug) VALUES (?, ?, ?)", g.ID, g.Name, tmdb.Slugify(g.Name))
			}
		}
		tGenres, err := client.FetchTVGenres()
		if err == nil {
			for _, g := range tGenres {
				_, _ = DB.Exec("INSERT OR IGNORE INTO genres (id, name, slug) VALUES (?, ?, ?)", g.ID, g.Name, tmdb.Slugify(g.Name))
			}
		}

		// Pre-seed languages (basic set for testing)
		_, _ = DB.Exec("INSERT OR IGNORE INTO languages (code, name) VALUES ('en', 'English'), ('ja', 'Japanese'), ('ko', 'Korean'), ('es', 'Spanish'), ('fr', 'French')")

		// 2. Fetch and Seed Movies
		movies, err := client.FetchTrendingMovies()
		if err != nil {
			log.Println("Error fetching movies from TMDB:", err)
		} else {
			for _, m := range movies {
				slug := tmdb.Slugify(m.Title)
				if slug == "" {
					slug = "movie"
				}
				langCode := m.OriginalLanguage
				if langCode == "" {
					langCode = "en"
				}
				res, err := DB.Exec("INSERT INTO movies (name, slug, date_published, aggregate_rating, description, image, language_code) VALUES (?, ?, ?, ?, ?, ?, ?)",
					m.Title, slug, m.ReleaseDate, m.VoteAverage, m.Overview, "https://image.tmdb.org/t/p/w500"+m.PosterPath, langCode)
				if err != nil {
					log.Printf("Error inserting movie %s: %v", m.Title, err)
					continue
				}
				movieID, _ := res.LastInsertId()

				for _, gID := range m.GenreIDs {
					_, _ = DB.Exec("INSERT INTO movie_genres (movie_id, genre_id) VALUES (?, ?)", movieID, gID)
				}

				// Fetch Credits
				credits, err := client.FetchMovieCredits(m.ID)
				if err == nil {
					type castInsert struct {
						movieID      int64
						personID     int64
						characterID  int64
						billingOrder int
					}
					var castMappings []castInsert

					for _, cast := range credits.Cast {
						personSlug := tmdb.Slugify(cast.Name)
						if personSlug == "" {
							personSlug = "person"
						}
						var personID int64
						err = DB.QueryRow("SELECT id FROM people WHERE name = ?", cast.Name).Scan(&personID)
						if err != nil {
							var image string
							if cast.ProfilePath != "" {
								image = "https://image.tmdb.org/t/p/w500" + cast.ProfilePath
							}
							res, _ := DB.Exec("INSERT INTO people (name, slug, gender, image) VALUES (?, ?, ?, ?)", cast.Name, personSlug, cast.Gender, image)
							personID, _ = res.LastInsertId()
						}

						characterSlug := tmdb.Slugify(cast.Character)
						if characterSlug == "" {
							characterSlug = "character"
						}
						var charID int64
						err = DB.QueryRow("SELECT id FROM characters WHERE name = ?", cast.Character).Scan(&charID)
						if err != nil {
							res, _ := DB.Exec("INSERT INTO characters (name, slug, gender) VALUES (?, ?, ?)", cast.Character, characterSlug, cast.Gender)
							charID, _ = res.LastInsertId()
						}

						castMappings = append(castMappings, castInsert{
							movieID:      movieID,
							personID:     personID,
							characterID:  charID,
							billingOrder: cast.Order,
						})
					}

					if len(castMappings) > 0 {
						chunkSize := 100 // Safe limit for SQLite (max 32766 params, we use 4 per row: 4 * 100 = 400 parameters)
						for i := 0; i < len(castMappings); i += chunkSize {
							end := i + chunkSize
							if end > len(castMappings) {
								end = len(castMappings)
							}
							chunk := castMappings[i:end]

							valueStrings := make([]string, 0, len(chunk))
							valueArgs := make([]interface{}, 0, len(chunk)*4)

							for _, c := range chunk {
								valueStrings = append(valueStrings, "(?, ?, ?, ?)")
								valueArgs = append(valueArgs, c.movieID, c.personID, c.characterID, c.billingOrder)
							}

							query := fmt.Sprintf("INSERT INTO movie_cast (movie_id, person_id, character_id, billing_order) VALUES %s", strings.Join(valueStrings, ","))
							_, _ = DB.Exec(query, valueArgs...)
						}
					}

					for _, crew := range credits.Crew {
						if crew.Job == "Director" || crew.Job == "Writer" || crew.Job == "Screenplay" || crew.Job == "Author" {
							crewSlug := tmdb.Slugify(crew.Name)
							if crewSlug == "" {
								crewSlug = "crew"
							}
							var personID int64
							err = DB.QueryRow("SELECT id FROM people WHERE name = ?", crew.Name).Scan(&personID)
							if err != nil {
								var image string
								if crew.ProfilePath != "" {
									image = "https://image.tmdb.org/t/p/w500" + crew.ProfilePath
								}
								res, _ := DB.Exec("INSERT INTO people (name, slug, gender, image) VALUES (?, ?, ?, ?)", crew.Name, crewSlug, 0, image)
								personID, _ = res.LastInsertId()
							}

							job := "writer"
							if crew.Job == "Director" {
								job = "director"
							}

							_, _ = DB.Exec("INSERT INTO media_crew (media_type, media_id, person_id, job_title) VALUES (?, ?, ?, ?)", "movie", movieID, personID, job)
						}
					}
					if len(crewParams) > 0 {
						crewQuery := fmt.Sprintf("INSERT INTO media_crew (media_type, media_id, person_id, job_title) VALUES %s", strings.Join(crewPlaceholders, ", "))
						_, _ = DB.Exec(crewQuery, crewParams...)
					}
				}
			}
		}

		// 3. Fetch and Seed TV Shows
		shows, err := client.FetchTrendingShows()
		if err != nil {
			log.Println("Error fetching shows from TMDB:", err)
		} else {
			// Batch Insert TV Series to resolve N+1 issue
			if len(shows) > 0 {
				query := "INSERT INTO tv_series (name, slug, start_date, aggregate_rating, description, image, language_code) VALUES "
				var args []interface{}
				for i, s := range shows {
					slug := tmdb.Slugify(s.Name)
					if slug == "" {
						slug = "show"
					}
					langCode := s.OriginalLanguage
					if langCode == "" {
						langCode = "en"
					}
					query += "(?, ?, ?, ?, ?, ?, ?)"
					if i < len(shows)-1 {
						query += ", "
					}
					args = append(args, s.Name, slug, s.FirstAirDate, s.VoteAverage, s.Overview, "https://image.tmdb.org/t/p/w500"+s.PosterPath, langCode)
				}
				query += " RETURNING id"

				rows, err := DB.Query(query, args...)
				if err != nil {
					log.Printf("Error batch inserting shows: %v", err)
				} else {
					var seriesIDs []int64
					for rows.Next() {
						var id int64
						if err := rows.Scan(&id); err == nil {
							seriesIDs = append(seriesIDs, id)
						}
					}
					rows.Close()

					// Now loop to insert the dependent relations per show
					for i, s := range shows {
						if i >= len(seriesIDs) {
							log.Printf("Skipping dependent insertions for show %s due to missing ID", s.Name)
							continue
						}
						seriesID := seriesIDs[i]

						for _, gID := range s.GenreIDs {
							_, _ = DB.Exec("INSERT INTO tv_genres (series_id, genre_id) VALUES (?, ?)", seriesID, gID)
						}

						// Fetch Credits
						credits, err := client.FetchTVCredits(s.ID)
						if err == nil {
							for _, cast := range credits.Cast {
								personSlug := tmdb.Slugify(cast.Name)
								if personSlug == "" {
									personSlug = "person"
								}
								var personID int64
								err = DB.QueryRow("SELECT id FROM people WHERE name = ?", cast.Name).Scan(&personID)
								if err != nil {
									var image string
									if cast.ProfilePath != "" {
										image = "https://image.tmdb.org/t/p/w500" + cast.ProfilePath
									}
									res, _ := DB.Exec("INSERT INTO people (name, slug, gender, image) VALUES (?, ?, ?, ?)", cast.Name, personSlug, cast.Gender, image)
									personID, _ = res.LastInsertId()
								}

								characterSlug := tmdb.Slugify(cast.Character)
								if characterSlug == "" {
									characterSlug = "character"
								}
								var charID int64
								err = DB.QueryRow("SELECT id FROM characters WHERE name = ?", cast.Character).Scan(&charID)
								if err != nil {
									res, _ := DB.Exec("INSERT INTO characters (name, slug, gender) VALUES (?, ?, ?)", cast.Character, characterSlug, cast.Gender)
									charID, _ = res.LastInsertId()
								}

								_, _ = DB.Exec("INSERT INTO tv_cast (series_id, person_id, character_id, billing_order) VALUES (?, ?, ?, ?)", seriesID, personID, charID, cast.Order)
							}

							for _, crew := range credits.Crew {
								if crew.Job == "Executive Producer" || crew.Job == "Creator" || crew.Job == "Writer" {
									crewSlug := tmdb.Slugify(crew.Name)
									if crewSlug == "" {
										crewSlug = "crew"
									}
									var personID int64
									err = DB.QueryRow("SELECT id FROM people WHERE name = ?", crew.Name).Scan(&personID)
									if err != nil {
										var image string
										if crew.ProfilePath != "" {
											image = "https://image.tmdb.org/t/p/w500" + crew.ProfilePath
										}
										res, _ := DB.Exec("INSERT INTO people (name, slug, gender, image) VALUES (?, ?, ?, ?)", crew.Name, crewSlug, 0, image)
										personID, _ = res.LastInsertId()
									}

									job := "writer"
									if strings.Contains(crew.Job, "Producer") {
										job = "director" // mapping series creators/producers as "directors" for UI simplicity
									}

									_, _ = DB.Exec("INSERT INTO media_crew (media_type, media_id, person_id, job_title) VALUES (?, ?, ?, ?)", "tv_series", seriesID, personID, job)
								}
							}
						}

				// Fetch Episodes for Season 1
				episodes, err := client.FetchTVSeasonEpisodes(s.ID, 1)
				if err == nil && len(episodes) > 0 {
					var vals []interface{}
					var placeholders []string

					for _, ep := range episodes {
						epSlug := tmdb.Slugify(ep.Name)
						if epSlug == "" {
							epSlug = fmt.Sprintf("episode-%d", ep.EpisodeNumber)
						}
						epSlug = fmt.Sprintf("%s-%s", slug, epSlug)

								var image string
								if ep.StillPath != "" {
									image = "https://image.tmdb.org/t/p/w500" + ep.StillPath
								}

						placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
						vals = append(vals, seriesID, ep.SeasonNumber, ep.EpisodeNumber, ep.Name, epSlug, ep.AirDate, ep.Overview, image, ep.Runtime)
					}

					query := fmt.Sprintf(`INSERT INTO tv_episodes
						(series_id, season_number, episode_number, name, slug, date_published, description, image, duration)
						VALUES %s`, strings.Join(placeholders, ","))

					_, err := DB.Exec(query, vals...)

					if err != nil {
						log.Printf("Error inserting batch episodes for series %s: %v", s.Name, err)
					}
				}
			}
		}
	}
}

func GetAllMovies(limit int, offset int, sort string) ([]models.Movie, error) {
	orderBy := "id ASC"
	if sort == "rating" {
		orderBy = "aggregate_rating DESC"
	} else if sort == "date" {
		orderBy = "date_published DESC"
	}

	query := fmt.Sprintf("SELECT id, name, slug, COALESCE(date_published, ''), COALESCE(aggregate_rating, 0.0), COALESCE(description, ''), COALESCE(image, '') FROM movies ORDER BY %s LIMIT %d OFFSET %d", orderBy, limit, offset)
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var m models.Movie
		err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.DatePublished, &m.AggregateRating, &m.Description, &m.Image)
		if err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}
	return movies, nil
}

func GetAllShows(limit int, offset int, sort string) ([]models.TVSeries, error) {
	orderBy := "id ASC"
	if sort == "rating" {
		orderBy = "aggregate_rating DESC"
	} else if sort == "date" {
		orderBy = "start_date DESC"
	}

	query := fmt.Sprintf("SELECT id, name, slug, COALESCE(start_date, ''), COALESCE(end_date, ''), COALESCE(aggregate_rating, 0.0), COALESCE(description, ''), COALESCE(image, ''), COALESCE(number_of_seasons, 0) FROM tv_series ORDER BY %s LIMIT %d OFFSET %d", orderBy, limit, offset)
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shows []models.TVSeries
	for rows.Next() {
		var s models.TVSeries
		err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.StartDate, &s.EndDate, &s.AggregateRating, &s.Description, &s.Image, &s.NumberOfSeasons)
		if err != nil {
			return nil, err
		}
		shows = append(shows, s)
	}
	return shows, nil
}

func GetPopularMovies(limit int) ([]models.Movie, error) {
	query := fmt.Sprintf("SELECT id, name, slug, COALESCE(date_published, ''), COALESCE(aggregate_rating, 0.0), COALESCE(description, ''), COALESCE(image, '') FROM movies ORDER BY aggregate_rating DESC LIMIT %d", limit)
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var m models.Movie
		err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.DatePublished, &m.AggregateRating, &m.Description, &m.Image)
		if err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}
	return movies, nil
}

func GetUpcomingMovies(limit int) ([]models.Movie, error) {
	query := fmt.Sprintf("SELECT id, name, slug, COALESCE(date_published, ''), COALESCE(aggregate_rating, 0.0), COALESCE(description, ''), COALESCE(image, '') FROM movies ORDER BY date_published DESC LIMIT %d", limit)
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var m models.Movie
		err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.DatePublished, &m.AggregateRating, &m.Description, &m.Image)
		if err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}
	return movies, nil
}

func GetPopularShows(limit int) ([]models.TVSeries, error) {
	query := fmt.Sprintf("SELECT id, name, slug, COALESCE(start_date, ''), COALESCE(end_date, ''), COALESCE(aggregate_rating, 0.0), COALESCE(description, ''), COALESCE(image, ''), COALESCE(number_of_seasons, 0) FROM tv_series ORDER BY aggregate_rating DESC LIMIT %d", limit)
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shows []models.TVSeries
	for rows.Next() {
		var s models.TVSeries
		err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.StartDate, &s.EndDate, &s.AggregateRating, &s.Description, &s.Image, &s.NumberOfSeasons)
		if err != nil {
			return nil, err
		}
		shows = append(shows, s)
	}
	return shows, nil
}

func GetNewShows(limit int) ([]models.TVSeries, error) {
	query := fmt.Sprintf("SELECT id, name, slug, COALESCE(start_date, ''), COALESCE(end_date, ''), COALESCE(aggregate_rating, 0.0), COALESCE(description, ''), COALESCE(image, ''), COALESCE(number_of_seasons, 0) FROM tv_series ORDER BY start_date DESC LIMIT %d", limit)
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shows []models.TVSeries
	for rows.Next() {
		var s models.TVSeries
		err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.StartDate, &s.EndDate, &s.AggregateRating, &s.Description, &s.Image, &s.NumberOfSeasons)
		if err != nil {
			return nil, err
		}
		shows = append(shows, s)
	}
	return shows, nil
}

func GetAllPeople(limit int, offset int, sort string) ([]models.Person, error) {
	orderBy := "id ASC"
	if sort == "name" {
		orderBy = "name ASC"
	}

	query := fmt.Sprintf("SELECT id, name, slug, COALESCE(gender, ''), COALESCE(image, '') FROM people ORDER BY %s LIMIT %d OFFSET %d", orderBy, limit, offset)
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []models.Person
	for rows.Next() {
		var p models.Person
		err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Gender, &p.Image)
		if err != nil {
			return nil, err
		}
		people = append(people, p)
	}
	return people, nil
}

func SearchMovies(searchQuery string, limit int, offset int) ([]models.Movie, error) {
	query := fmt.Sprintf("SELECT id, name, slug, COALESCE(date_published, ''), COALESCE(aggregate_rating, 0.0), COALESCE(description, ''), COALESCE(image, '') FROM movies WHERE name LIKE ? COLLATE NOCASE LIMIT %d OFFSET %d", limit, offset)
	rows, err := DB.Query(query, "%"+searchQuery+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var m models.Movie
		err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.DatePublished, &m.AggregateRating, &m.Description, &m.Image)
		if err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}
	return movies, nil
}

func SearchShows(searchQuery string, limit int, offset int) ([]models.TVSeries, error) {
	query := fmt.Sprintf("SELECT id, name, slug, COALESCE(start_date, ''), COALESCE(end_date, ''), COALESCE(aggregate_rating, 0.0), COALESCE(description, ''), COALESCE(image, ''), COALESCE(number_of_seasons, 0) FROM tv_series WHERE name LIKE ? COLLATE NOCASE LIMIT %d OFFSET %d", limit, offset)
	rows, err := DB.Query(query, "%"+searchQuery+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shows []models.TVSeries
	for rows.Next() {
		var s models.TVSeries
		err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.StartDate, &s.EndDate, &s.AggregateRating, &s.Description, &s.Image, &s.NumberOfSeasons)
		if err != nil {
			return nil, err
		}
		shows = append(shows, s)
	}
	return shows, nil
}

func SearchPeople(searchQuery string, limit int, offset int) ([]models.Person, error) {
	query := fmt.Sprintf("SELECT id, name, slug, COALESCE(gender, ''), COALESCE(image, '') FROM people WHERE name LIKE ? COLLATE NOCASE LIMIT %d OFFSET %d", limit, offset)
	rows, err := DB.Query(query, "%"+searchQuery+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []models.Person
	for rows.Next() {
		var p models.Person
		err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Gender, &p.Image)
		if err != nil {
			return nil, err
		}
		people = append(people, p)
	}
	return people, nil
}

func GetAllUsers(limit int, offset int) ([]models.User, error) {
	query := fmt.Sprintf("SELECT id, username, email, COALESCE(avatar, ''), COALESCE(reputation_score, 0), COALESCE(role, 'user') FROM users LIMIT %d OFFSET %d", limit, offset)
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Avatar, &u.ReputationScore, &u.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// GetMovieByID fetches a single movie by ID
// GetMovieGenres fetches genres for a specific movie
func GetMovieGenres(movieID int) ([]models.Genre, error) {
	query := `
		SELECT g.id, g.name, g.slug
		FROM genres g
		JOIN movie_genres mg ON g.id = mg.genre_id
		WHERE mg.movie_id = ?
	`
	rows, err := DB.Query(query, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []models.Genre
	for rows.Next() {
		var g models.Genre
		if err := rows.Scan(&g.ID, &g.Name, &g.Slug); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	return genres, nil
}

func GetMovieByID(id int) (models.Movie, error) {
	var m models.Movie
	var budget, boxOffice, langCode, countryCode, tagline, subtitle sql.NullString
	err := DB.QueryRow("SELECT id, name, slug, COALESCE(date_published, ''), COALESCE(aggregate_rating, 0.0), COALESCE(description, ''), COALESCE(image, ''), COALESCE(budget, ''), COALESCE(box_office, ''), COALESCE(language_code, ''), COALESCE(country_code, ''), COALESCE(tagline, ''), rating_count, review_count, best_rating, worst_rating, is_family_friendly, COALESCE(subtitle, '') FROM movies WHERE id = ?", id).
		Scan(&m.ID, &m.Name, &m.Slug, &m.DatePublished, &m.AggregateRating, &m.Description, &m.Image, &budget, &boxOffice, &langCode, &countryCode, &tagline, &m.RatingCount, &m.ReviewCount, &m.BestRating, &m.WorstRating, &m.IsFamilyFriendly, &subtitle)
	m.Budget = budget.String
	m.BoxOffice = boxOffice.String
	m.LanguageCode = langCode.String
	m.CountryCode = countryCode.String
	m.Tagline = tagline.String
	m.Subtitle = subtitle.String

	// Fetch genres
	genres, _ := GetMovieGenres(m.ID)
	m.Genres = genres

	return m, err
}

// GetMovieCast fetches the cast list for a given movie ID
func GetMovieCast(movieID int) ([]models.CastMember, error) {
	query := `
		SELECT 
			p.id, p.name, p.slug, COALESCE(p.image, ''), 
			c.id, c.name, c.slug, COALESCE(c.image, ''), 
			COALESCE(mc.billing_order, 0)
		FROM movie_cast mc
		JOIN people p ON mc.person_id = p.id
		JOIN characters c ON mc.character_id = c.id
		WHERE mc.movie_id = ?
		ORDER BY mc.billing_order ASC
	`
	rows, err := DB.Query(query, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cast []models.CastMember
	for rows.Next() {
		var cm models.CastMember
		// Handle potential NULL images gracefully if needed using sql.NullString,
		// but since we define them as TEXT we can scan directly to string for now.
		var pImg, cImg sql.NullString
		err := rows.Scan(
			&cm.Person.ID, &cm.Person.Name, &cm.Person.Slug, &pImg,
			&cm.Character.ID, &cm.Character.Name, &cm.Character.Slug, &cImg,
			&cm.BillingOrder,
		)
		if err != nil {
			return nil, err
		}
		cm.Person.Image = pImg.String
		cm.Character.Image = cImg.String
		cast = append(cast, cm)
	}
	return cast, nil
}

// GetMovieDetail fetches a movie and all its related entities by ID.
// This acts as a single aggregation function so the UI layer gets a complete
// MovieDetail model without having to make multiple separate DB calls itself.
func GetMovieDetail(id int) (models.MovieDetail, error) {
	var detail models.MovieDetail

	// First, fetch the core movie information from the 'movies' table.
	movie, err := GetMovieByID(id)
	if err != nil {
		return detail, err // If the main movie data fails, abort and return the error.
	}
	detail.Movie = movie // Assign the fetched movie data to the response struct.

	// Second, fetch the cast list by joining 'movie_cast', 'people', and 'characters'.
	cast, err := GetMovieCast(id)
	if err != nil {
		// Log the error but don't fail the entire request; the movie page can still render without cast.
		log.Printf("Error fetching cast for movie %d: %v", id, err)
	} else {
		detail.Cast = cast // Assign the slice of cast members.
	}

	// Third, fetch crew members (directors and writers) using the 'media_crew' table.
	directors, writers, err := GetMediaCrew("movie", movie.ID)
	if err == nil {
		detail.Directors = directors
		detail.Writers = writers
	}

	// Return the populated detail struct containing the movie, cast, directors, and writers.
	return detail, nil
}

// GetTVSeriesByID fetches a single TV show by ID
// GetTVSeriesGenres fetches genres for a specific TV show
func GetTVSeriesGenres(seriesID int) ([]models.Genre, error) {
	query := `
		SELECT g.id, g.name, g.slug
		FROM genres g
		JOIN tv_genres tg ON g.id = tg.genre_id
		WHERE tg.series_id = ?
	`
	rows, err := DB.Query(query, seriesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []models.Genre
	for rows.Next() {
		var g models.Genre
		if err := rows.Scan(&g.ID, &g.Name, &g.Slug); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	return genres, nil
}

func GetTVSeriesByID(id int) (models.TVSeries, error) {
	var s models.TVSeries
	var langCode, countryCode, tagline, subtitle sql.NullString
	err := DB.QueryRow("SELECT id, name, slug, COALESCE(start_date, ''), COALESCE(end_date, ''), COALESCE(aggregate_rating, 0.0), COALESCE(description, ''), COALESCE(image, ''), COALESCE(number_of_seasons, 0), COALESCE(language_code, ''), COALESCE(country_code, ''), COALESCE(tagline, ''), rating_count, review_count, best_rating, worst_rating, COALESCE(subtitle, '') FROM tv_series WHERE id = ?", id).
		Scan(&s.ID, &s.Name, &s.Slug, &s.StartDate, &s.EndDate, &s.AggregateRating, &s.Description, &s.Image, &s.NumberOfSeasons, &langCode, &countryCode, &tagline, &s.RatingCount, &s.ReviewCount, &s.BestRating, &s.WorstRating, &subtitle)
	s.LanguageCode = langCode.String
	s.CountryCode = countryCode.String
	s.Tagline = tagline.String
	s.Subtitle = subtitle.String

	genres, _ := GetTVSeriesGenres(s.ID)
	s.Genres = genres

	return s, err
}

// GetTVSeriesCast fetches the cast list for a given series ID
func GetTVSeriesCast(seriesID int) ([]models.CastMember, error) {
	query := `
		SELECT 
			p.id, p.name, p.slug, p.image, 
			c.id, c.name, c.slug, c.image, 
			tc.billing_order
		FROM tv_cast tc
		JOIN people p ON tc.person_id = p.id
		JOIN characters c ON tc.character_id = c.id
		WHERE tc.series_id = ?
		ORDER BY tc.billing_order ASC
	`
	rows, err := DB.Query(query, seriesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cast []models.CastMember
	for rows.Next() {
		var cm models.CastMember
		var pImg, cImg sql.NullString
		err := rows.Scan(
			&cm.Person.ID, &cm.Person.Name, &cm.Person.Slug, &pImg,
			&cm.Character.ID, &cm.Character.Name, &cm.Character.Slug, &cImg,
			&cm.BillingOrder,
		)
		if err != nil {
			return nil, err
		}
		cm.Person.Image = pImg.String
		cm.Character.Image = cImg.String
		cast = append(cast, cm)
	}
	return cast, nil
}

// GetTVEpisodes fetches episodes for a given series
func GetTVEpisodes(seriesID int) ([]models.TVEpisode, error) {
	query := `
		SELECT id, series_id, season_number, episode_number, name, slug, date_published, description, image, duration
		FROM tv_episodes
		WHERE series_id = ?
		ORDER BY season_number ASC, episode_number ASC
	`
	rows, err := DB.Query(query, seriesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eps []models.TVEpisode
	for rows.Next() {
		var ep models.TVEpisode
		var dp, desc, img sql.NullString
		var dur sql.NullInt64
		err := rows.Scan(&ep.ID, &ep.SeriesID, &ep.SeasonNumber, &ep.EpisodeNumber, &ep.Name, &ep.Slug, &dp, &desc, &img, &dur)
		if err != nil {
			return nil, err
		}
		ep.DatePublished = dp.String
		ep.Description = desc.String
		ep.Image = img.String
		ep.Duration = int(dur.Int64)
		eps = append(eps, ep)
	}
	return eps, nil
}

// GetTVSeriesDetail fetches a show, its cast, and episodes by ID
func GetTVSeriesDetail(id int) (models.TVSeriesDetail, error) {
	var detail models.TVSeriesDetail

	series, err := GetTVSeriesByID(id)
	if err != nil {
		return detail, err
	}
	detail.Series = series

	cast, err := GetTVSeriesCast(id)
	if err == nil {
		detail.Cast = cast
	}

	eps, err := GetTVEpisodes(id)
	if err == nil {
		detail.Episodes = eps
	}

	directors, writers, err := GetMediaCrew("tv_series", series.ID)
	if err == nil {
		detail.Directors = directors
		detail.Writers = writers
	}

	return detail, nil
}

// GetMediaCrew fetches the directors and writers for a given media (movie or tv_series)
func GetMediaCrew(mediaType string, mediaID int) (directors []models.Person, writers []models.Person, err error) {
	query := `
		SELECT p.id, p.name, p.slug, COALESCE(p.gender, 0), pos.job_title
		FROM media_crew pos
		JOIN people p ON pos.person_id = p.id
		WHERE pos.media_type = ? AND pos.media_id = ?
	`
	rows, err := DB.Query(query, mediaType, mediaID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Person
		var job string
		var gender sql.NullString // Catch generic DB null fields

		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &gender, &job); err != nil {
			continue // Skip errors and grab what we can
		}
		p.Gender = gender.String

		if job == "director" {
			directors = append(directors, p)
		} else if job == "writer" {
			writers = append(writers, p)
		}
	}

	return directors, writers, nil
}

// GetPersonByID fetches a single person
func GetPersonByID(id int) (models.Person, error) {
	var p models.Person
	var knowsLang, natCode, dept sql.NullString
	err := DB.QueryRow("SELECT id, name, slug, COALESCE(gender, ''), COALESCE(birth_date, ''), COALESCE(description, ''), COALESCE(image, ''), COALESCE(knows_language, ''), COALESCE(nationality_code, ''), COALESCE(known_for_department, ''), popularity_score FROM people WHERE id = ?", id).
		Scan(&p.ID, &p.Name, &p.Slug, &p.Gender, &p.BirthDate, &p.Description, &p.Image, &knowsLang, &natCode, &dept, &p.PopularityScore)
	p.KnowsLanguage = knowsLang.String
	p.NationalityCode = natCode.String
	p.KnownForDepartment = dept.String
	return p, err
}

// GetPersonDetailByID fetches a person profile
func GetPersonDetailByID(id int) (models.PersonDetail, error) {
	var detail models.PersonDetail
	person, err := GetPersonByID(id)
	if err != nil {
		return detail, err
	}
	detail.Person = person
	movies, _ := GetPersonMovies(person.ID)
	shows, _ := GetPersonShows(person.ID)
	detail.Movies = movies
	detail.Shows = shows
	return detail, nil
}

// GetPersonMovies fetches movies a person was cast in or crewed on
func GetPersonMovies(personID int) ([]models.Movie, error) {
	query := `
		SELECT DISTINCT m.id, m.name, m.slug, COALESCE(m.date_published, ''), COALESCE(m.aggregate_rating, 0.0), COALESCE(m.description, ''), COALESCE(m.image, '')
		FROM movies m
		LEFT JOIN movie_cast mc ON mc.movie_id = m.id AND mc.person_id = ?
		LEFT JOIN media_crew cr ON cr.media_id = m.id AND cr.media_type = 'movie' AND cr.person_id = ?
		WHERE mc.person_id IS NOT NULL OR cr.person_id IS NOT NULL
		ORDER BY m.date_published DESC
	`
	rows, err := DB.Query(query, personID, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.DatePublished, &m.AggregateRating, &m.Description, &m.Image); err == nil {
			movies = append(movies, m)
		}
	}
	return movies, nil
}

// GetPersonShows fetches shows a person was cast in or crewed on
func GetPersonShows(personID int) ([]models.TVSeries, error) {
	query := `
		SELECT DISTINCT s.id, s.name, s.slug, COALESCE(s.start_date, ''), COALESCE(s.end_date, ''), COALESCE(s.aggregate_rating, 0.0), COALESCE(s.description, ''), COALESCE(s.image, ''), COALESCE(s.number_of_seasons, 0)
		FROM tv_series s
		LEFT JOIN tv_cast tc ON tc.series_id = s.id AND tc.person_id = ?
		LEFT JOIN media_crew cr ON cr.media_id = s.id AND cr.media_type = 'tv_series' AND cr.person_id = ?
		WHERE tc.person_id IS NOT NULL OR cr.person_id IS NOT NULL
		ORDER BY s.start_date DESC
	`
	rows, err := DB.Query(query, personID, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shows []models.TVSeries
	for rows.Next() {
		var s models.TVSeries
		if err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.StartDate, &s.EndDate, &s.AggregateRating, &s.Description, &s.Image, &s.NumberOfSeasons); err == nil {
			shows = append(shows, s)
		}
	}
	return shows, nil
}

// GetUserWatchlist fetches all movies and tv shows from a user's primary watchlist
func GetUserWatchlist(userID int) ([]models.Movie, []models.TVSeries, error) {
	var watchlistID int
	err := DB.QueryRow("SELECT id FROM watchlists WHERE user_id = ? LIMIT 1", userID).Scan(&watchlistID)
	if err != nil {
		return []models.Movie{}, []models.TVSeries{}, nil
	}

	movieQuery := `
		SELECT m.id, m.name, m.slug, COALESCE(m.date_published, ''), COALESCE(m.aggregate_rating, 0.0), COALESCE(m.description, ''), COALESCE(m.image, '')
		FROM movies m
		JOIN watchlist_items wi ON m.id = wi.media_id
		WHERE wi.watchlist_id = ? AND wi.media_type = 'movie'
		ORDER BY wi.added_at DESC
	`
	movieRows, err := DB.Query(movieQuery, watchlistID)
	if err != nil {
		return nil, nil, err
	}
	defer movieRows.Close()

	var movies []models.Movie
	for movieRows.Next() {
		var m models.Movie
		if err := movieRows.Scan(&m.ID, &m.Name, &m.Slug, &m.DatePublished, &m.AggregateRating, &m.Description, &m.Image); err != nil {
			return nil, nil, err
		}
		movies = append(movies, m)
	}

	showQuery := `
		SELECT s.id, s.name, s.slug, COALESCE(s.start_date, ''), COALESCE(s.end_date, ''), COALESCE(s.aggregate_rating, 0.0), COALESCE(s.description, ''), COALESCE(s.image, ''), COALESCE(s.number_of_seasons, 0)
		FROM tv_series s
		JOIN watchlist_items wi ON s.id = wi.media_id
		WHERE wi.watchlist_id = ? AND wi.media_type = 'tv'
		ORDER BY wi.added_at DESC
	`
	showRows, err := DB.Query(showQuery, watchlistID)
	if err != nil {
		return nil, nil, err
	}
	defer showRows.Close()

	var shows []models.TVSeries
	for showRows.Next() {
		var s models.TVSeries
		if err := showRows.Scan(&s.ID, &s.Name, &s.Slug, &s.StartDate, &s.EndDate, &s.AggregateRating, &s.Description, &s.Image, &s.NumberOfSeasons); err != nil {
			return nil, nil, err
		}
		shows = append(shows, s)
	}

	return movies, shows, nil
}
