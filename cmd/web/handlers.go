package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"filmgap/internal/models"
)

func (app *application) movieView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	slug := r.PathValue("slug")

	id, err := strconv.Atoi(idStr)
	if err != nil || slug == "" {
		app.notFound(w)
		return
	}

	movie, cast, crew, genres, _, err := app.Service.GetMovieDetail(id, r.Host)
	if err != nil {
		app.logger.Error("error fetching movie details", "error", err, "id", id)
		app.notFound(w)
		return
	}

	if movie == nil {
		app.notFound(w)
		return
	}

	var directors, writers []models.Person
	for _, c := range crew {
		if c.Job == "Director" {
			directors = append(directors, c.Person)
		} else if c.Job == "Writer" || c.Job == "Screenplay" {
			writers = append(writers, c.Person)
		}
	}

	detail := models.MovieDetail{
		Movie:     *movie,
		Cast:      cast,
		Directors: directors,
		Writers:   writers,
		Genres:    genres,
	}

	data := app.getTemplateData(detail.Movie.Name, r)
	data.MovieDetail = detail
	watchProviderGroups, err := app.Service.GetWatchProviderGroups(movie.ID, "Movie", movie.TmdbID, movie.CountryCode)
	if err == nil {
		data.WatchProviderGroups = watchProviderGroups
	} else {
		app.logger.Warn("error fetching watch providers", "error", err, "mediaID", movie.ID)
	}

	if data.AuthenticatedUser != nil {
		data.WatchlistMap = app.getWatchlistMap(data.AuthenticatedUser.ID)
	}

	app.render(w, r, http.StatusOK, "movies.html", data)
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	// Fetch comprehensive homepage data from the Service layer
	homepageData, err := app.Service.GetHomepageData(r.Host)
	if err != nil {
		app.logger.Error("error fetching homepage data", "error", err)
		app.serverError(w, r, err)
		return
	}

	data := app.getTemplateData("Home", r)

	// Map service data to templateData fields
	if val, ok := homepageData["Hero"]; ok {
		data.Hero = val
	}
	if val, ok := homepageData["Stats"]; ok {
		data.Stats = val.(models.HomepageStats)
	}
	if val, ok := homepageData["Trending"]; ok {
		data.Trending = val.([]models.MediaSummary)
	}
	if val, ok := homepageData["Activity"]; ok {
		data.Activity = val.([]models.Activity)
	}
	if val, ok := homepageData["PopularLists"]; ok {
		data.PopularLists = val.([]models.List)
	}
	if val, ok := homepageData["Franchise"]; ok {
		data.Franchise = val.(models.Franchise)
	}
	if val, ok := homepageData["BlogPosts"]; ok {
		data.BlogPosts = val.([]models.BlogPost)
	}
	if val, ok := homepageData["Photos"]; ok {
		data.Photos = val.([]models.Photo)
	}
	if val, ok := homepageData["Birthdays"]; ok {
		data.Birthdays = val.([]models.Person)
	}
	if val, ok := homepageData["FanFavorites"]; ok {
		data.FanFavorites = val.([]models.MediaSummary)
	}
	if val, ok := homepageData["BoxOffice"]; ok {
		data.BoxOffice = val.([]models.MediaSummary)
	}
	if val, ok := homepageData["PopularCelebs"]; ok {
		data.PopularCelebs = val.([]models.Person)
	}

	if data.AuthenticatedUser != nil {
		data.WatchlistMap = app.getWatchlistMap(data.AuthenticatedUser.ID)
	}

	app.render(w, r, http.StatusOK, "index.html", data)
}

func (app *application) myFeedView(w http.ResponseWriter, r *http.Request) {
	data := app.getTemplateData("My Feed", r)

	if data.AuthenticatedUser != nil {
		activities, err := app.Service.GetFollowerFeed(data.AuthenticatedUser.ID, 20)
		if err == nil {
			data.Activity = activities
		} else {
			app.logger.Error("error fetching follower feed", "error", err, "userID", data.AuthenticatedUser.ID)
		}
	}

	app.render(w, r, http.StatusOK, "my-feed.html", data)
}

func (app *application) notificationsView(w http.ResponseWriter, r *http.Request) {
	data := app.getTemplateData("Notifications", r)
	app.render(w, r, http.StatusOK, "notifications.html", data)
}

