package main

import (
	"log"
	"net/http"
	"strconv"

	"movieweb/internal/database"
	"movieweb/internal/models"
	"movieweb/internal/monetization"
)

func (app *application) movieView(w http.ResponseWriter, r *http.Request) {
	// Extract the 'id' and 'slug' path values from the URL, defined in the route: /movies/{id}/{slug}
	idStr := r.PathValue("id")
	slug := r.PathValue("slug")

	// Convert the ID string to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil || slug == "" {
		// If the ID is invalid or slug is missing, return a 404 Not Found error
		app.notFound(w)
		return
	}

	// Define the slice of HTML template files needed to render the page.
	// The order typically includes the base layout, partials like nav/sidebar, and the specific page template.

	// We use this to convert our HTML files into executable templates.

	// Fetch the complete movie details (including cast, crew, etc.) from the database layer.
	detail, err := database.GetMovieDetail(id)
	if err != nil {
		log.Println("Error fetching movie details:", err)
		app.notFound(w) // Show a 404 if the movie doesn't exist in the DB
		return
	}

	// Prepare the template data structure. getTemplateData gives us common data (like nav info).
	data := app.getTemplateData(detail.Movie.Name, r)
	data.MovieDetail = detail
	// Fetch real-time affiliate listings for monetization based on the movie's name.
	data.EbayListings = monetization.FetchEbayListings(detail.Movie.Name)

	// Execute the "base" template, passing in our data structure.
	// This generates the final HTML and writes it to the http.ResponseWriter (w) to send to the client.
	app.render(w, http.StatusOK, "movies.html", data)
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	data := app.getTemplateData("Home", r)
	data.Ads = monetization.FetchAdvertisements("home")

	// Fetch data for guest view, which is the default view for the home page
	popularMovies, _ := database.GetPopularMovies(10)
	data.PopularMovies = popularMovies

	upcomingMovies, _ := database.GetUpcomingMovies(10)
	data.UpcomingMovies = upcomingMovies

	popularShows, _ := database.GetPopularShows(10)
	data.PopularShows = popularShows

	newShows, _ := database.GetNewShows(10)
	data.NewShows = newShows

	app.render(w, http.StatusOK, "index.html", data)
}

func (app *application) myFeedView(w http.ResponseWriter, r *http.Request) {

	data := app.getTemplateData("My Feed", r)
	data.Ads = monetization.FetchAdvertisements("home")

	app.render(w, http.StatusOK, "my-feed.html", data)
}

func (app *application) tvView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	slug := r.PathValue("slug")
	id, err := strconv.Atoi(idStr)
	if err != nil || slug == "" {
		app.notFound(w)
		return
	}

	detail, err := database.GetTVSeriesDetail(id)
	if err != nil {
		log.Println("Error fetching series details:", err)
		app.notFound(w)
		return
	}

	data := app.getTemplateData(detail.Series.Name, r)
	data.TVSeriesDetail = detail
	data.EbayListings = monetization.FetchEbayListings(detail.Series.Name)

	app.render(w, http.StatusOK, "tv_shows.html", data)
}

func (app *application) personView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	slug := r.PathValue("slug")
	id, err := strconv.Atoi(idStr)
	if err != nil || slug == "" {
		app.notFound(w)
		return
	}

	detail, err := database.GetPersonDetailByID(id)
	if err != nil {
		log.Println("Error fetching person details:", err)
		app.notFound(w)
		return
	}

	data := app.getTemplateData(detail.Person.Name, r)
	data.PersonDetail = detail

	app.render(w, http.StatusOK, "people.html", data)
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

	movies, err := database.GetAllMovies(limit, offset, sortParam)
	if err == nil {
		data.Movies = movies
	}

	app.render(w, http.StatusOK, "movies-list.html", data)
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

	shows, err := database.GetAllShows(limit, offset, sortParam)
	if err == nil {
		data.Shows = shows
	}

	app.render(w, http.StatusOK, "tv-shows-list.html", data)
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

	people, err := database.GetAllPeople(limit, offset, sortParam)
	if err == nil {
		data.People = people
	}

	app.render(w, http.StatusOK, "people-list.html", data)
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
			movies, err := database.SearchMovies(q, limit, offset)
			if err == nil {
				data.Movies = movies
			}
		} else {
			movies, err := database.GetAllMovies(limit, offset, "")
			if err == nil {
				data.Movies = movies
			}
		}
	}

	if filter == "tv" || filter == "" {
		if q != "" {
			shows, err := database.SearchShows(q, limit, offset)
			if err == nil {
				data.Shows = shows
			}
		} else {
			shows, err := database.GetAllShows(limit, offset, "")
			if err == nil {
				data.Shows = shows
			}
		}
	}

	if filter == "people" || filter == "" {
		if q != "" {
			people, err := database.SearchPeople(q, limit, offset)
			if err == nil {
				data.People = people
			}
		} else {
			people, err := database.GetAllPeople(limit, offset, "")
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

	app.render(w, http.StatusOK, "search.html", data)
}

func (app *application) watchlistView(w http.ResponseWriter, r *http.Request) {

	data := app.getTemplateData("My Watchlist", r)

	movies, shows, err := database.GetUserWatchlist(data.AuthenticatedUser.ID)
	if err != nil {
		log.Println("Error fetching watchlist:", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	data.Movies = movies
	data.Shows = shows

	app.render(w, http.StatusOK, "watchlist.html", data)
}

func (app *application) aboutView(w http.ResponseWriter, r *http.Request) {

	data := app.getTemplateData("About", r)

	app.render(w, http.StatusOK, "about.html", data)
}

func (app *application) contactView(w http.ResponseWriter, r *http.Request) {

	data := app.getTemplateData("Contact", r)

	app.render(w, http.StatusOK, "contact.html", data)
}

func (app *application) termsView(w http.ResponseWriter, r *http.Request) {

	data := app.getTemplateData("Terms of Service", r)

	app.render(w, http.StatusOK, "terms.html", data)
}

func (app *application) privacyView(w http.ResponseWriter, r *http.Request) {

	data := app.getTemplateData("Privacy Policy", r)

	app.render(w, http.StatusOK, "privacy.html", data)
}
