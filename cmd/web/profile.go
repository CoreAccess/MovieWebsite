package main

import (
	"net/http"
	"strconv"
)

// profileView renders the user dashboard
func (app *application) profileView(w http.ResponseWriter, r *http.Request) {

	data := app.getTemplateData("My Profile", r)
	user := app.getUser(r)
	if user != nil {
		followers, following, err := app.Service.GetFollowCounts(user.ID)
		if err == nil {
			user.FollowerCount = followers
			user.FollowingCount = following
		}
		data.AuthenticatedUser = user

		watchlists, err := app.Service.GetUserWatchlists(user.ID)
		if err == nil {
			data.Watchlists = watchlists
		}
	}

	app.render(w, r, http.StatusOK, "profile.html", data)
}

func (app *application) publicProfileView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		app.notFound(w)
		return
	}

	targetUser, err := app.Service.GetUserByID(id)
	if err != nil {
		app.notFound(w)
		return
	}

	data := app.getTemplateData(targetUser.Username+"'s Profile", r)
	data.ProfileUser = targetUser

	// Fetch follow counts
	followers, following, err := app.Service.GetFollowCounts(id)
	if err == nil {
		data.ProfileUser.FollowerCount = followers
		data.ProfileUser.FollowingCount = following
	}

	// If authenticated, check follow status
	currUser := app.getUser(r)
	if currUser != nil {
		isFollowing, _ := app.Service.IsFollowingUser(currUser.ID, id)
		data.IsFollowing = isFollowing
	}

	app.render(w, r, http.StatusOK, "profile.html", data)
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
	err = app.Service.UpdateUserProfile(user.ID, email, avatar)
	if err != nil {
		app.logger.Error("error updating profile", "error", err, "userID", user.ID)
		http.Redirect(w, r, "/profile?error=update_failed", http.StatusSeeOther)
		return
	}

	// Stay on the profile page
	http.Redirect(w, r, "/profile?success=1", http.StatusSeeOther)
}



