package main

import (
	"net/http/httptest"
	"testing"
)

func TestIsSafeRedirect(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"Root path", "/", true},
		{"Local path", "/movies", true},
		{"Local path with query", "/movies?id=1", true},
		{"Empty string", "", false},
		{"External URL", "http://evil.com", false},
		{"HTTPS URL", "https://evil.com", false},
		{"Protocol relative URL", "//evil.com", false},
		{"Data URL", "data:text/html,<script>alert(1)</script>", false},
		{"Javascript URL", "javascript:alert(1)", false},
		{"Multiple slashes", "///evil.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSafeRedirect(tt.url)
			if result != tt.expected {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, result)
			}
		})
	}
}

func TestGetSafeReferer(t *testing.T) {
	tests := []struct {
		name     string
		referer  string
		host     string
		fallback string
		expected string
	}{
		{
			name:     "Valid internal referer",
			referer:  "http://localhost:8080/movies",
			host:     "localhost:8080",
			fallback: "/",
			expected: "/movies",
		},
		{
			name:     "Internal referer with query",
			referer:  "http://localhost:8080/search?q=test",
			host:     "localhost:8080",
			fallback: "/",
			expected: "/search?q=test",
		},
		{
			name:     "External referer",
			referer:  "http://evil.com/malicious",
			host:     "localhost:8080",
			fallback: "/profile",
			expected: "/profile",
		},
		{
			name:     "Empty referer",
			referer:  "",
			host:     "localhost:8080",
			fallback: "/",
			expected: "/",
		},
		{
			name:     "Relative referer (should be handled by browser but test logic)",
			referer:  "/local",
			host:     "localhost:8080",
			fallback: "/",
			expected: "/local",
		},
		{
			name:     "Malformed referer",
			referer:  "http://[fe80::%31%25en0]/",
			host:     "localhost:8080",
			fallback: "/",
			expected: "/",
		},
		{
			name:     "Protocol-relative referer",
			referer:  "//evil.com/test",
			host:     "localhost:8080",
			fallback: "/",
			expected: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://"+tt.host, nil)
			if tt.referer != "" {
				req.Header.Set("Referer", tt.referer)
			}

			result := getSafeReferer(req, tt.fallback)
			if result != tt.expected {
				t.Errorf("%s: expected %q, got %q", tt.name, tt.expected, result)
			}
		})
	}
}
