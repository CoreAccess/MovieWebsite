package main

import (
	"net/http"
	"net/url"
	"strings"
)

// isSafeRedirect checks if a given target URL is safe for redirection.
// It is safe if it is a local path (starts with / but not //).
// This helps prevent Open Redirect vulnerabilities.
func isSafeRedirect(target string) bool {
	if target == "" {
		return false
	}
	// Ensure it starts with a single '/' to prevent protocol-relative redirects (e.g. //attacker.com)
	// and absolute URLs (e.g. http://attacker.com)
	return strings.HasPrefix(target, "/") && !strings.HasPrefix(target, "//")
}

// getSafeReferer extracts a safe local path from the request's Referer header.
// If the Referer is missing, invalid, or points to a different host, it returns the fallback.
func getSafeReferer(r *http.Request, fallback string) string {
	referer := r.Header.Get("Referer")
	if referer == "" {
		return fallback
	}

	u, err := url.Parse(referer)
	if err != nil {
		return fallback
	}

	// If the Referer has a host, it must match the current request's host.
	// This ensures we only redirect back to our own application.
	if u.Host != "" && u.Host != r.Host {
		return fallback
	}

	// Reconstruct the local path and query from the Referer URL.
	path := u.Path
	if path == "" {
		path = "/"
	}

	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}

	// Final check to ensure the resulting path is safe for redirection.
	if !isSafeRedirect(path) {
		return fallback
	}

	return path
}
