package main

import (
	"net/http"
	"testing"
)

func TestIsSafeRedirect(t *testing.T) {
	tests := []struct {
		path string
		safe bool
	}{
		{"/", true},
		{"/profile", true},
		{"/movies?id=123", true},
		{"/admin/", true},
		{"", false},
		{"//evil.com", false},
		{"/\\evil.com", false},
		{"https://evil.com", false},
		{"http://evil.com", false},
		{"javascript:alert(1)", false},
		{"evil.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isSafeRedirect(tt.path); got != tt.safe {
				t.Errorf("isSafeRedirect(%q) = %v, want %v", tt.path, got, tt.safe)
			}
		})
	}
}

func TestGetSafeReferer(t *testing.T) {
	tests := []struct {
		name     string
		referer  string
		fallback string
		want     string
	}{
		{
			name:     "Empty referer",
			referer:  "",
			fallback: "/profile",
			want:     "/profile",
		},
		{
			name:     "Safe relative path",
			referer:  "/movies/1",
			fallback: "/profile",
			want:     "/movies/1",
		},
		{
			name:     "Safe relative path with query",
			referer:  "/search?q=batman",
			fallback: "/profile",
			want:     "/search?q=batman",
		},
		{
			name:     "Safe absolute URL",
			referer:  "http://localhost:8080/movies/1",
			fallback: "/profile",
			want:     "/movies/1", // Extracts path from absolute url
		},
		{
			name:     "Unsafe external URL",
			referer:  "https://evil.com/phishing",
			fallback: "/profile",
			want:     "/phishing", // url.Parse on https://evil.com/phishing gives path /phishing. This is acceptable since we redirect locally to /phishing.
		},
		{
			name:     "Protocol-relative URL",
			referer:  "//evil.com",
			fallback: "/profile",
			want:     "/profile", // gets blocked by empty path + host check
		},
		{
			name:     "Protocol-relative URL with path",
			referer:  "//evil.com/hello",
			fallback: "/profile",
			want:     "/hello", // extracts /hello, which is safe locally
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{Header: http.Header{}}
			if tt.referer != "" {
				r.Header.Set("Referer", tt.referer)
			}

			if got := getSafeReferer(r, tt.fallback); got != tt.want {
				t.Errorf("getSafeReferer(%q) = %q, want %q", tt.referer, got, tt.want)
			}
		})
	}
}
