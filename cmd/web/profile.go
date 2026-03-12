package main

import (
	"fmt"
	"html/template"
	"log"
	"movieweb/internal/database"
	"net/http"
)

// profileView renders the user dashboard
func profileView(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/profile.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("My Profile", r)
	user := getUser(r)
	if user != nil {
		watchlists, err := database.GetUserWatchlists(user.ID)
		if err == nil {
			data.Watchlists = watchlists
		}
	}
	
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
	}
}

// profileEditPost handles updating user settings
func profileEditPost(w http.ResponseWriter, r *http.Request) {
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

	email := r.PostForm.Get("email")
	avatar := r.PostForm.Get("avatarUrl")

	// Update the database
	err = database.UpdateUserProfile(user.ID, email, avatar)
	if err != nil {
		log.Println("Error updating profile:", err)
		http.Redirect(w, r, "/profile?error=update_failed", http.StatusSeeOther)
		return
	}

	// Stay on the profile page
	http.Redirect(w, r, "/profile?success=1", http.StatusSeeOther)
}

// toggleWatchlistPost handles adding logic
func toggleWatchlistPost(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil || r.PostForm.Get("media_id") == "" || r.PostForm.Get("media_type") == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Simplification: Auto-create a default watchlist if the user has none
	watchlists, err := database.GetUserWatchlists(user.ID)
	var watchlistID int
	if err != nil || len(watchlists) == 0 {
		database.CreateWatchlist(user.ID, "My Watchlist", "Stuff I want to watch")
		// Fetch again to get the new ID
		watchlists, _ = database.GetUserWatchlists(user.ID)
	}
	if len(watchlists) > 0 {
		watchlistID = watchlists[0].ID
	} else {
		http.Error(w, "Could not create or find watchlist", http.StatusInternalServerError)
		return
	}

	mediaIDStr := r.PostForm.Get("media_id")
	var mediaID int
	fmt.Sscanf(mediaIDStr, "%d", &mediaID)
	mediaType := r.PostForm.Get("media_type")

	err = database.AddToWatchlist(watchlistID, mediaType, mediaID)
	if err != nil {
		log.Println("Error adding to watchlist:", err)
	}

	// Redirect back where they came from
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/profile"
	}
	http.Redirect(w, r, referer, http.StatusSeeOther)
}
