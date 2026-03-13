package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"movieweb/internal/database"
)

// adminRoleCheck is a middleware that restricts access to admin-only routes.
// It retrieves the authenticated user from the request context and verifies their role.
func (app *application) adminRoleCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Data Flow Trace:
		// 1. sessionMiddleware (in cmd/web/auth.go) extracts the session cookie.
		// 2. It looks up the session in the database and retrieves the User.
		// 3. The User object is stored in the request's Context.
		// 4. app.getUser(r) retrieves that User object from the Context.
		user := app.getUser(r)

		// If no user is found in the context, it means the request is unauthenticated.
		// We redirect them to the login page to establish a session.
		// Pedagogical Note: We include a 'next' parameter so that after a successful login,
		// the user can be automatically redirected back to the page they were trying to access.
		if user == nil {
			nextParam := ""
			if r.Method == "GET" && r.URL.Path != "" {
				nextParam = "?next=" + url.QueryEscape(r.URL.Path)
			}
			http.Redirect(w, r, "/login"+nextParam, http.StatusSeeOther)
			return
		}

		// Security Fix: Explicitly check the user's role.
		// We only allow users with the 'admin' role to proceed to administrative functions.
		// This prevents regular users or moderators from accessing the admin dashboard
		// and performing sensitive actions like approving/rejecting wiki edits.
		//
		// Why this is important: Even if a user is authenticated, they should only have access
		// to resources that their role permits (Principle of Least Privilege).
		if user.Role != "admin" {
			// Security Audit: Log unauthorized access attempts to the admin area.
			// This helps administrators monitor for potential malicious behavior.
			log.Printf("SECURITY: Unauthorized admin access attempt by User ID %d (%s) on %s", user.ID, user.Username, r.URL.Path)

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
func (app *application) adminDashboardView(w http.ResponseWriter, r *http.Request) {

	data := app.getTemplateData("Admin Dashboard", r)

	// Fast counting logic for MVP metrics
	var userCount, mediaCount, pendingEdits, activeAds int
	query := `
		SELECT
			(SELECT COUNT(*) FROM users),
			(SELECT COUNT(*) FROM movies) + (SELECT COUNT(*) FROM tv_series),
			(SELECT COUNT(*) FROM edit_suggestions WHERE status = 'pending'),
			(SELECT COUNT(*) FROM ad_campaigns)
	`
	database.DB.QueryRow(query).Scan(&userCount, &mediaCount, &pendingEdits, &activeAds)

	data.UserCount = userCount
	data.MediaCount = mediaCount
	data.PendingEdits = pendingEdits
	data.ActiveAds = activeAds

	// Fetch pending wiki suggestions
	suggestions, _ := database.GetPendingWikiEdits()
	data.EditSuggestions = suggestions

	app.render(w, http.StatusOK, "admin.html", data)
}

// wikiApprovePost handles approving an edit
func (app *application) wikiApprovePost(w http.ResponseWriter, r *http.Request) {
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
func (app *application) wikiRejectPost(w http.ResponseWriter, r *http.Request) {
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
