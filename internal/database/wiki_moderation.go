package database

import (
	"log"
)

// ApproveWikiEdit marks the suggestion as approved and handles basic points.
// Note: MVP does not truly alter the underlying struct, it just marks it approved.
func ApproveWikiEdit(suggestionID int) error {
	_, err := DB.Exec("UPDATE edit_suggestions SET status = 'approved' WHERE id = ?", suggestionID)
	if err != nil {
		log.Println("Error approving wiki edit:", err)
		return err
	}
	return nil
}

// RejectWikiEdit marks the suggestion as rejected.
func RejectWikiEdit(suggestionID int) error {
	_, err := DB.Exec("UPDATE edit_suggestions SET status = 'rejected' WHERE id = ?", suggestionID)
	if err != nil {
		log.Println("Error rejecting wiki edit:", err)
		return err
	}
	return nil
}
