package database

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// setupBenchDB creates an in-memory database and the tv_cast table.
func setupCastBenchDB(b *testing.B) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		b.Fatalf("failed to open memory db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tv_cast (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			series_id INTEGER NOT NULL,
			person_id INTEGER NOT NULL,
			character_id INTEGER NOT NULL,
			billing_order INTEGER
		);
	`)
	if err != nil {
		b.Fatalf("failed to create table: %v", err)
	}

	return db
}

// mockCastData generates mock data for the benchmark.
func mockCastData(num int) [][4]int64 {
	data := make([][4]int64, num)
	for i := 0; i < num; i++ {
		// seriesID, personID, characterID, billingOrder
		data[i] = [4]int64{1, int64(i + 1), int64(i + 100), int64(i)}
	}
	return data
}

// BenchmarkTVCastInsert_NPlus1 measures the current approach.
func BenchmarkTVCastInsert_NPlus1(b *testing.B) {
	db := setupCastBenchDB(b)
	defer db.Close()

	numRecords := 50
	data := mockCastData(numRecords)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// We can't wrap the whole benchmark in a transaction easily since the original code doesn't use one,
		// but since SQLite defaults to auto-commit per statement, this accurately simulates the original N+1 behaviour.
		// However, to avoid filling up the DB too much and slowing down purely from size, we truncate occasionally.
		b.StopTimer()
		db.Exec("DELETE FROM tv_cast")
		b.StartTimer()

		for _, row := range data {
			_, err := db.Exec("INSERT INTO tv_cast (series_id, person_id, character_id, billing_order) VALUES (?, ?, ?, ?)", row[0], row[1], row[2], row[3])
			if err != nil {
				b.Fatalf("insert failed: %v", err)
			}
		}
	}
}

// BenchmarkTVCastInsert_Batched measures the optimized chunked approach.
func BenchmarkTVCastInsert_Batched(b *testing.B) {
	db := setupCastBenchDB(b)
	defer db.Close()

	numRecords := 50
	data := mockCastData(numRecords)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		db.Exec("DELETE FROM tv_cast")
		b.StartTimer()

		// Simplified batched insert simulation
		chunkSize := 250
		for j := 0; j < len(data); j += chunkSize {
			end := j + chunkSize
			if end > len(data) {
				end = len(data)
			}

			chunk := data[j:end]
			placeholders := make([]string, len(chunk))
			args := make([]interface{}, len(chunk)*4)

			for k, row := range chunk {
				placeholders[k] = "(?, ?, ?, ?)"
				args[k*4] = row[0]
				args[k*4+1] = row[1]
				args[k*4+2] = row[2]
				args[k*4+3] = row[3]
			}

			query := fmt.Sprintf("INSERT INTO tv_cast (series_id, person_id, character_id, billing_order) VALUES %s", strings.Join(placeholders, ", "))
			_, err := db.Exec(query, args...)
			if err != nil {
				b.Fatalf("batched insert failed: %v", err)
			}
		}
	}
}
