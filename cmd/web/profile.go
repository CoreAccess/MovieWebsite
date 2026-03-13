package main

import (
	"fmt"
	"log"
	"movieweb/internal/database"
	"net/http"
)

// profileView renders the user dashboard
func (app *application) profileView(w http.ResponseWriter, r *http.Request) {

	data := app.getTemplateData("My Profile", r)
	user := app.getUser(r)
	if user != nil {
		watchlists, err := database.GetUserWatchlists(user.ID)
		if err == nil {
			data.Watchlists = watchlists
		}
	}

	app.render(w, http.StatusOK, "profile.html", data)
}

// profileEditPost handles updating user settings
func (app *application) profileEditPost(w http.ResponseWriter, r *http.Request) {
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
func (app *application) toggleWatchlistPost(w http.ResponseWriter, r *http.Request) {
	user := app.getUser(r)
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
	referer := getSafeReferer(r, "/profile")
	http.Redirect(w, r, referer, http.StatusSeeOther)
}
