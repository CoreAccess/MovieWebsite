package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"

	"movieweb/internal/database"
	"movieweb/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func signupView(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/signup.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("Sign Up", r)
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
	}
}

// signupPost handles the HTTP POST request to create a new user account.
// It parses form data, hashes the password for security, and saves the user to the database.
func signupPost(w http.ResponseWriter, r *http.Request) {
	// Parse the incoming form submission to make form fields available
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Extract form values
	username := r.PostForm.Get("username")
	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	// Hashes the user's password using the bcrypt algorithm with a cost of 12.
	// We never store plain text passwords in the database to prevent exposure if the DB is compromised.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Insert the new user record into the database with the hashed password
	err = database.CreateUser(username, email, string(hash))
	if err != nil {
		// If creation fails (e.g., username or email already exists due to UNIQUE constraints),
		// redirect back to the signup page with an error flag.
		http.Redirect(w, r, "/signup?error=1", http.StatusSeeOther)
		return
	}

	// On successful account creation, redirect the user to the login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func loginView(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/login.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("Login", r)
	data.Next = r.URL.Query().Get("next")
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
	}
}

// loginPost processes user login attempts. It verifies credentials against the database,
// establishes a session, and sets a secure cookie on the client's browser.
func loginPost(w http.ResponseWriter, r *http.Request) {
	// Parse the login form submission
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	// Attempt to find a matching user by their email address
	user, err := database.GetUserByEmail(email)
	if err != nil {
		// User not found
		http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
		return
	}

	// Compare the provided password with the hashed password stored in the database.
	// If they match, the user is authenticated.
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// Passwords do not match
		http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
		return
	}

	// Create a new Session
	// Generate a unique identifier (UUID) for the session to prevent session hijacking
	sessionID := uuid.New().String()
	session := models.Session{
		ID:        sessionID,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Set the session to expire in 24 hours
	}

	// Store the session in the database so the server can track active logins
	err = database.CreateSession(session)
	if err != nil {
		log.Println(err)
		http.Error(w, "Server Error", 500)
		return
	}

	// Send the session ID back to the user's browser in a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,   // The UUID we just generated
		Path:     "/",         // The cookie is valid for the entire site
		Expires:  session.ExpiresAt,
		HttpOnly: true,        // Mitigates XSS attacks (client-side scripts cannot access the cookie)
		Secure:   false,       // Set to true in production if using HTTPS
		SameSite: http.SameSiteStrictMode, // Mitigates Cross-Site Request Forgery (CSRF)
	})

	// Check if the user was trying to access a protected page before logging in.
	// If 'next' is present and valid (starts with '/'), redirect them there.
	nextURL := r.PostForm.Get("next")
	if nextURL != "" && nextURL[0] == '/' {
		http.Redirect(w, r, nextURL, http.StatusSeeOther)
		return
	}

	// Default redirect upon successful login is the homepage
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func logoutPost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		database.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// sessionMiddleware is an HTTP middleware function that intercepts incoming requests
// to determine if the user is authenticated. It checks for a session cookie and adds user data
// to the request context if a valid session exists.
func sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Attempt to read the 'session' cookie from the incoming request
		cookie, err := r.Cookie("session")
		if err == nil && cookie.Value != "" {
			// Look up the session ID in the database
			session, err := database.GetSession(cookie.Value)

			// Verify the session exists and hasn't expired
			if err == nil && time.Now().Before(session.ExpiresAt) {
				// Retrieve the associated user from the database
				user, err := database.GetUserByID(session.UserID)
				if err == nil {
					// Attach the user object to the request context.
					// This allows subsequent handlers (like 'profileView') to access the logged-in user's data
					// by calling `r.Context().Value("user")`.
					ctx := context.WithValue(r.Context(), "user", user)
					r = r.WithContext(ctx)
				}
			} else {
				// If the session is invalid or expired, delete it from the database to clean up
				database.DeleteSession(cookie.Value)
			}
		}
		// Pass control to the next handler in the chain (with the potentially updated context)
		next.ServeHTTP(w, r)
	})
}

// Helper to get user from request context
func getUser(r *http.Request) *models.User {
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		return nil
	}
	return &user
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := getUser(r)
		if user == nil {
			nextParam := ""
			if r.Method == "GET" && r.URL.Path != "" {
				nextParam = "?next=" + r.URL.Path
			}
			http.Redirect(w, r, "/login"+nextParam, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}
