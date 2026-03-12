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

func signupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	username := r.PostForm.Get("username")
	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = database.CreateUser(username, email, string(hash))
	if err != nil {
		http.Redirect(w, r, "/signup?error=1", http.StatusSeeOther)
		return
	}

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

func loginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	user, err := database.GetUserByEmail(email)
	if err != nil {
		http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
		return
	}

	// Create Session
	sessionID := uuid.New().String()
	session := models.Session{
		ID:        sessionID,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err = database.CreateSession(session)
	if err != nil {
		log.Println(err)
		http.Error(w, "Server Error", 500)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   false, // Change correctly in prod
		SameSite: http.SameSiteStrictMode,
	})

	nextURL := r.PostForm.Get("next")
	if nextURL != "" && nextURL[0] == '/' {
		http.Redirect(w, r, nextURL, http.StatusSeeOther)
		return
	}

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

// Middleware
func sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err == nil && cookie.Value != "" {
			session, err := database.GetSession(cookie.Value)
			if err == nil && time.Now().Before(session.ExpiresAt) {
				user, err := database.GetUserByID(session.UserID)
				if err == nil {
					// Attach user to context
					ctx := context.WithValue(r.Context(), "user", user)
					r = r.WithContext(ctx)
				}
			} else {
				// Invalid or expired session - Clear it
				database.DeleteSession(cookie.Value)
			}
		}
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
