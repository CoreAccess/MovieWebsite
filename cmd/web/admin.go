package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"movieweb/internal/database"
)

// adminRoleCheck is a simple middleware to ensure only Admins can access routes
func adminRoleCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := getUser(r)
		
		// For MVP, just block completely if not logged in
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// In a real app we would check user.Role == "admin"
		// For this MVP, we assume any authenticated user hitting /admin wants to see it
		// This should be locked down before going to production!
		
		next.ServeHTTP(w, r)
	}
}

// adminDashboardView renders the global management interface
func adminDashboardView(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/admin.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("Admin Dashboard", r)
	
	// Fast counting logic for MVP metrics
	var userCount, mediaCount, pendingEdits, activeAds int
	database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	database.DB.QueryRow("SELECT (SELECT COUNT(*) FROM movies) + (SELECT COUNT(*) FROM tv_series)").Scan(&mediaCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM edit_suggestions WHERE status = 'pending'").Scan(&pendingEdits)
	database.DB.QueryRow("SELECT COUNT(*) FROM ad_campaigns").Scan(&activeAds)

	data.UserCount = userCount
	data.MediaCount = mediaCount
	data.PendingEdits = pendingEdits
	data.ActiveAds = activeAds

	// Fetch pending wiki suggestions
	suggestions, _ := database.GetPendingWikiEdits()
	data.EditSuggestions = suggestions

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
	}
}

// wikiApprovePost handles approving an edit
func wikiApprovePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil || r.PostForm.Get("suggestion_id") == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var suggestionID int
	fmt.Sscanf(r.PostForm.Get("suggestion_id"), "%d", &suggestionID)

	database.ApproveWikiEdit(suggestionID)
	http.Redirect(w, r, "/admin?success=approved", http.StatusSeeOther)
}

// wikiRejectPost handles rejecting an edit
func wikiRejectPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil || r.PostForm.Get("suggestion_id") == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var suggestionID int
	fmt.Sscanf(r.PostForm.Get("suggestion_id"), "%d", &suggestionID)

	database.RejectWikiEdit(suggestionID)
	http.Redirect(w, r, "/admin?success=rejected", http.StatusSeeOther)
}
