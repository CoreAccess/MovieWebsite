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

var DB *sql.DB

func InitDB(dataSourceName string, tmdbAPIKey string) (*sql.DB, error) {
	var err error
	DB, err = sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = DB.Ping(); err != nil {
		return nil, err
	}

	log.Println("Connected to SQLite Database")
	createTables()
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
		founding_date TEXT
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
		also_known_as TEXT,
		awards TEXT,
		knows_language TEXT
	);

	CREATE TABLE IF NOT EXISTS characters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		slug TEXT NOT NULL,
		gender TEXT,
		birth_date TEXT,
		death_date TEXT,
		description TEXT,
		image TEXT
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
		genre JSON,
		budget TEXT,
		box_office TEXT,
		in_language TEXT,
		production_company TEXT,
		keywords TEXT,
		UNIQUE(name, date_published)
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
		genre JSON,
		production_company TEXT,
		in_language TEXT,
		UNIQUE(name, start_date)
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
		changes TEXT NOT NULL,
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
		target_pages TEXT, -- JSON
		impressions INTEGER DEFAULT 0,
		clicks INTEGER DEFAULT 0,
		start_date TIMESTAMP,
		end_date TIMESTAMP
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
		options TEXT NOT NULL, -- JSON
		FOREIGN KEY(post_id) REFERENCES posts(id)
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
		for _, u := range users {
			_, err = DB.Exec("INSERT INTO users (username, email, password_hash, avatar, reputation_score, role) VALUES (?, ?, ?, ?, ?, ?)", u.Username, u.Email, string(hashedPassword), u.Avatar, u.ReputationScore, u.Role)
			if err != nil {
				log.Println("Error inserting user:", err)
			}
		}

		client := tmdb.NewClient(tmdbAPIKey)
		
		// 2. Fetch and Seed Movies
		movies, err := client.FetchTrendingMovies()
		if err != nil {
			log.Println("Error fetching movies from TMDB:", err)
		} else {
			for _, m := range movies {
				slug := tmdb.Slugify(m.Title)
				if slug == "" { slug = "movie" }
				res, err := DB.Exec("INSERT INTO movies (name, slug, date_published, aggregate_rating, description, image, genre) VALUES (?, ?, ?, ?, ?, ?, ?)", 
					m.Title, slug, m.ReleaseDate, m.VoteAverage, m.Overview, "https://image.tmdb.org/t/p/w500"+m.PosterPath, "[]")
				if err != nil {
					log.Printf("Error inserting movie %s: %v", m.Title, err)
					continue
				}
				movieID, _ := res.LastInsertId()

				// Fetch Credits
				credits, err := client.FetchMovieCredits(m.ID)
				if err == nil {
					for _, cast := range credits.Cast {
						personSlug := tmdb.Slugify(cast.Name)
						if personSlug == "" { personSlug = "person" }
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
						
						_, _ = DB.Exec("INSERT INTO movie_cast (movie_id, person_id, character_id, billing_order) VALUES (?, ?, ?, ?)", movieID, personID, charID, cast.Order)
					}
					
					for _, crew := range credits.Crew {
						if crew.Job == "Director" || crew.Job == "Writer" || crew.Job == "Screenplay" || crew.Job == "Author" {
							crewSlug := tmdb.Slugify(crew.Name)
							if crewSlug == "" { crewSlug = "crew" }
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
				}
			}
		}

		// 3. Fetch and Seed TV Shows
		shows, err := client.FetchTrendingShows()
		if err != nil {
			log.Println("Error fetching shows from TMDB:", err)
		} else {
			for _, s := range shows {
				slug := tmdb.Slugify(s.Name)
				if slug == "" { slug = "show" }
				res, err := DB.Exec("INSERT INTO tv_series (name, slug, start_date, aggregate_rating, description, image, genre) VALUES (?, ?, ?, ?, ?, ?, ?)", 
					s.Name, slug, s.FirstAirDate, s.VoteAverage, s.Overview, "https://image.tmdb.org/t/p/w500"+s.PosterPath, "[]")
				if err != nil {
					log.Printf("Error inserting show %s: %v", s.Name, err)
					continue
				}
				seriesID, _ := res.LastInsertId()

				// Fetch Credits
				credits, err := client.FetchTVCredits(s.ID)
				if err == nil {
					for _, cast := range credits.Cast {
						personSlug := tmdb.Slugify(cast.Name)
						if personSlug == "" { personSlug = "person" }
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
							if crewSlug == "" { crewSlug = "crew" }
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
				if err == nil {
					for _, ep := range episodes {
						epSlug := tmdb.Slugify(ep.Name)
						if epSlug == "" { epSlug = fmt.Sprintf("episode-%d", ep.EpisodeNumber) }
						epSlug = fmt.Sprintf("%s-%s", slug, epSlug)
						
						var image string
						if ep.StillPath != "" {
							image = "https://image.tmdb.org/t/p/w500" + ep.StillPath
						}

						_, err := DB.Exec(`INSERT INTO tv_episodes 
							(series_id, season_number, episode_number, name, slug, date_published, description, image, duration) 
							VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
							seriesID, ep.SeasonNumber, ep.EpisodeNumber, ep.Name, epSlug, ep.AirDate, ep.Overview, image, ep.Runtime)
						
						if err != nil {
							log.Printf("Error inserting episode %s: %v", ep.Name, err)
						}
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
func GetMovieByID(id int) (models.Movie, error) {
	var m models.Movie
	var budget, boxOffice, inLanguage, productionCompany, keywords sql.NullString
	err := DB.QueryRow("SELECT id, name, slug, COALESCE(date_published, ''), COALESCE(aggregate_rating, 0.0), COALESCE(description, ''), COALESCE(image, ''), COALESCE(budget, ''), COALESCE(box_office, ''), COALESCE(in_language, ''), COALESCE(production_company, ''), COALESCE(keywords, '') FROM movies WHERE id = ?", id).
		Scan(&m.ID, &m.Name, &m.Slug, &m.DatePublished, &m.AggregateRating, &m.Description, &m.Image, &budget, &boxOffice, &inLanguage, &productionCompany, &keywords)
	m.Budget = budget.String
	m.BoxOffice = boxOffice.String
	m.InLanguage = inLanguage.String
	m.ProductionCompany = productionCompany.String
	m.Keywords = keywords.String
	return m, err
}

// GetMovieBySlug fetches a single movie by its slug
func GetMovieBySlug(slug string) (models.Movie, error) {
	var m models.Movie
	var budget, boxOffice, inLanguage, productionCompany, keywords sql.NullString
	err := DB.QueryRow("SELECT id, name, slug, COALESCE(date_published, ''), COALESCE(aggregate_rating, 0.0), COALESCE(description, ''), COALESCE(image, ''), COALESCE(budget, ''), COALESCE(box_office, ''), COALESCE(in_language, ''), COALESCE(production_company, ''), COALESCE(keywords, '') FROM movies WHERE slug = ?", slug).
		Scan(&m.ID, &m.Name, &m.Slug, &m.DatePublished, &m.AggregateRating, &m.Description, &m.Image, &budget, &boxOffice, &inLanguage, &productionCompany, &keywords)
	m.Budget = budget.String
	m.BoxOffice = boxOffice.String
	m.InLanguage = inLanguage.String
	m.ProductionCompany = productionCompany.String
	m.Keywords = keywords.String
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

// GetMovieDetail fetches a movie and all its related entities by ID
func GetMovieDetail(id int) (models.MovieDetail, error) {
	var detail models.MovieDetail
	
	movie, err := GetMovieByID(id)
	if err != nil {
		return detail, err
	}
	detail.Movie = movie

	cast, err := GetMovieCast(id)
	if err != nil {
		log.Printf("Error fetching cast for movie %d: %v", id, err)
		// Don't fail the whole request just because cast is missing
	} else {
		detail.Cast = cast
	}

	directors, writers, err := GetMediaCrew("movie", movie.ID)
	if err == nil {
		detail.Directors = directors
		detail.Writers = writers
	}

	return detail, nil
}

// GetMovieDetailBySlug fetches a movie and all its related entities by slug
func GetMovieDetailBySlug(slug string) (models.MovieDetail, error) {
	var detail models.MovieDetail

	movie, err := GetMovieBySlug(slug)
	if err != nil {
		return detail, err
	}
	detail.Movie = movie

	cast, err := GetMovieCast(movie.ID)
	if err != nil {
		log.Printf("Error fetching cast for movie %s: %v", slug, err)
	} else {
		detail.Cast = cast
	}

	directors, writers, err := GetMediaCrew("movie", movie.ID)
	if err == nil {
		detail.Directors = directors
		detail.Writers = writers
	}

	return detail, nil
}

// GetTVSeriesByID fetches a single TV show by ID
func GetTVSeriesByID(id int) (models.TVSeries, error) {
	var s models.TVSeries
	var prodCo, inLang sql.NullString
	err := DB.QueryRow("SELECT id, name, slug, COALESCE(start_date, ''), COALESCE(end_date, ''), COALESCE(aggregate_rating, 0.0), COALESCE(description, ''), COALESCE(image, ''), COALESCE(number_of_seasons, 0), COALESCE(production_company, ''), COALESCE(in_language, '') FROM tv_series WHERE id = ?", id).
		Scan(&s.ID, &s.Name, &s.Slug, &s.StartDate, &s.EndDate, &s.AggregateRating, &s.Description, &s.Image, &s.NumberOfSeasons, &prodCo, &inLang)
	s.ProductionCompany = prodCo.String
	s.InLanguage = inLang.String
	return s, err
}

// GetTVSeriesBySlug fetches a single TV show by its slug
func GetTVSeriesBySlug(slug string) (models.TVSeries, error) {
	var s models.TVSeries
	var prodCo, inLang sql.NullString
	err := DB.QueryRow("SELECT id, name, slug, COALESCE(start_date, ''), COALESCE(end_date, ''), COALESCE(aggregate_rating, 0.0), COALESCE(description, ''), COALESCE(image, ''), COALESCE(number_of_seasons, 0), COALESCE(production_company, ''), COALESCE(in_language, '') FROM tv_series WHERE slug = ?", slug).
		Scan(&s.ID, &s.Name, &s.Slug, &s.StartDate, &s.EndDate, &s.AggregateRating, &s.Description, &s.Image, &s.NumberOfSeasons, &prodCo, &inLang)
	s.ProductionCompany = prodCo.String
	s.InLanguage = inLang.String
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

// GetTVSeriesDetailBySlug fetches a show, its cast, and episodes by slug
func GetTVSeriesDetailBySlug(slug string) (models.TVSeriesDetail, error) {
	var detail models.TVSeriesDetail

	series, err := GetTVSeriesBySlug(slug)
	if err != nil {
		return detail, err
	}
	detail.Series = series

	cast, err := GetTVSeriesCast(series.ID)
	if err == nil {
		detail.Cast = cast
	}

	eps, err := GetTVEpisodes(series.ID)
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

// GetPersonBySlug fetches a single person
func GetPersonBySlug(slug string) (models.Person, error) {
	var p models.Person
	var aka, awards, knowsLang sql.NullString
	err := DB.QueryRow("SELECT id, name, slug, COALESCE(gender, ''), COALESCE(birth_date, ''), COALESCE(description, ''), COALESCE(image, ''), COALESCE(also_known_as, ''), COALESCE(awards, ''), COALESCE(knows_language, '') FROM people WHERE slug = ?", slug).
		Scan(&p.ID, &p.Name, &p.Slug, &p.Gender, &p.BirthDate, &p.Description, &p.Image, &aka, &awards, &knowsLang)
	p.AlsoKnownAs = aka.String
	p.Awards = awards.String
	p.KnowsLanguage = knowsLang.String
	return p, err
}

// GetPersonDetail fetches a person profile
func GetPersonDetail(slug string) (models.PersonDetail, error) {
	var detail models.PersonDetail
	person, err := GetPersonBySlug(slug)
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

// GetPersonByID fetches a single person
func GetPersonByID(id int) (models.Person, error) {
	var p models.Person
	var aka, awards, knowsLang sql.NullString
	err := DB.QueryRow("SELECT id, name, slug, COALESCE(gender, ''), COALESCE(birth_date, ''), COALESCE(description, ''), COALESCE(image, ''), COALESCE(also_known_as, ''), COALESCE(awards, ''), COALESCE(knows_language, '') FROM people WHERE id = ?", id).
		Scan(&p.ID, &p.Name, &p.Slug, &p.Gender, &p.BirthDate, &p.Description, &p.Image, &aka, &awards, &knowsLang)
	p.AlsoKnownAs = aka.String
	p.Awards = awards.String
	p.KnowsLanguage = knowsLang.String
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
