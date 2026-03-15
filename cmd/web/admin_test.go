package main

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"movieweb/internal/models"
)

func TestAdminRoleCheck(t *testing.T) {
	// Create an application struct with discarded log output
	app := &application{
		errorLog: log.New(io.Discard, "", 0),
		infoLog:  log.New(io.Discard, "", 0),
		logger:   slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

	// Create a dummy next handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name           string
		user           *models.User
		expectedStatus int
	}{
		{
			name:           "Unauthenticated",
			user:           nil,
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "Authenticated Non-Admin",
			user:           &models.User{ID: 1, Username: "testuser", Role: "user"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Authenticated Admin",
			user:           &models.User{ID: 2, Username: "adminuser", Role: "admin"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/admin", nil)

			if tt.user != nil {
				// The getUser helper expects a models.User struct value, not a pointer.
				ctx := context.WithValue(req.Context(), "user", *tt.user)
				req = req.WithContext(ctx)
			}

			handler := app.adminRoleCheck(nextHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d; got %d", tt.name, tt.expectedStatus, rr.Code)
			}
		})
	}
}
