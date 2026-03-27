package main

import (
	"net/http"
	"runtime/debug"
	"time"
)

// logRequest middleware logs details about all incoming requests using structured slog output.
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		app.logger.Info("request received",
			"method", r.Method,
			"uri", r.URL.RequestURI(),
			"proto", r.Proto,
			"remote_addr", r.RemoteAddr,
		)

		next.ServeHTTP(w, r)

		app.logger.Info("request completed",
			"method", r.Method,
			"uri", r.URL.RequestURI(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// recoverPanic middleware gracefully handles application panics to prevent server crashes.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.logger.Error("panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"method", r.Method,
					"uri", r.URL.RequestURI(),
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// secureHeaders adds defensive HTTP headers to enhance application security.
// CSP is now enforced (not Report-Only) since inline JS handlers have been removed.
func (app *application) secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://fonts.googleapis.com; "+
				"script-src 'self' https://cdn.jsdelivr.net; "+
				"font-src 'self' https://fonts.gstatic.com https://cdn.jsdelivr.net; "+
				"img-src 'self' data: https:; "+
				"frame-src 'self' https://www.youtube.com https://youtube.com; "+
				"connect-src 'self' https://cdn.jsdelivr.net;",
		)
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}
