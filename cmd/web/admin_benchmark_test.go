package main

import (
	"testing"
	"movieweb/internal/database"
)

func BenchmarkAdminDashboardMetricsQuery(b *testing.B) {
	// Initialize an in-memory database for testing
	_, err := database.InitDB(":memory:", "")
	if err != nil {
		b.Fatalf("Failed to initialize database: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var userCount, mediaCount, pendingEdits, activeAds int
		database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
		database.DB.QueryRow("SELECT (SELECT COUNT(*) FROM movies) + (SELECT COUNT(*) FROM tv_series)").Scan(&mediaCount)
		database.DB.QueryRow("SELECT COUNT(*) FROM edit_suggestions WHERE status = 'pending'").Scan(&pendingEdits)
		database.DB.QueryRow("SELECT COUNT(*) FROM ad_campaigns").Scan(&activeAds)
	}
}

func BenchmarkAdminDashboardMetricsQueryCombined(b *testing.B) {
	// Initialize an in-memory database for testing
	_, err := database.InitDB(":memory:", "")
	if err != nil {
		b.Fatalf("Failed to initialize database: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var userCount, mediaCount, pendingEdits, activeAds int
		query := `
			SELECT
				(SELECT COUNT(*) FROM users),
				(SELECT COUNT(*) FROM movies) + (SELECT COUNT(*) FROM tv_series),
				(SELECT COUNT(*) FROM edit_suggestions WHERE status = 'pending'),
				(SELECT COUNT(*) FROM ad_campaigns)
		`
		database.DB.QueryRow(query).Scan(&userCount, &mediaCount, &pendingEdits, &activeAds)
	}
}
