package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CORSMiddleware handles CORS headers for all requests
func CORSMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Define allowed origins
			allowedOrigins := []string{
				"http://localhost:3000",
				"http://localhost:5173",
				"http://127.0.0.1:3000",
				"http://127.0.0.1:5173",
				"http://localhost:8080",
				"http://127.0.0.1:8080",
			}

			// Add production frontend URL from environment
			if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
				allowedOrigins = append(allowedOrigins, frontendURL)
			}

			// Add Vercel preview deployment domains (pattern: https://*-username.vercel.app)
			if strings.Contains(origin, ".vercel.app") {
				allowedOrigins = append(allowedOrigins, origin)
			}

			// Check if origin is allowed
			isAllowed := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					isAllowed = true
					break
				}
			}

			// For development, be more permissive with localhost
			if !isAllowed && (strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1")) {
				isAllowed = true
			}

			if isAllowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				// Default to localhost for development
				w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
			}

			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}
