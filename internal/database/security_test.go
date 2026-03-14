package database

import (
	"database/sql"
	"testing"
	_ "modernc.org/sqlite"
)

func TestSearchMovies_Security(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}
	defer db.Close()

	// Set the global DB for the test
	originalDB := DB
	DB = db
	defer func() { DB = originalDB }()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS movies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL,
			date_published TEXT,
			aggregate_rating REAL,
			description TEXT,
			image TEXT
		);
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Insert dummy data
	_, err = db.Exec("INSERT INTO movies (name, slug) VALUES ('Inception', 'inception'), ('Interstellar', 'interstellar'), ('The Prestige', 'the-prestige')")
	if err != nil {
		t.Fatalf("failed to insert data: %v", err)
	}

	tests := []struct {
		name        string
		searchQuery string
		limit       int
		offset      int
		wantCount   int
	}{
		{"Normal search", "Inception", 10, 0, 1},
		{"Empty search", "", 10, 0, 3},
		{"Limit 1", "", 1, 0, 1},
		{"Offset 1", "", 10, 1, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SearchMovies(tt.searchQuery, tt.limit, tt.offset)
			if err != nil {
				t.Errorf("SearchMovies() error = %v", err)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("SearchMovies() got %d results, want %d", len(got), tt.wantCount)
			}
		})
	}
}
