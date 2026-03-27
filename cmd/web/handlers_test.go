package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMovieView(t *testing.T) {
	app := &application{
		logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

	tests := []struct {
		name           string
		id             string
		slug           string
		expectedStatus int
	}{
		{
			name:           "Invalid ID",
			id:             "abc",
			slug:           "test-movie",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty Slug",
			id:             "1",
			slug:           "",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/movies/", nil)
			req.SetPathValue("id", tt.id)
			req.SetPathValue("slug", tt.slug)

			rr := httptest.NewRecorder()

			app.movieView(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d; got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