func (app *application) tvView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	slug := r.PathValue("slug")
	id, err := strconv.Atoi(idStr)
	if err != nil || slug == "" {
		app.notFound(w)
		return
	}

	show, cast, crew, episodes, genres, _, err := app.Service.GetShowDetail(id, r.Host)
	if err != nil {
		app.logger.Error("error fetching series details", "error", err, "id", id)
		app.notFound(w)
		return
	}

	if show == nil {
		app.notFound(w)
		return
	}

	var directors []models.Person
	for _, c := range crew {
		if c.Job == "Director" || c.Job == "Executive Producer" || c.Job == "Creator" {
			directors = append(directors, c.Person)
		}
	}

	detail := models.TVSeriesDetail{
		TVSeries:     *show,
		Cast:         cast,
		Directors:    directors,
		Genres:       genres,
		Episodes:     episodes,
		SeasonGroups: buildTVSeasonGroups(episodes),
		Series:       *show,
	}

	data := app.getTemplateData(detail.TVSeries.Name, r)
	data.TVSeriesDetail = detail
	watchProviderGroups, err := app.Service.GetWatchProviderGroups(show.ID, "TVSeries", show.TmdbID, show.CountryCode)
	if err == nil {
		data.WatchProviderGroups = watchProviderGroups
	} else {
		app.logger.Warn("error fetching TV watch providers", "error", err, "mediaID", show.ID)
	}

	if data.AuthenticatedUser != nil {
		data.WatchlistMap = app.getWatchlistMap(data.AuthenticatedUser.ID)
	}

	app.render(w, r, http.StatusOK, "tv_shows.html", data)
}

func (app *application) personView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	slug := r.PathValue("slug")
	id, err := strconv.Atoi(idStr)
	if err != nil || slug == "" {
		app.notFound(w)
		return
	}

	person, movies, shows, _, err := app.Service.GetPersonDetail(id, r.Host)
	if err != nil {
		app.logger.Error("error fetching person details", "error", err, "id", id)
		app.notFound(w)
		return
	}

	if person == nil {
		app.notFound(w)
		return
	}

	// Fetch follow counts
	followers, _, err := app.Service.GetPersonFollowCounts(id)
	if err == nil {
		person.FollowerCount = followers
	}

	detail := models.PersonDetail{
		Person: *person,
		Movies: movies,
		Shows:  shows,
	}

	data := app.getTemplateData(detail.Person.Name, r)
	data.PersonDetail = detail

	// If authenticated, check follow status
	currUser := app.getUser(r)
	if currUser != nil {
		isFollowing, _ := app.Service.IsFollowingPerson(currUser.ID, id)
		data.IsFollowing = isFollowing
	}

	app.render(w, r, http.StatusOK, "people.html", data)
}

func (app *application) moviesListView(w http.ResponseWriter, r *http.Request) {
	data := app.getTemplateData("Movies", r)
	data.Movies = nil
	pageStr := r.URL.Query().Get("page")
	sortParam := r.URL.Query().Get("sort")
	data.Sort = sortParam

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 24
	offset := (page - 1) * limit

	movies, err := app.Service.GetAllMovies(limit, offset, sortParam)
	if err == nil {
		data.Movies = movies
	}

	app.render(w, r, http.StatusOK, "movies-list.html", data)
}

func (app *application) showsListView(w http.ResponseWriter, r *http.Request) {
	data := app.getTemplateData("TV Shows", r)
	data.Shows = nil
	pageStr := r.URL.Query().Get("page")
	sortParam := r.URL.Query().Get("sort")
	data.Sort = sortParam

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 24
	offset := (page - 1) * limit

	shows, err := app.Service.GetAllShows(limit, offset, sortParam)
	if err == nil {
		data.Shows = shows
	}

	app.render(w, r, http.StatusOK, "tv-shows-list.html", data)
}

func (app *application) peopleListView(w http.ResponseWriter, r *http.Request) {
	data := app.getTemplateData("People", r)
	data.People = nil
	pageStr := r.URL.Query().Get("page")
	sortParam := r.URL.Query().Get("sort")
	data.Sort = sortParam

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 24
	offset := (page - 1) * limit

	people, err := app.Service.GetAllPeople(limit, offset, sortParam)
	if err == nil {
		data.People = people
	}

	app.render(w, r, http.StatusOK, "people-list.html", data)
}

