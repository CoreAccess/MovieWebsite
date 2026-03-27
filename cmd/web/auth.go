package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"filmgap/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) signupView(w http.ResponseWriter, r *http.Request) {
	data := app.getTemplateData("Sign Up", r)
	app.render(w, r, http.StatusOK, "signup.html", data)
}

// signupPost handles the HTTP POST request to create a new user account.
func (app *application) signupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	username := r.PostForm.Get("username")
	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.Service.CreateUser(username, email, string(hash))
	if err != nil {
		// Duplicate username/email — redirect with error flag (no server error needed)
		http.Redirect(w, r, "/signup?error=1", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *application) loginView(w http.ResponseWriter, r *http.Request) {
	data := app.getTemplateData("Login", r)
	data.Next = r.URL.Query().Get("next")
	app.render(w, r, http.StatusOK, "login.html", data)
}

// loginPost processes user login attempts.
func (app *application) loginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	user, err := app.Service.GetUserByEmail(email)
	if err != nil {
		// User not found — redirect with error flag
		http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// Wrong password — redirect with error flag
		http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
		return
	}

	sessionID := uuid.New().String()
	session := models.Session{
		ID:        sessionID,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err = app.Service.CreateSession(session)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	nextURL := r.PostForm.Get("next")
	if isSafeRedirect(nextURL) {
		http.Redirect(w, r, nextURL, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) logoutPost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		app.Service.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// sessionMiddleware resolves the session cookie and attaches the authenticated user to context.
func (app *application) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err == nil && cookie.Value != "" {
			session, err := app.Service.GetSession(cookie.Value)

			if err == nil && time.Now().Before(session.ExpiresAt) {
				user, err := app.Service.GetUserByID(session.UserID)
				if err == nil {
					ctx := context.WithValue(r.Context(), "user", user)
					r = r.WithContext(ctx)
				}
			} else {
				app.Service.DeleteSession(cookie.Value)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// getUser retrieves the authenticated user from the request context.
func (app *application) getUser(r *http.Request) *models.User {
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		return nil
	}
	return &user
}

func (app *application) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := app.getUser(r)
		if user == nil {
			if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintf(w, `{"error": "Authentication required", "login_url": "/login"}`)
				return
			}

			nextParam := ""
			if r.Method == "GET" && r.URL.Path != "" {
				nextParam = "?next=" + url.QueryEscape(r.URL.Path)
			}
			http.Redirect(w, r, "/login"+nextParam, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}
