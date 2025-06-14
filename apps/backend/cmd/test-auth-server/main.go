package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"tennis-booker/internal/auth"
	"tennis-booker/internal/ratelimit"
	"tennis-booker/internal/secrets"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type ProtectedResponse struct {
	Message  string `json:"message"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

func main() {
	// Set up environment variables for testing
	os.Setenv("VAULT_ADDR", "http://localhost:8200")
	os.Setenv("VAULT_TOKEN", "dev-token")

	// Initialize secrets manager
	secretsManager, err := secrets.NewSecretsManagerFromEnv()
	if err != nil {
		log.Fatal("Failed to create secrets manager:", err)
	}
	defer secretsManager.Close()

	// Create JWT service
	jwtService := auth.NewJWTService(secretsManager, "tennis-booker-test")

	// Initialize rate limiter
	rateLimitConfig := ratelimit.DefaultConfig()
	// Override some defaults for testing
	rateLimitConfig.AuthEndpointLimit = ratelimit.RateLimit{
		Requests: 5, // 5 requests per minute for auth endpoints
		Window:   time.Minute,
	}
	rateLimitConfig.DefaultIPLimit = ratelimit.RateLimit{
		Requests: 20, // 20 requests per minute for general IP limiting
		Window:   time.Minute,
	}
	rateLimitConfig.DefaultUserLimit = ratelimit.RateLimit{
		Requests: 50, // 50 requests per minute for authenticated users
		Window:   time.Minute,
	}

	rateLimiter, err := ratelimit.NewLimiter(rateLimitConfig)
	if err != nil {
		log.Printf("Warning: Failed to create rate limiter (Redis may not be available): %v", err)
		log.Println("Continuing without rate limiting...")
		setupRoutesWithoutRateLimit(jwtService, secretsManager)
	} else {
		defer rateLimiter.Close()
		log.Println("✅ Rate limiter initialized successfully")
		setupRoutesWithRateLimit(jwtService, secretsManager, rateLimiter)
	}

	fmt.Println("🚀 Test Auth Server with Rate Limiting starting on :8080")
	fmt.Println("📋 Available endpoints:")
	fmt.Println("  POST /login - Login with username/password (Rate Limited: 5/min)")
	fmt.Println("  POST /refresh - Refresh access token (Rate Limited: 5/min)")
	fmt.Println("  GET /protected - Protected route (Rate Limited: 50/min per user)")
	fmt.Println("  GET /health - Health check (Rate Limited: 20/min per IP)")
	fmt.Println()
	fmt.Println("🧪 Test commands:")
	fmt.Println("  curl -X POST http://localhost:8080/login -H 'Content-Type: application/json' -d '{\"username\":\"testuser\",\"password\":\"testpass\"}'")
	fmt.Println("  curl -H 'Authorization: Bearer <token>' http://localhost:8080/protected")
	fmt.Println()
	fmt.Println("🔒 Rate Limiting:")
	fmt.Println("  - Auth endpoints: 5 requests/minute per IP")
	fmt.Println("  - Protected endpoints: 50 requests/minute per user + 20/minute per IP")
	fmt.Println("  - Public endpoints: 20 requests/minute per IP")
	fmt.Println("  - Rate limit headers included in responses")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupRoutesWithRateLimit(jwtService *auth.JWTService, secretsManager *secrets.SecretsManager, rateLimiter *ratelimit.Limiter) {
	// Create rate limiting middleware instances
	authRateLimit := ratelimit.AuthRateLimitMiddleware(rateLimiter)
	ipRateLimit := ratelimit.IPRateLimitMiddleware(rateLimiter)
	userRateLimit := ratelimit.UserRateLimitMiddleware(rateLimiter)

	// Authentication endpoints with strict rate limiting
	http.Handle("/login", authRateLimit(http.HandlerFunc(loginHandler(jwtService))))
	http.Handle("/refresh", authRateLimit(http.HandlerFunc(refreshHandler(jwtService))))

	// Protected route with user-based rate limiting (after JWT middleware)
	protectedHandler := userRateLimit(auth.JWTMiddleware(jwtService)(http.HandlerFunc(protectedRouteHandler)))
	http.Handle("/protected", protectedHandler)

	// Health check with basic IP rate limiting
	http.Handle("/health", ipRateLimit(http.HandlerFunc(healthHandler(secretsManager))))

	// Rate limit status endpoint (no rate limiting for monitoring)
	http.HandleFunc("/rate-limit-status", rateLimitStatusHandler(rateLimiter))
}

func setupRoutesWithoutRateLimit(jwtService *auth.JWTService, secretsManager *secrets.SecretsManager) {
	// Fallback setup without rate limiting
	http.HandleFunc("/login", loginHandler(jwtService))
	http.HandleFunc("/refresh", refreshHandler(jwtService))

	protectedHandler := auth.JWTMiddleware(jwtService)(http.HandlerFunc(protectedRouteHandler))
	http.Handle("/protected", protectedHandler)

	http.HandleFunc("/health", healthHandler(secretsManager))
	http.HandleFunc("/rate-limit-status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "disabled",
			"reason": "Redis not available",
		})
	})
}

func rateLimitStatusHandler(rateLimiter *ratelimit.Limiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check rate limiter health
		if err := rateLimiter.HealthCheck(context.Background()); err != nil {
			http.Error(w, fmt.Sprintf("Rate limiter health check failed: %v", err), http.StatusServiceUnavailable)
			return
		}

		status := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"redis":     "connected",
			"limits": map[string]interface{}{
				"auth_endpoints": "5 requests/minute",
				"user_endpoints": "50 requests/minute",
				"ip_endpoints":   "20 requests/minute",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}

func loginHandler(jwtService *auth.JWTService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Simple authentication (in real app, check against database)
		if req.Username == "" || req.Password == "" {
			http.Error(w, "Username and password required", http.StatusBadRequest)
			return
		}

		// For testing, accept any non-empty credentials
		userID := fmt.Sprintf("user_%s", req.Username)

		// Generate tokens
		accessToken, err := jwtService.GenerateToken(userID, req.Username, 15*time.Minute)
		if err != nil {
			log.Printf("Failed to generate access token: %v", err)
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		refreshToken, err := jwtService.GenerateRefreshToken(userID, req.Username)
		if err != nil {
			log.Printf("Failed to generate refresh token: %v", err)
			http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
			return
		}

		response := LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    900, // 15 minutes
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func refreshHandler(jwtService *auth.JWTService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get refresh token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		refreshToken := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			refreshToken = authHeader[7:]
		}

		// Generate new access token
		newAccessToken, err := jwtService.RefreshAccessToken(refreshToken, 15*time.Minute)
		if err != nil {
			log.Printf("Failed to refresh token: %v", err)
			http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
			return
		}

		response := map[string]interface{}{
			"access_token": newAccessToken,
			"expires_in":   900,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func protectedRouteHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user claims from context (set by middleware)
	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		http.Error(w, "Failed to get user claims", http.StatusInternalServerError)
		return
	}

	response := ProtectedResponse{
		Message:  "🎾 Welcome to the protected tennis court booking area!",
		UserID:   claims.UserID,
		Username: claims.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func healthHandler(secretsManager *secrets.SecretsManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check Vault connection
		if err := secretsManager.HealthCheck(); err != nil {
			http.Error(w, fmt.Sprintf("Vault health check failed: %v", err), http.StatusServiceUnavailable)
			return
		}

		// Try to fetch a test secret to verify everything works
		_, err := secretsManager.GetJWTSecret()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to fetch JWT secret: %v", err), http.StatusServiceUnavailable)
			return
		}

		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"vault":     "connected",
			"jwt":       "configured",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
