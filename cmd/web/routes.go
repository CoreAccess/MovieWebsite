package main

import (
	"net/http"

	"github.com/justinas/nosurf"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /movies", app.moviesListView)
	mux.HandleFunc("GET /movies/{id}/{slug}", app.movieView)
	mux.HandleFunc("GET /tv-shows", app.showsListView)
	mux.HandleFunc("GET /tv-shows/{id}/{slug}", app.tvView)
	mux.HandleFunc("GET /people", app.peopleListView)
	mux.HandleFunc("GET /people/{id}/{slug}", app.personView)
	mux.HandleFunc("GET /search", app.searchView)
	mux.HandleFunc("GET /about", app.aboutView)
	mux.HandleFunc("GET /contact", app.contactView)
	mux.HandleFunc("GET /terms", app.termsView)
	mux.HandleFunc("GET /privacy", app.privacyView)
	mux.HandleFunc("GET /watchlist", app.requireAuth(app.watchlistView))
	mux.HandleFunc("GET /watchlist/{id}", app.requireAuth(app.watchlistView))
	mux.HandleFunc("GET /notifications", app.requireAuth(app.notificationsView))
	mux.HandleFunc("GET /blog", app.blogListView)
	mux.HandleFunc("GET /blog/{slug}", app.blogPostView)

	mux.HandleFunc("GET /signup", app.signupView)
	mux.HandleFunc("POST /signup", authRateLimit(app.signupPost))
	mux.HandleFunc("GET /login", app.loginView)
	mux.HandleFunc("POST /login", authRateLimit(app.loginPost))
	mux.HandleFunc("POST /logout", app.logoutPost)

	mux.HandleFunc("GET /profile", app.requireAuth(app.profileView))
	mux.HandleFunc("GET /profile/{id}/{username}", app.publicProfileView)
	mux.HandleFunc("POST /profile/edit", app.requireAuth(app.profileEditPost))
	mux.HandleFunc("GET /my-feed", app.requireAuth(app.myFeedView))
	mux.HandleFunc("POST /watchlist/toggle", app.requireAuth(app.toggleWatchlistPost))
	
	mux.HandleFunc("POST /rate", app.requireAuth(app.rateMedia))
	mux.HandleFunc("POST /review", app.requireAuth(app.reviewPost))
	
	mux.HandleFunc("POST /follow", app.requireAuth(app.followPost))
	mux.HandleFunc("POST /unfollow", app.requireAuth(app.unfollowPost))
	
	mux.HandleFunc("GET /lists", app.requireAuth(app.userListsView))
	mux.HandleFunc("GET /list/{id}/{slug}", app.listView)
	mux.HandleFunc("GET /franchise/{slug}", app.franchiseView)
	mux.HandleFunc("GET /list/create", app.requireAuth(app.listCreateView))
	mux.HandleFunc("POST /list/create", app.requireAuth(app.listCreatePost))
	mux.HandleFunc("POST /list/add-item", app.requireAuth(app.listItemAddPost))
	mux.HandleFunc("POST /list/remove-item", app.requireAuth(app.listItemRemovePost))

	mux.HandleFunc("GET /admin", app.adminRoleCheck(app.adminDashboardView))

	// API Endpoints
	mux.HandleFunc("GET /api/trending", app.trendingAPI)

	csrfHandler := nosurf.New(app.sessionMiddleware(mux))
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	return rateLimit(app.recoverPanic(app.logRequest(app.secureHeaders(csrfHandler))))
}

