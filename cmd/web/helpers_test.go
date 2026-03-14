package main

import (
	"net/http"
	"testing"
)

func TestIsSafeRedirect(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		// Valid cases
		{"Root path", "/", true},
		{"Relative path", "/login", true},
		{"Deep relative path", "/about/team?sort=asc", true},

		// Invalid cases
		{"Empty string", "", false},
		{"Absolute URL (http)", "http://example.com", false},
		{"Absolute URL (https)", "https://example.com", false},
		{"No slash", "login", false},

		// Bypass vectors
		{"Protocol relative (//)", "//example.com", false},
		{"Protocol relative (/\\)", "/\\example.com", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isSafeRedirect(tc.url)
			if result != tc.expected {
				t.Errorf("isSafeRedirect(%q) = %v, expected %v", tc.url, result, tc.expected)
			}
		})
	}
}

func TestGetSafeReferer(t *testing.T) {
	fallback := "/fallback"

	tests := []struct {
		name     string
		referer  string
		expected string
	}{
		// Valid referers
		{"Empty referer", "", fallback},
		{"Safe local referer", "http://localhost:8080/movies", "/movies"},
		{"Safe local referer with query", "http://localhost:8080/search?q=test", "/search?q=test"},
		{"Path only referer", "/search?q=foo", "/search?q=foo"},

		// Unsafe or malicious referers
		{"External referer", "http://evil.com", fallback}, // url.Parse gives empty Path for http://evil.com, which fails isSafeRedirect
		{"External referer with path", "http://evil.com/malicious", "/malicious"}, // the host is dropped, so we end up redirecting to /malicious on our domain, which is safe.
		{"Protocol relative referer", "//evil.com", fallback}, // u.Path empty
		{"Malformed referer", ": malformed", fallback},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			if tc.referer != "" {
				req.Header.Set("Referer", tc.referer)
			}
			result := getSafeReferer(req, fallback)
			if result != tc.expected {
				t.Errorf("getSafeReferer(%q) = %v, expected %v", tc.referer, result, tc.expected)
			}
		})
	}
}
