package main

import (
	"fmt"
	"log"
	"net/http"

	"movieweb/internal/database"
)

// wikiEditView renders the form to submit an edit
func (app *application) wikiEditView(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("type")
	entityID := r.URL.Query().Get("id")

	if entityType == "" || entityID == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := app.getTemplateData("Edit Page", r)

	// Add context for the form template
	data.EntityType = entityType
	fmt.Sscanf(entityID, "%d", &data.EntityID)

	app.render(w, http.StatusOK, "wiki_edit.html", data)
}

// wikiEditPost handles the actual database insertion into moderation queue
func (app *application) wikiEditPost(w http.ResponseWriter, r *http.Request) {
	user := app.getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	entityType := r.PostForm.Get("entity_type")
	entityIDStr := r.PostForm.Get("entity_id")
	changes := r.PostForm.Get("changes")

	var entityID int
	fmt.Sscanf(entityIDStr, "%d", &entityID)

	if entityType == "" || entityID == 0 || changes == "" {
		http.Error(w, "Invalid inputs", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO edit_suggestions (user_id, entity_type, entity_id, suggested_data) VALUES (?, ?, ?, ?)`
	_, err = database.DB.Exec(query, user.ID, entityType, entityID, changes)
	if err != nil {
		log.Println("Error submitting edit:", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Redirect back where they came from
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}
	http.Redirect(w, r, referer+"?success=edit_submitted", http.StatusSeeOther)
}
