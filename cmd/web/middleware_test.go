package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogRequest(t *testing.T) {
	// Keep track of the original output
	oldOutput := log.Writer()
	defer log.SetOutput(oldOutput)

	var buf bytes.Buffer
	log.SetOutput(&buf)

	// Create a dummy handler that we want to wrap
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	rr := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Mock RemoteAddr as logRequest uses it
	req.RemoteAddr = "127.0.0.1:1234"

	handler := logRequest(nextHandler)
	handler.ServeHTTP(rr, req)

	output := buf.String()

	if !strings.Contains(output, "REQ: 127.0.0.1:1234 - HTTP/1.1 GET /foo") {
		t.Errorf("expected log to contain REQ line; got %q", output)
	}

	if !strings.Contains(output, "RES: GET /foo completed in") {
		t.Errorf("expected log to contain RES line; got %q", output)
	}

	if rr.Body.String() != "OK" {
		t.Errorf("expected body to be 'OK'; got %q", rr.Body.String())
	}
}

func TestRecoverPanic(t *testing.T) {
	// Keep track of the original output
	oldOutput := log.Writer()
	defer log.SetOutput(oldOutput)

	var buf bytes.Buffer
	log.SetOutput(&buf)

	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	handler := recoverPanic(nextHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d; got %d", http.StatusInternalServerError, rr.Code)
	}

	if rr.Header().Get("Connection") != "close" {
		t.Errorf("expected Connection header to be 'close'; got %q", rr.Header().Get("Connection"))
	}

	expectedBody := "Internal Server Error\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected body to be %q; got %q", expectedBody, rr.Body.String())
	}

	output := buf.String()
	if !strings.Contains(output, "PANIC: test panic") {
		t.Errorf("expected log to contain PANIC line; got %q", output)
	}
}

func TestSecureHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	handler := secureHeaders(nextHandler)
	handler.ServeHTTP(rr, req)

	expectedHeaders := map[string]string{
		"Content-Security-Policy": "default-src 'self'; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://fonts.googleapis.com; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; font-src 'self' https://fonts.gstatic.com https://cdn.jsdelivr.net; img-src 'self' data: https:; frame-src 'self'",
		"Referrer-Policy":         "origin-when-cross-origin",
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "deny",
		"X-XSS-Protection":        "0",
	}

	for header, expectedValue := range expectedHeaders {
		actualValue := rr.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("expected %s header to be %q; got %q", header, expectedValue, actualValue)
		}
	}

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, rr.Code)
	}

	if rr.Body.String() != "OK" {
		t.Errorf("expected body to be 'OK'; got %q", rr.Body.String())
	}
}
