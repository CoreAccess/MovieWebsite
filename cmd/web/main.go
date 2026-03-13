package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"movieweb/internal/config"
	"movieweb/internal/database"
	"movieweb/internal/models"
	"movieweb/internal/monetization"
	"strings"

	"github.com/justinas/nosurf"
)

type templateData struct {
	Title          string
	Movies         any
	Shows          any
	People         any
	Users          any
	PopularMovies  any
	UpcomingMovies any
	PopularShows   any
	NewShows       any
	MovieDetail    any
	TVSeriesDetail any
	PersonDetail   any
	SearchQuery       string
	ResultCount       int
	Reviews           []any
	AuthenticatedUser *models.User
	CSRFToken         string
	Watchlists        []models.Watchlist
	EbayListings      []models.EbayListing
	Ads               []models.Advertisement
	EntityType        string
	EntityID          int
	UserCount         int
	MediaCount        int
	PendingEdits      int
	ActiveAds         int
	EditSuggestions   []database.EditSuggestion
	AdCampaigns       []database.AdCampaign
	AdsList           []database.Advertisement
	Sort              string
	Next              string
}

// main is the entry point of the application. It initializes the database, sets up the router (mux),
// configures routes for serving static files and handling various HTTP requests, and starts the web server.
func main() {
	// Load environment variables from .env file if it exists
	config.LoadEnv(".env")

	// Initialize the SQLite database and seed it with data from TMDB if empty.
	// We check for TMDB_ACCESS_TOKEN (v4) first, then fallback to TMDB_API_KEY (v3).
	tmdbKey := os.Getenv("TMDB_ACCESS_TOKEN")
	if tmdbKey == "" {
		tmdbKey = os.Getenv("TMDB_API_KEY")
	}

	if tmdbKey == "" {
		tmdbKey = "eyJhbGciOiJIUzI1NiJ9.eyJhdWQiOiJhOWJkZTc1NTdkZTNmNTBiN2FiNzRhODU2MGU0YTc2NCIsIm5iZiI6MTY4ODY3NDU1OC4zOTIsInN1YiI6IjY0YTcyMGZlZjkyNTMyMDE0ZTljNmE4NCIsInNjb3BlcyI6WyJhcGpfcmVhZCJdLCJ2ZXJzaW9uIjoxfQ.8dDf7xLb6lSf1n6TwUgxV3loKu3ieuB0yQw0J4MXCg4"
		log.Println("Note: Using hardcoded TMDB API key. If you see 401 Unauthorized errors, please set the TMDB_ACCESS_TOKEN or TMDB_API_KEY environment variable.")
	}
	
	if _, err := database.InitDB("./streamline.db", tmdbKey); err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err) // Log fatal error if DB initialization fails
	}

	// http.NewServeMux() creates a new HTTP request multiplexer (router).
	// It matches the URL of each incoming request against a list of registered patterns and calls the handler for the pattern that most closely matches the URL.
	mux := http.NewServeMux()

	// Serve static files (CSS, JS, images) from the ./ui/static/ directory.
	// http.StripPrefix is used to remove the "/static/" prefix from the URL path before searching the file system.
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Register route handlers. Each mux.HandleFunc associates a specific URL path with a Go function.
	// For example, when a GET request is made to "/", the 'home' function will be executed.
	mux.HandleFunc("GET /{$}", home)
	mux.HandleFunc("GET /movies", moviesListView)
	mux.HandleFunc("GET /movies/{id}/{slug}", movieView)
	mux.HandleFunc("GET /tv-shows", showsListView)
	mux.HandleFunc("GET /tv-shows/{id}/{slug}", tvView)
	mux.HandleFunc("GET /people", peopleListView)
	mux.HandleFunc("GET /people/{id}/{slug}", personView)
	mux.HandleFunc("GET /search", searchView)
	mux.HandleFunc("GET /about", aboutView)
	mux.HandleFunc("GET /contact", contactView)
	mux.HandleFunc("GET /terms", termsView)
	mux.HandleFunc("GET /privacy", privacyView)
	mux.HandleFunc("GET /watchlist", requireAuth(watchlistView))

	mux.HandleFunc("GET /signup", signupView)
	mux.HandleFunc("POST /signup", signupPost)
	mux.HandleFunc("GET /login", loginView)
	mux.HandleFunc("POST /login", loginPost)
	mux.HandleFunc("POST /logout", logoutPost)
    
	mux.HandleFunc("GET /profile", requireAuth(profileView))
	mux.HandleFunc("POST /profile/edit", requireAuth(profileEditPost))
	mux.HandleFunc("GET /my-feed", requireAuth(myFeedView))
	mux.HandleFunc("POST /watchlist/toggle", requireAuth(toggleWatchlistPost))
	mux.HandleFunc("GET /wiki/edit", requireAuth(wikiEditView))
	mux.HandleFunc("POST /wiki/edit", requireAuth(wikiEditPost))
	mux.HandleFunc("GET /admin", adminRoleCheck(adminDashboardView))
	mux.HandleFunc("POST /admin/wiki/approve", adminRoleCheck(wikiApprovePost))
	mux.HandleFunc("POST /admin/wiki/reject", adminRoleCheck(wikiRejectPost))
	mux.HandleFunc("GET /ads", requireAuth(adsPortalView))
	mux.HandleFunc("POST /ads/create", requireAuth(createAdCampaignPost))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting web server on port :%s\n", port)
	
	// Create a custom nosurf handler to exclude static files if needed, and secure POST routes
	csrfHandler := nosurf.New(sessionMiddleware(mux))
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   false, // set true in prod
		SameSite: http.SameSiteStrictMode,
	})

	err := http.ListenAndServe(":"+port, rateLimit(recoverPanic(logRequest(secureHeaders(csrfHandler)))))
	log.Fatal(err)
}

