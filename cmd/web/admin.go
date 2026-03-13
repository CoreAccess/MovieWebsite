package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"movieweb/internal/database"
)

// adminRoleCheck is a middleware that restricts access to admin-only routes.
// It retrieves the authenticated user from the request context and verifies their role.
func adminRoleCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Data Flow Trace:
		// 1. sessionMiddleware (in cmd/web/auth.go) extracts the session cookie.
		// 2. It looks up the session in the database and retrieves the User.
		// 3. The User object is stored in the request's Context.
		// 4. getUser(r) retrieves that User object from the Context.
		user := getUser(r)

		// If no user is found in the context, it means the request is unauthenticated.
		// We redirect them to the login page to establish a session.
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Security Fix: Explicitly check the user's role.
		// We only allow users with the 'admin' role to proceed to administrative functions.
		// This prevents regular users or moderators from accessing the admin dashboard
		// and performing sensitive actions like approving/rejecting wiki edits.
		if user.Role != "admin" {
			// If the user is authenticated but doesn't have the required role,
			// we return a 403 Forbidden status code.
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}

		// If the user is an admin, we call the next handler in the chain.
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
