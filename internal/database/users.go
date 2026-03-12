package database

import (
	"movieweb/internal/models"
)

func CreateUser(username, email, hash string) error {
	_, err := DB.Exec("INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)", username, email, hash)
	return err
}

func GetUserByEmail(email string) (models.User, error) {
	var u models.User
	err := DB.QueryRow("SELECT id, username, email, password_hash, role, reputation_score FROM users WHERE email = ?", email).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.ReputationScore)
	return u, err
}

func CreateSession(s models.Session) error {
	_, err := DB.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", s.ID, s.UserID, s.ExpiresAt)
	return err
}

func GetSession(id string) (models.Session, error) {
	var s models.Session
	err := DB.QueryRow("SELECT id, user_id, expires_at FROM sessions WHERE id = ?", id).
		Scan(&s.ID, &s.UserID, &s.ExpiresAt)
	return s, err
}

// UpdateUserProfile updates a user's email and avatar
func UpdateUserProfile(userID int, email string, avatar string) error {
	query := `UPDATE user SET email = ?, avatar = ? WHERE id = ?`
	_, err := DB.Exec(query, email, avatar, userID)
	return err
}

func DeleteSession(id string) error {
	_, err := DB.Exec("DELETE FROM sessions WHERE id = ?", id)
	return err
}

func GetUserByID(id int) (models.User, error) {
	var u models.User
	err := DB.QueryRow("SELECT id, username, email, COALESCE(role, 'user'), COALESCE(reputation_score, 0), COALESCE(avatar, '') FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.ReputationScore, &u.Avatar)
	return u, err
}
