package main

import (
	"context"
	"movieweb/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminRoleCheck(t *testing.T) {
	// Create a dummy handler that we want to protect
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
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
			user:           &models.User{Role: "user"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Authenticated Admin",
			user:           &models.User{Role: "admin"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/admin", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tt.user != nil {
				ctx := context.WithValue(req.Context(), "user", *tt.user)
				req = req.WithContext(ctx)
			}

			handler := adminRoleCheck(nextHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d; got %d", tt.name, tt.expectedStatus, rr.Code)
			}
		})
	}
}
