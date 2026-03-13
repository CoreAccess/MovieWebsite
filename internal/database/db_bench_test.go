package database

import (
	"database/sql"
	"fmt"
	"testing"

	_ "modernc.org/sqlite"
)

func setupBenchDB(b *testing.B) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		b.Fatalf("Failed to open DB: %v", err)
	}

	createStmt := `
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
		UNIQUE(name, start_date)
	);`

	_, err = db.Exec(createStmt)
	if err != nil {
		b.Fatalf("Failed to create tables: %v", err)
	}
	return db
}

// Local mock of TMDB structure for the benchmark
type TMDBShow struct {
	Name             string
	FirstAirDate     string
	VoteAverage      float64
	Overview         string
	PosterPath       string
	OriginalLanguage string
}

func generateShows(count int) []TMDBShow {
	shows := make([]TMDBShow, count)
	for i := 0; i < count; i++ {
		shows[i] = TMDBShow{
			Name:             fmt.Sprintf("Test Show %d", i),
			FirstAirDate:     "2023-01-01",
			VoteAverage:      8.5,
			Overview:         "A test show.",
			PosterPath:       "/test.jpg",
			OriginalLanguage: "en",
		}
	}
	return shows
}

func BenchmarkTVSeriesInsert_Sequential(b *testing.B) {
	db := setupBenchDB(b)
	defer db.Close()
	shows := generateShows(20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clean table for consistent benchmark
		db.Exec("DELETE FROM tv_series")

		for _, s := range shows {
			slug := "test-show"
			langCode := s.OriginalLanguage
			if langCode == "" { langCode = "en" }
			res, err := db.Exec("INSERT INTO tv_series (name, slug, start_date, aggregate_rating, description, image, language_code) VALUES (?, ?, ?, ?, ?, ?, ?)",
				s.Name, slug, s.FirstAirDate, s.VoteAverage, s.Overview, "https://image.tmdb.org/t/p/w500"+s.PosterPath, langCode)
			if err != nil {
				b.Fatalf("Error inserting show %s: %v", s.Name, err)
			}
			_, _ = res.LastInsertId()
		}
	}
}

func BenchmarkTVSeriesInsert_Batched(b *testing.B) {
	db := setupBenchDB(b)
	defer db.Close()
	shows := generateShows(20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clean table for consistent benchmark
		db.Exec("DELETE FROM tv_series")

		if len(shows) > 0 {
			query := "INSERT INTO tv_series (name, slug, start_date, aggregate_rating, description, image, language_code) VALUES "
			var args []interface{}
			for j, s := range shows {
				slug := "test-show"
				langCode := s.OriginalLanguage
				if langCode == "" { langCode = "en" }
				query += "(?, ?, ?, ?, ?, ?, ?)"
				if j < len(shows)-1 {
					query += ", "
				}
				args = append(args, s.Name, slug, s.FirstAirDate, s.VoteAverage, s.Overview, "https://image.tmdb.org/t/p/w500"+s.PosterPath, langCode)
			}
			query += " RETURNING id"

			rows, err := db.Query(query, args...)
			if err != nil {
				b.Fatalf("Error inserting batch: %v", err)
			}

			var seriesIDs []int64
			for rows.Next() {
				var id int64
				if err := rows.Scan(&id); err != nil {
					b.Fatalf("Error scanning ID: %v", err)
				}
				seriesIDs = append(seriesIDs, id)
			}
			rows.Close()

			if len(seriesIDs) != len(shows) {
				b.Fatalf("Expected %d IDs, got %d", len(shows), len(seriesIDs))
			}
		}
	}
}
