package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"movieweb/internal/database"
)

// wikiEditView renders the form to submit an edit
func wikiEditView(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("type")
	entityID := r.URL.Query().Get("id")
	
	if entityType == "" || entityID == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/wiki_edit.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("Edit Page", r)
	
	// Add context for the form template
	data.EntityType = entityType
	fmt.Sscanf(entityID, "%d", &data.EntityID)
	
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
	}
}

// wikiEditPost handles the actual database insertion into moderation queue
func wikiEditPost(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
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