func (app *application) searchView(w http.ResponseWriter, r *http.Request) {
	data := app.getTemplateData("Search", r)
	data.Movies = nil
	data.Shows = nil
	data.People = nil
	filter := r.URL.Query().Get("filter")
	pageStr := r.URL.Query().Get("page")
	q := r.URL.Query().Get("q")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	offset := (page - 1) * limit

	if filter == "movies" || filter == "" {
		if q != "" {
			movies, err := app.Service.SearchMovies(q, limit, offset)
			if err == nil {
				data.Movies = movies
			}
		} else {
			movies, err := app.Service.GetAllMovies(limit, offset, "")
			if err == nil {
				data.Movies = movies
			}
		}
	}

	if filter == "tv" || filter == "" {
		if q != "" {
			shows, err := app.Service.SearchShows(q, limit, offset)
			if err == nil {
				data.Shows = shows
			}
		} else {
			shows, err := app.Service.GetAllShows(limit, offset, "")
			if err == nil {
				data.Shows = shows
			}
		}
	}

	if filter == "people" || filter == "" {
		if q != "" {
			people, err := app.Service.SearchPeople(q, limit, offset)
			if err == nil {
				data.People = people
			}
		} else {
			people, err := app.Service.GetAllPeople(limit, offset, "")
			if err == nil {
				data.People = people
			}
		}
	}

	if data.Movies != nil {
		data.ResultCount += len(data.Movies.([]models.Movie))
	}
	if data.Shows != nil {
		data.ResultCount += len(data.Shows.([]models.TVSeries))
	}
	if data.People != nil {
		data.ResultCount += len(data.People.([]models.Person))
	}
	data.SearchQuery = q
	data.Filter = filter

	app.render(w, r, http.StatusOK, "search.html", data)
}

func (app *application) watchlistView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var movies []models.Movie
	var shows []models.TVSeries
	var err error
	var watchlist models.Watchlist

	user := app.getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			app.notFound(w)
			return
		}
		watchlist, movies, shows, err = app.Service.GetWatchlistByID(id)
	} else {
		// Default to primary watchlist
		movies, shows, err = app.Service.GetUserWatchlist(user.ID)
		watchlist.Name = "My Watchlist"
	}

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.getTemplateData(watchlist.Name, r)
	data.Movies = movies
	data.Shows = shows

	app.render(w, r, http.StatusOK, "watchlist.html", data)
}

func (app *application) aboutView(w http.ResponseWriter, r *http.Request) {
	data := app.getTemplateData("About", r)
	app.render(w, r, http.StatusOK, "about.html", data)
}

func (app *application) contactView(w http.ResponseWriter, r *http.Request) {
	data := app.getTemplateData("Contact", r)
	app.render(w, r, http.StatusOK, "contact.html", data)
}

func (app *application) termsView(w http.ResponseWriter, r *http.Request) {
	data := app.getTemplateData("Terms of Service", r)
	app.render(w, r, http.StatusOK, "terms.html", data)
}

func (app *application) rateMedia(w http.ResponseWriter, r *http.Request) {
	user := app.getUser(r)
	if user == nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	mediaID, _ := strconv.Atoi(r.PostForm.Get("media_id"))
	rating, _ := strconv.ParseFloat(r.PostForm.Get("rating"), 64)

	if mediaID == 0 || rating < 1 || rating > 10 {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.Service.SubmitReview(models.Review{
		UserID:     user.ID,
		MediaID:    mediaID,
		Rating:     rating,
		ReviewType: "quick",
	})

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, getSafeReferer(r, "/"), http.StatusSeeOther)
}

