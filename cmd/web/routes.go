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

	mux.HandleFunc("GET /signup", app.signupView)
	mux.HandleFunc("POST /signup", app.signupPost)
	mux.HandleFunc("GET /login", app.loginView)
	mux.HandleFunc("POST /login", app.loginPost)
	mux.HandleFunc("POST /logout", app.logoutPost)

	mux.HandleFunc("GET /profile", app.requireAuth(app.profileView))
	mux.HandleFunc("POST /profile/edit", app.requireAuth(app.profileEditPost))
	mux.HandleFunc("GET /my-feed", app.requireAuth(app.myFeedView))
	mux.HandleFunc("POST /watchlist/toggle", app.requireAuth(app.toggleWatchlistPost))
	mux.HandleFunc("GET /wiki/edit", app.requireAuth(app.wikiEditView))
	mux.HandleFunc("POST /wiki/edit", app.requireAuth(app.wikiEditPost))
	mux.HandleFunc("GET /admin", app.adminRoleCheck(app.adminDashboardView))
	mux.HandleFunc("POST /admin/wiki/approve", app.adminRoleCheck(app.wikiApprovePost))
	mux.HandleFunc("POST /admin/wiki/reject", app.adminRoleCheck(app.wikiRejectPost))
	mux.HandleFunc("GET /ads", app.requireAuth(app.adsPortalView))
	mux.HandleFunc("POST /ads/create", app.requireAuth(app.createAdCampaignPost))

	csrfHandler := nosurf.New(app.sessionMiddleware(mux))
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})

	return rateLimit(app.recoverPanic(app.logRequest(app.secureHeaders(csrfHandler))))
}
