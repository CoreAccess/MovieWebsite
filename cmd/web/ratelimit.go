package main

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type rateLimiter struct {
	mu      sync.Mutex
	clients map[string]int
}

var limiter = &rateLimiter{
	clients: make(map[string]int),
}

var authLimiter = &rateLimiter{
	clients: make(map[string]int),
}

func init() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.mu.Lock()
			// Reset map every minute (basic fixed-window rate limiter)
			limiter.clients = make(map[string]int)
			limiter.mu.Unlock()

			authLimiter.mu.Lock()
			// Reset map every minute for auth limiter
			authLimiter.clients = make(map[string]int)
			authLimiter.mu.Unlock()
		}
	}()
}

// rateLimit is a simple IP-based rate limiting middleware
func rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		// Extract IP without port
		if strings.Contains(ip, ":") {
			host, _, err := net.SplitHostPort(ip)
			if err == nil {
				ip = host
			}
		}

		limiter.mu.Lock()
		limiter.clients[ip]++
		count := limiter.clients[ip]
		limiter.mu.Unlock()

		// Limit to 100 requests per minute per IP
		if count > 100 {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authRateLimit is a stricter IP-based rate limiting middleware for authentication endpoints
func authRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		// Extract IP without port
		if strings.Contains(ip, ":") {
			host, _, err := net.SplitHostPort(ip)
			if err == nil {
				ip = host
			}
		}

		authLimiter.mu.Lock()
		authLimiter.clients[ip]++
		count := authLimiter.clients[ip]
		authLimiter.mu.Unlock()

		// Limit to 5 requests per minute per IP for sensitive endpoints like login/signup
		if count > 5 {
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}