func (app *application) reviewPost(w http.ResponseWriter, r *http.Request) {
	user := app.getUser(r)
	if user == nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	mediaID, _ := strconv.Atoi(r.PostForm.Get("media_id"))
	listID, _ := strconv.Atoi(r.PostForm.Get("list_id"))
	rating, _ := strconv.ParseFloat(r.PostForm.Get("rating"), 64)
	content := r.PostForm.Get("content")
	isSpoiler := r.PostForm.Get("is_spoiler") == "true" || r.PostForm.Get("is_spoiler") == "1"

	if mediaID == 0 {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Handle Review/Rating
	if rating > 0 {
		err = app.Service.SubmitReview(models.Review{
			UserID:           user.ID,
			MediaID:          mediaID,
			Rating:           rating,
			Body:             content,
			ContainsSpoilers: isSpoiler,
			ReviewType:       "user",
		})
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}

	// Handle List Addition
	if listID > 0 {
		// Verify list ownership
		list, err := app.Service.GetListByID(listID)
		if err == nil && list.UserID == user.ID {
			_ = app.Service.AddListItem(models.ListItem{
				ListID:  listID,
				MediaID: mediaID,
				AddedBy: user.ID,
				Note:    content, // Optional: use review content as note
			})
		}
	}

	http.Redirect(w, r, getSafeReferer(r, "/"), http.StatusSeeOther)
}

func (app *application) toggleWatchlistPost(w http.ResponseWriter, r *http.Request) {
	user := app.getUser(r)
	if user == nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	mediaID, _ := strconv.Atoi(r.PostForm.Get("media_id"))
	mediaType := r.PostForm.Get("media_type")
	action := r.PostForm.Get("action") // "add" or "remove"

	if mediaID == 0 || (mediaType != "Movie" && mediaType != "TV" && mediaType != "TVSeries") {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if mediaType == "TV" {
		mediaType = "TVSeries"
	}

	add := action != "remove"
	err = app.Service.ToggleWatchlist(user.ID, mediaType, mediaID, add)

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": true, "added": %t}`, add)
		return
	}

	http.Redirect(w, r, getSafeReferer(r, "/"), http.StatusSeeOther)
}

func (app *application) followPost(w http.ResponseWriter, r *http.Request) {
	user := app.getUser(r)
	if user == nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	targetID, _ := strconv.Atoi(r.PostForm.Get("target_id"))
	targetType := r.PostForm.Get("target_type") // "user", "person", "list"

	if targetID == 0 || targetType == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	switch targetType {
	case "user":
		err = app.Service.FollowUser(user.ID, targetID)
	case "person":
		err = app.Service.FollowPerson(user.ID, targetID)
	case "list":
		err = app.Service.FollowList(user.ID, targetID)
	default:
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": true}`)
		return
	}

	http.Redirect(w, r, getSafeReferer(r, "/"), http.StatusSeeOther)
}

func (app *application) unfollowPost(w http.ResponseWriter, r *http.Request) {
	user := app.getUser(r)
	if user == nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	targetID, _ := strconv.Atoi(r.PostForm.Get("target_id"))
	targetType := r.PostForm.Get("target_type")

	if targetID == 0 || targetType == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	switch targetType {
	case "user":
		err = app.Service.UnfollowUser(user.ID, targetID)
	case "person":
		err = app.Service.UnfollowPerson(user.ID, targetID)
	case "list":
		err = app.Service.UnfollowList(user.ID, targetID)
	default:
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": true}`)
		return
	}

	http.Redirect(w, r, getSafeReferer(r, "/"), http.StatusSeeOther)
}

func (app *application) privacyView(w http.ResponseWriter, r *http.Request) {
	data := app.getTemplateData("Privacy Policy", r)
	app.render(w, r, http.StatusOK, "privacy.html", data)
}

func (app *application) trendingAPI(w http.ResponseWriter, r *http.Request) {
	mediaType := r.URL.Query().Get("type") // "Movie", "TV", or "People"

	var data any
	var err error

	if mediaType == "People" {
		people, pErr := app.Service.GetTrendingPeople(12, 0)
		if pErr != nil {
			app.serverError(w, r, pErr)
			return
		}
		// Convert to MediaSummary shape so JS gets consistent field names
		summaries := make([]models.MediaSummary, 0, len(people))
		for _, p := range people {
			age := ""
			if p.BirthDate != "" {
				if bd, parseErr := time.Parse("2006-01-02", p.BirthDate); parseErr == nil {
					age = fmt.Sprintf("%d", int(time.Since(bd).Hours()/8766))
				}
			}
			summaries = append(summaries, models.MediaSummary{
				ID:        p.ID,
				MediaType: "People",
				Name:      p.Name,
				Slug:      p.Slug,
				Image:     p.Image,
				Year:      age,
			})
		}
		data = summaries
	} else {
		// mediaType could be "", "Movie", or "TV"
		// The DB uses "TVSeries" for TV shows. Normalize here.
		if mediaType == "TV" {
			mediaType = "TVSeries"
		}
		data, err = app.Service.GetTrendingMedia(mediaType, 12, 0)
	}

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (app *application) franchiseView(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		app.notFound(w)
		return
	}

	franchise, media, err := app.Service.GetFranchiseDetail(slug)
	if err != nil {
		if err == sql.ErrNoRows {
			app.notFound(w)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.getTemplateData(franchise.Name, r)
	data.Franchise = franchise
	data.Trending = media // Reusing Trending field for the media grid

	app.render(w, r, http.StatusOK, "franchise.html", data)
}

func (app *application) blogListView(w http.ResponseWriter, r *http.Request) {
	data := app.getTemplateData("Blog", r)

	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	offset := (page - 1) * limit

	posts, err := app.Service.GetAllBlogPosts(limit, offset)
	if err != nil {
		app.logger.Error("error fetching blog posts", "error", err)
	}
	data.BlogPosts = posts

	app.render(w, r, http.StatusOK, "blog.html", data)
}

func (app *application) blogPostView(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		app.notFound(w)
		return
	}

	post, jsonLD, err := app.Service.GetBlogPostBySlug(slug, r.Host)
	if err != nil {
		if err == sql.ErrNoRows {
			app.notFound(w)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.getTemplateData(post.Title, r)
	data.JSONLD = jsonLD
	data.BlogPosts = []models.BlogPost{post}

	app.render(w, r, http.StatusOK, "blog-post.html", data)
}