// movieView is an HTTP handler function that responds to requests for individual movie details.
// It extracts the movie ID and slug from the URL path, retrieves the movie's data from the database,
// and renders the HTML template to display the movie page.
func movieView(w http.ResponseWriter, r *http.Request) {
	// Extract the 'id' and 'slug' path values from the URL, defined in the route: /movies/{id}/{slug}
	idStr := r.PathValue("id")
	slug := r.PathValue("slug")

	// Convert the ID string to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil || slug == "" {
		// If the ID is invalid or slug is missing, return a 404 Not Found error
		http.NotFound(w, r)
		return
	}

	// Define the slice of HTML template files needed to render the page.
	// The order typically includes the base layout, partials like nav/sidebar, and the specific page template.
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/movies.html",
	}

	// template.ParseFiles parses the template files and creates a new *template.Template.
	// We use this to convert our HTML files into executable templates.
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Fetch the complete movie details (including cast, crew, etc.) from the database layer.
	detail, err := database.GetMovieDetail(id)
	if err != nil {
		log.Println("Error fetching movie details:", err)
		http.NotFound(w, r) // Show a 404 if the movie doesn't exist in the DB
		return
	}

	// Prepare the template data structure. getTemplateData gives us common data (like nav info).
	data := getTemplateData(detail.Movie.Name, r)
	data.MovieDetail = detail
	// Fetch real-time affiliate listings for monetization based on the movie's name.
	data.EbayListings = monetization.FetchEbayListings(detail.Movie.Name)

	// Execute the "base" template, passing in our data structure.
	// This generates the final HTML and writes it to the http.ResponseWriter (w) to send to the client.
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/index.html",
	}

	ts, err := template.New("base").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"split": strings.Split,
		"trim": strings.Trim,
		"trimSpace": strings.TrimSpace,
		"trimLeft": strings.TrimLeft,
		"trimRight": strings.TrimRight,
		"firstGenre": func(genreJSON string) string {
			// Genre is stored like ["Action","Drama"] - parse out the first genre
			genreJSON = strings.TrimSpace(genreJSON)
			if len(genreJSON) == 0 || genreJSON == "[]" {
				return ""
			}
			// Strip leading/trailing brackets
			genreJSON = strings.TrimPrefix(genreJSON, "[")
			genreJSON = strings.TrimSuffix(genreJSON, "]")
			// Split by comma and get first entry
			parts := strings.SplitN(genreJSON, ",", 2)
			if len(parts) == 0 {
				return ""
			}
			// Remove surrounding quotes from genre name
			return strings.Trim(strings.TrimSpace(parts[0]), `"`)
		},
	}).ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("Home", r)
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

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func myFeedView(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/my-feed.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("My Feed", r)
	data.Ads = monetization.FetchAdvertisements("home")

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func tvView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	slug := r.PathValue("slug")
	id, err := strconv.Atoi(idStr)
	if err != nil || slug == "" {
		http.NotFound(w, r)
		return
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/tv_shows.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	detail, err := database.GetTVSeriesDetail(id)
	if err != nil {
		log.Println("Error fetching series details:", err)
		http.NotFound(w, r)
		return
	}

	data := getTemplateData(detail.Series.Name, r)
	data.TVSeriesDetail = detail
	data.EbayListings = monetization.FetchEbayListings(detail.Series.Name)

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func personView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	slug := r.PathValue("slug")
	id, err := strconv.Atoi(idStr)
	if err != nil || slug == "" {
		http.NotFound(w, r)
		return
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/people.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	detail, err := database.GetPersonDetailByID(id)
	if err != nil {
		log.Println("Error fetching person details:", err)
		http.NotFound(w, r)
		return
	}

	data := getTemplateData(detail.Person.Name, r)
	data.PersonDetail = detail

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func moviesListView(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/movies-list.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("Movies", r)
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

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func showsListView(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/tv-shows-list.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("TV Shows", r)
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

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func peopleListView(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/people-list.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("People", r)
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

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func searchView(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/search.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("Search", r)
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

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func watchlistView(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/watchlist.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("My Watchlist", r)
	
	movies, shows, err := database.GetUserWatchlist(data.AuthenticatedUser.ID)
	if err != nil {
		log.Println("Error fetching watchlist:", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	data.Movies = movies
	data.Shows = shows

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func aboutView(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/about.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("About", r)

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func contactView(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/contact.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("Contact", r)

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func termsView(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/terms.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("Terms of Service", r)

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func privacyView(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/privacy.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("Privacy Policy", r)

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func getTemplateData(title string, r *http.Request) templateData {
	movies, _ := database.GetAllMovies(10, 0, "")
	shows, _ := database.GetAllShows(10, 0, "")
	users, _ := database.GetAllUsers(10, 0)
	
	var authUser *models.User
	var csrfToken string
	if r != nil {
		authUser = getUser(r)
		csrfToken = nosurf.Token(r)
	}

	return templateData{
		Title:             title,
		Movies:            movies,
		Shows:             shows,
		Users:             users,
		AuthenticatedUser: authUser,
		CSRFToken:         csrfToken,
	}
}
