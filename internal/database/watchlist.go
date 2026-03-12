package database

import (
	"movieweb/internal/models"
)

// GetUserWatchlists fetches all watchlists for a user
func GetUserWatchlists(userID int) ([]models.Watchlist, error) {
	query := `SELECT id, user_id, name, description, is_public, created_at FROM watchlist WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var watchlists []models.Watchlist
	for rows.Next() {
		var w models.Watchlist
		err := rows.Scan(&w.ID, &w.UserID, &w.Name, &w.Description, &w.IsPublic, &w.CreatedAt)
		if err != nil {
			return nil, err
		}
		watchlists = append(watchlists, w)
	}
	return watchlists, nil
}

// CreateWatchlist creates a new custom watchlist
func CreateWatchlist(userID int, name string, description string) error {
	query := `INSERT INTO watchlist (user_id, name, description, is_public) VALUES (?, ?, ?, ?)`
	_, err := DB.Exec(query, userID, name, description, false)
	return err
}

// AddToWatchlist appends media to a specific watchlist
func AddToWatchlist(watchlistID int, mediaType string, mediaID int) error {
	query := `INSERT INTO watchlist_item (watchlist_id, media_type, media_id) VALUES (?, ?, ?)`
	_, err := DB.Exec(query, watchlistID, mediaType, mediaID)
	return err
}
