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

func init() {
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			limiter.mu.Lock()
			// Reset map every minute (basic fixed-window rate limiter)
			limiter.clients = make(map[string]int)
			limiter.mu.Unlock()
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
