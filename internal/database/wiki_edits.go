package database

import (
	"log"
	"time"
)

// EditSuggestion represents a proposed change to a media entity
type EditSuggestion struct {
	ID            int
	UserID        int // Submitter ID
	EntityType    string
	EntityID      int
	SuggestedData string
	Status        string
	CreatedAt     time.Time
}

// GetPendingWikiEdits fetches all wiki edits that are awaiting moderation
func GetPendingWikiEdits() ([]EditSuggestion, error) {
	var suggestions []EditSuggestion

	query := `
		SELECT 
			id, user_id, entity_type, entity_id, suggested_data, status, created_at 
		FROM edit_suggestions 
		WHERE status = 'pending'
		ORDER BY created_at DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		log.Println("Error fetching pending edits:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var es EditSuggestion
		err := rows.Scan(
			&es.ID,
			&es.UserID,
			&es.EntityType,
			&es.EntityID,
			&es.SuggestedData,
			&es.Status,
			&es.CreatedAt,
		)
		if err != nil {
			log.Println("Error scanning edit suggestion row:", err)
			continue
		}
		suggestions = append(suggestions, es)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return suggestions, nil
}
