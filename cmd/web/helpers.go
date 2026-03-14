package main

import (
	"bytes"
	"fmt"
	"html/template"

	"movieweb/internal/service"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime/debug"
	"strings"

	"movieweb/internal/database"
	"movieweb/internal/models"

	"github.com/justinas/nosurf"
)

// application holds the application-wide dependencies.
type application struct {
	Service *service.AppService

	errorLog      *log.Logger
	infoLog       *log.Logger
	templateCache map[string]*template.Template
}

type templateData struct {
	JSONLD          string // Raw JSON-LD payload to be injected into the <head> of base.tmpl
	CurrentYear     int
	Flash           string
	SiteName        string

	Title             string
	Movies            any
	Shows             any
	People            any
	Users             any
	PopularMovies     any
	UpcomingMovies    any
	PopularShows      any
	NewShows          any
	MovieDetail       any
	TVSeriesDetail    any
	PersonDetail      any
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

// newTemplateCache initializes an in-memory cache of parsed HTML templates.
// It parses all page templates along with their required base and partials.
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// Define our custom template functions.
	functions := template.FuncMap{
		"add":       func(a, b int) int { return a + b },
		"split":     strings.Split,
		"trim":      strings.Trim,
		"trimSpace": strings.TrimSpace,
		"trimLeft":  strings.TrimLeft,
		"trimRight": strings.TrimRight,
		"firstGenre": func(genreJSON string) string {
			genreJSON = strings.TrimSpace(genreJSON)
			if len(genreJSON) == 0 || genreJSON == "[]" {
				return ""
			}
			genreJSON = strings.TrimPrefix(genreJSON, "[")
			genreJSON = strings.TrimSuffix(genreJSON, "]")
			parts := strings.SplitN(genreJSON, ",", 2)
			if len(parts) == 0 {
				return ""
			}
			return strings.Trim(strings.TrimSpace(parts[0]), `"`)
		},
	}

	// Use filepath.Glob to get a slice of all filepaths with the .html extension
	// in the ui/html/pages directory.
	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		// Create a slice containing the base template.
		files := []string{
			"./ui/html/base.tmpl",
			"./ui/html/partials/nav.tmpl",
			"./ui/html/partials/sidebar.tmpl",
			page,
		}

		// Parse the files into a template set.
		ts, err := template.New(name).Funcs(functions).ParseFiles(files...)
		if err != nil {
			return nil, err
		}

		// Add the template set to the cache, using the name of the page as the key.
		cache[name] = ts
	}

	return cache, nil
}

// serverError logs the detailed error and sends a generic 500 Internal Server Error response.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// clientError sends a specific status code and corresponding description to the client.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// notFound is a convenience wrapper around clientError which sends a 404 Not Found response.
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

// getTemplateData returns a pointer to a templateData struct initialized with common dynamic data.
func (app *application) getTemplateData(title string, r *http.Request) *templateData {
	movies, _ := database.GetAllMovies(10, 0, "")
	shows, _ := database.GetAllShows(10, 0, "")
	users, _ := database.GetAllUsers(10, 0)

	var authUser *models.User
	var csrfToken string
	if r != nil {
		authUser = app.getUser(r)
		csrfToken = nosurf.Token(r)
	}

	return &templateData{
		Title:             title,
		Movies:            movies,
		Shows:             shows,
		Users:             users,
		AuthenticatedUser: authUser,
		CSRFToken:         csrfToken,
	}
}

// render fetches the requested template from the cache, executes it, and writes the output.
func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	// Initialize a new buffer.
	buf := new(bytes.Buffer)

	// Execute the template to the buffer, to catch any execution errors.
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// If the template is written to the buffer without any errors, we are safe to write the
	// HTTP status code to the http.ResponseWriter.
	w.WriteHeader(status)

	// Write the contents of the buffer to the http.ResponseWriter.
	buf.WriteTo(w)
}

// isSafeRedirect checks if the provided URL path is a safe relative path for a redirect.
// It mitigates Open Redirect vulnerabilities by ensuring the path starts with a single '/'
// and not a double slash ('//' or '/\') which browsers interpret as a protocol-relative URL.
func isSafeRedirect(path string) bool {
	if len(path) == 0 {
		return false
	}
	if path[0] != '/' {
		return false
	}
	// Prevent protocol-relative URL bypasses like //malicious.com or /\malicious.com
	if len(path) > 1 && (path[1] == '/' || path[1] == '\\') {
		return false
	}
	return true
}

// getSafeReferer safely extracts the relative path and query from the HTTP Referer header.
// If the Referer is missing, invalid, or considered unsafe, it returns the provided fallback path.
// This prevents attackers from constructing malicious links that exploit Referer-based redirects.
func getSafeReferer(r *http.Request, fallback string) string {
	referer := r.Header.Get("Referer")
	if referer == "" {
		return fallback
	}

	// Parse the Referer URL
	u, err := url.Parse(referer)
	if err != nil {
		return fallback
	}

	// We only want to redirect to the path (and query), dropping the scheme and host.
	// This ensures we always redirect locally regardless of what the host was.
	redirectPath := u.Path
	if u.RawQuery != "" {
		redirectPath += "?" + u.RawQuery
	}

	// Even though we extracted the path from url.Parse, we still double-check it
	if !isSafeRedirect(redirectPath) {
		return fallback
	}

	return redirectPath
}

// Inject JSON-LD Schema.org data for SEO and AI Agent understanding

// Provide the JSON-LD to the template context
func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}
	td.CurrentYear = 2026
	td.Flash = "Notice: Application Architecture is being upgraded to Enterprise standards"
	td.SiteName = "MovieWeb Schema.org Optimized"
	return td
}
