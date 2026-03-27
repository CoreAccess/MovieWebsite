package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"filmgap/internal/models"
	"filmgap/internal/service"

	"github.com/justinas/nosurf"
)

// application holds the application-wide dependencies.
type application struct {
	Service *service.AppService

	logger        *slog.Logger
	templateCache map[string]*template.Template
}

type templateData struct {
	JSONLD      string // Raw JSON-LD payload to be injected into the <head> of base.tmpl
	CurrentYear int
	Flash       string
	SiteName    string

	Title             string
	Hero              any
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
	Lists             []models.List
	List              models.List
	ListItems         []models.ListItem
	AuthenticatedUser *models.User
	CSRFToken         string
	ProfileUser       models.User
	IsFollowing       bool
	Watchlists        []models.Watchlist
	EntityType        string
	EntityID          int
	UserCount         int
	MediaCount        int
	Sort              string
	Next              string
	Filter            string

	// Homepage Expansion Fields
	Stats               models.HomepageStats
	Trending            []models.MediaSummary
	Activity            []models.Activity
	PopularLists        []models.List
	Franchise           models.Franchise
	BlogPosts           []models.BlogPost
	Photos              []models.Photo
	Birthdays           []models.Person
	FanFavorites        []models.MediaSummary
	BoxOffice           []models.MediaSummary
	PopularCelebs       []models.Person
	WatchlistMap        map[int]bool
	WatchProviderGroups []models.WatchProviderGroup
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
		"lower":     strings.ToLower,
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
		"formatRating": func(rating float64) string {
			return fmt.Sprintf("%.1f", rating)
		},
		"formatDate": func(dateStr string) string {
			parsed, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return dateStr
			}
			return parsed.Format("January 2, 2006")
		},
		"yearFromDate": func(dateStr string) string {
			dateStr = strings.TrimSpace(dateStr)
			if len(dateStr) >= 4 {
				return dateStr[:4]
			}
			return ""
		},
		"truncateWords": func(text string, limit int) string {
			words := strings.Fields(strings.TrimSpace(text))
			if len(words) == 0 {
				return ""
			}
			if limit <= 0 || len(words) <= limit {
				return strings.Join(words, " ")
			}
			return strings.Join(words[:limit], " ") + "..."
		},
		"humanDate": func(t time.Time) string {
			if t.IsZero() {
				return "Unknown"
			}
			return t.Format("Jan 02, 2006")
		},
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
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
			"./ui/html/partials/interaction_modal.tmpl",
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

// serverError logs the error with structured context and sends a generic 500 response.
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.logger.Error("server error",
		"error", trace,
		"method", r.Method,
		"uri", r.URL.RequestURI(),
		"remote_addr", r.RemoteAddr,
	)
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
	movies, _ := app.Service.GetAllMovies(10, 0, "")
	shows, _ := app.Service.GetAllShows(10, 0, "")
	users, _ := app.Service.GetAllUsers(10, 0)

	var authUser *models.User
	var csrfToken string
	var userLists []models.List
	if r != nil {
		authUser = app.getUser(r)
		csrfToken = nosurf.Token(r)
		if authUser != nil {
			userLists, _ = app.Service.GetListsByUserID(authUser.ID)
		}
	}

	return &templateData{
		Title:             title,
		Movies:            movies,
		Shows:             shows,
		Users:             users,
		AuthenticatedUser: authUser,
		CSRFToken:         csrfToken,
		Lists:             userLists,
	}
}

// render fetches the requested template from the cache, executes it, and writes the output.
func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	// Initialize a new buffer.
	buf := new(bytes.Buffer)

	// Execute the template to the buffer, to catch any execution errors.
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
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

// addDefaultData provides default template context values.
func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}
	td.CurrentYear = 2026
	td.Flash = "Notice: Application Architecture is being upgraded to Enterprise standards"
	td.SiteName = "filmgap Schema.org Optimized"
	return td
}

func (app *application) getWatchlistMap(userID int) map[int]bool {
	watchlistMap := make(map[int]bool)
	movies, shows, err := app.Service.Repo.GetUserWatchlist(userID)
	if err == nil {
		for _, m := range movies {
			watchlistMap[m.ID] = true
		}
		for _, s := range shows {
			watchlistMap[s.ID] = true
		}
	}
	return watchlistMap
}
