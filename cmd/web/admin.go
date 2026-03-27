package main

import (
	"net/http"
	"net/url"
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

		// Explicitly check the user's role.
		// We only allow users with the 'admin' role to proceed to administrative functions.
		// This prevents regular users or moderators from accessing the admin dashboard
		// and performing sensitive administrative actions.
		//
		// Why this is important: Even if a user is authenticated, they should only have access
		// to resources that their role permits (Principle of Least Privilege).
		if user.Role != "admin" {
			// Security Audit: Log unauthorized access attempts to the admin area using structured logging.
			// This helps administrators monitor for potential malicious behavior.
			app.logger.Warn("SECURITY: Unauthorized admin access attempt",
				"userID", user.ID,
				"username", user.Username,
				"path", r.URL.Path,
			)

			// If the user is authenticated but doesn't have the required role,
			// we return a 403 Forbidden status code.
			app.clientError(w, http.StatusForbidden)
			return
		}

		// If the user is an admin, we call the next handler in the chain.
		next.ServeHTTP(w, r)
	}
}

// adminDashboardView renders the global management interface
func (app *application) adminDashboardView(w http.ResponseWriter, r *http.Request) {

	data := app.getTemplateData("Admin Dashboard", r)

	// Fast counting logic for MVP metrics from service layer
	userCount, mediaCount, err := app.Service.GetAdminMetrics()
	if err != nil {
		app.logger.Error("Error fetching admin metrics", "error", err)
	}

	data.UserCount = userCount
	data.MediaCount = mediaCount

	app.render(w, r, http.StatusOK, "admin.html", data)
}

