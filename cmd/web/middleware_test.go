package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoverPanic(t *testing.T) {
	// Create an application struct with a discarded log output to keep tests quiet
	app := &application{
		errorLog: log.New(io.Discard, "", 0),
		infoLog:  log.New(io.Discard, "", 0),
	}

	// Capture the standard log output to prevent the panic log from printing to the console
	// since recoverPanic uses log.Printf
	originalLogOutput := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(originalLogOutput)

	// Create a dummy handler that panics
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Wrap the dummy handler with the recoverPanic middleware
	wrappedHandler := app.recoverPanic(dummyHandler)

	// Create a test request and response recorder
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	// Execute the handler
	wrappedHandler.ServeHTTP(rr, req)

	// Check if the panic was recovered and response is 500
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	// Check if the Connection: close header was set
	if rr.Header().Get("Connection") != "close" {
		t.Errorf("expected Connection header to be close, got %s", rr.Header().Get("Connection"))
	}

	// Check if the response body contains Internal Server Error
	expectedBody := "Internal Server Error\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, rr.Body.String())
	}
}
