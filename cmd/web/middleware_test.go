package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecureHeaders(t *testing.T) {
	// Initialize a dummy application struct, if needed by the method receiver
	app := &application{}

	// Create a dummy handler that we will wrap with the middleware
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap the dummy handler with our secureHeaders middleware
	secureHandler := app.secureHeaders(nextHandler)

	// Create a new HTTP request (the method and URL don't matter for this test)
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Create a response recorder to capture the response from the middleware
	rr := httptest.NewRecorder()

	// Serve the request using the wrapped handler
	secureHandler.ServeHTTP(rr, req)

	// Assert that control passed to the next handler by checking the status code and body
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, rr.Code)
	}
	if rr.Body.String() != "OK" {
		t.Errorf("expected body %q; got %q", "OK", rr.Body.String())
	}

	// Define the list of expected headers
	expectedHeaders := []string{
		"Content-Security-Policy",
		"Referrer-Policy",
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
	}

	// Verify that each expected header has been set in the response
	for _, header := range expectedHeaders {
		if rr.Header().Get(header) == "" {
			t.Errorf("expected header %q to be set", header)
		}
	}
}
