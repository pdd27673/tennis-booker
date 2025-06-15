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

// CORS middleware to handle cross-origin requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // In production, specify exact origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continue with the next handler
		next.ServeHTTP(w, r)
	})
}

// Enhanced CORS middleware for production (allows specific origins)
func corsMiddlewareProduction(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow specific origins (add your frontend URLs here)
		allowedOrigins := []string{
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:3000",
			"http://127.0.0.1:5173",
			"http://127.0.0.1:5174",
			"http://127.0.0.1:3000",
		}

		// Check if origin is allowed
		originAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				originAllowed = true
				break
			}
		}

		if originAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continue with the next handler
		next.ServeHTTP(w, r)
	})
}

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
	fmt.Println("🌐 CORS enabled for frontend integration")
	fmt.Println("📋 Available endpoints:")
	fmt.Println("  POST /api/auth/login - Login with username/password (Rate Limited: 5/min)")
	fmt.Println("  POST /api/auth/register - Register new user (Rate Limited: 5/min)")
	fmt.Println("  POST /api/auth/refresh - Refresh access token (Rate Limited: 5/min)")
	fmt.Println("  POST /api/auth/logout - Logout user (Rate Limited: 5/min)")
	fmt.Println("  GET /api/auth/me - Get user info (Rate Limited: 50/min per user)")
	fmt.Println("  GET/PUT /api/users/preferences - User preferences (Rate Limited: 50/min per user)")
	fmt.Println("  GET /api/venues - Get available venues (Rate Limited: 20/min per IP)")
	fmt.Println("  GET /api/courts - Get court slots (Rate Limited: 20/min per IP)")
	fmt.Println("  GET /api/health - Health check (Rate Limited: 20/min per IP)")
	fmt.Println("  GET /api/system/health - System health (Rate Limited: 20/min per IP)")
	fmt.Println("  GET /api/system/status - System status (Rate Limited: 20/min per IP)")
	fmt.Println("  POST /api/system/pause - Pause scraping (Rate Limited: 20/min per IP)")
	fmt.Println("  POST /api/system/resume - Resume scraping (Rate Limited: 20/min per IP)")
	fmt.Println("  POST /api/system/restart - Restart system (Rate Limited: 20/min per IP)")
	fmt.Println()
	fmt.Println("🧪 Test commands:")
	fmt.Println("  curl -X POST http://localhost:8080/api/auth/login -H 'Content-Type: application/json' -d '{\"username\":\"testuser\",\"password\":\"testpass\"}'")
	fmt.Println("  curl -H 'Authorization: Bearer <token>' http://localhost:8080/api/auth/me")
	fmt.Println("  curl http://localhost:8080/api/venues")
	fmt.Println("  curl http://localhost:8080/api/courts")
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

	// Apply CORS middleware to all routes
	// Authentication endpoints with strict rate limiting
	http.Handle("/api/auth/login", corsMiddlewareProduction(authRateLimit(http.HandlerFunc(loginHandler(jwtService)))))
	http.Handle("/api/auth/register", corsMiddlewareProduction(authRateLimit(http.HandlerFunc(registerHandler(jwtService)))))
	http.Handle("/api/auth/refresh", corsMiddlewareProduction(authRateLimit(http.HandlerFunc(refreshHandler(jwtService)))))
	http.Handle("/api/auth/logout", corsMiddlewareProduction(authRateLimit(http.HandlerFunc(logoutHandler()))))

	// Protected route with user-based rate limiting (after JWT middleware)
	protectedHandler := userRateLimit(auth.JWTMiddleware(jwtService)(http.HandlerFunc(protectedRouteHandler)))
	http.Handle("/api/auth/me", corsMiddlewareProduction(protectedHandler))

	// User preferences endpoint (protected)
	preferencesHandler := userRateLimit(auth.JWTMiddleware(jwtService)(http.HandlerFunc(userPreferencesHandler)))
	http.Handle("/api/users/preferences", corsMiddlewareProduction(preferencesHandler))

	// Court and venue endpoints
	http.Handle("/api/venues", corsMiddlewareProduction(ipRateLimit(http.HandlerFunc(venuesHandler))))
	http.Handle("/api/courts", corsMiddlewareProduction(ipRateLimit(http.HandlerFunc(courtsHandler))))

	// System endpoints
	http.Handle("/api/health", corsMiddlewareProduction(ipRateLimit(http.HandlerFunc(healthHandler(secretsManager)))))
	http.Handle("/api/system/health", corsMiddlewareProduction(ipRateLimit(http.HandlerFunc(healthHandler(secretsManager)))))
	http.Handle("/api/system/status", corsMiddlewareProduction(ipRateLimit(http.HandlerFunc(systemStatusHandler))))
	http.Handle("/api/system/pause", corsMiddlewareProduction(ipRateLimit(http.HandlerFunc(systemPauseHandler))))
	http.Handle("/api/system/resume", corsMiddlewareProduction(ipRateLimit(http.HandlerFunc(systemResumeHandler))))
	http.Handle("/api/system/restart", corsMiddlewareProduction(ipRateLimit(http.HandlerFunc(systemRestartHandler))))

	// Rate limit status endpoint (no rate limiting for monitoring)
	http.HandleFunc("/api/system/rate-limit-status", func(w http.ResponseWriter, r *http.Request) {
		corsMiddlewareProduction(http.HandlerFunc(rateLimitStatusHandler(rateLimiter))).ServeHTTP(w, r)
	})
}

func setupRoutesWithoutRateLimit(jwtService *auth.JWTService, secretsManager *secrets.SecretsManager) {
	// Fallback setup without rate limiting but with CORS
	http.Handle("/api/auth/login", corsMiddlewareProduction(http.HandlerFunc(loginHandler(jwtService))))
	http.Handle("/api/auth/register", corsMiddlewareProduction(http.HandlerFunc(registerHandler(jwtService))))
	http.Handle("/api/auth/refresh", corsMiddlewareProduction(http.HandlerFunc(refreshHandler(jwtService))))
	http.Handle("/api/auth/logout", corsMiddlewareProduction(http.HandlerFunc(logoutHandler())))

	protectedHandler := auth.JWTMiddleware(jwtService)(http.HandlerFunc(protectedRouteHandler))
	http.Handle("/api/auth/me", corsMiddlewareProduction(protectedHandler))

	// User preferences endpoint (protected)
	preferencesHandler := auth.JWTMiddleware(jwtService)(http.HandlerFunc(userPreferencesHandler))
	http.Handle("/api/users/preferences", corsMiddlewareProduction(preferencesHandler))

	// Court and venue endpoints
	http.Handle("/api/venues", corsMiddlewareProduction(http.HandlerFunc(venuesHandler)))
	http.Handle("/api/courts", corsMiddlewareProduction(http.HandlerFunc(courtsHandler)))

	// System endpoints
	http.Handle("/api/health", corsMiddlewareProduction(http.HandlerFunc(healthHandler(secretsManager))))
	http.Handle("/api/system/health", corsMiddlewareProduction(http.HandlerFunc(healthHandler(secretsManager))))
	http.Handle("/api/system/status", corsMiddlewareProduction(http.HandlerFunc(systemStatusHandler)))
	http.Handle("/api/system/pause", corsMiddlewareProduction(http.HandlerFunc(systemPauseHandler)))
	http.Handle("/api/system/resume", corsMiddlewareProduction(http.HandlerFunc(systemResumeHandler)))
	http.Handle("/api/system/restart", corsMiddlewareProduction(http.HandlerFunc(systemRestartHandler)))

	http.HandleFunc("/api/system/rate-limit-status", func(w http.ResponseWriter, r *http.Request) {
		corsMiddlewareProduction(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"status": "disabled",
				"reason": "Redis not available",
			})
		})).ServeHTTP(w, r)
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

// Additional handler functions for complete API coverage

func registerHandler(jwtService *auth.JWTService) http.HandlerFunc {
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

		// Simple registration validation
		if req.Username == "" || req.Password == "" {
			http.Error(w, "Username and password required", http.StatusBadRequest)
			return
		}

		if len(req.Password) < 6 {
			http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
			return
		}

		// For testing, accept any valid registration
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

func logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// In a real implementation, you would invalidate the token
		response := map[string]interface{}{
			"message": "Successfully logged out",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func userPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user claims from context (set by JWT middleware)
	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		http.Error(w, "Failed to get user claims", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Return mock user preferences
		preferences := map[string]interface{}{
			"user_id": claims.UserID,
			"notification_settings": map[string]interface{}{
				"email_notifications": true,
				"push_notifications":  true,
				"sms_notifications":   false,
				"booking_reminders":   true,
				"price_alerts":        true,
				"availability_alerts": false,
			},
			"booking_settings": map[string]interface{}{
				"preferred_court_types":    []string{"Hard Court", "Clay Court"},
				"preferred_time_slots":     []string{"morning", "evening"},
				"auto_book_favorites":      false,
				"max_booking_days_ahead":   7,
				"default_booking_duration": 60,
			},
			"display_settings": map[string]interface{}{
				"theme":       "light",
				"currency":    "USD",
				"timezone":    "America/New_York",
				"date_format": "MM/DD/YYYY",
				"time_format": "12h",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(preferences)

	case http.MethodPut:
		// Update user preferences
		var preferences map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&preferences); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// In a real implementation, you would save to database
		// For testing, just return the updated preferences
		response := map[string]interface{}{
			"message":     "Preferences updated successfully",
			"preferences": preferences,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func venuesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Mock venue data
	venues := []map[string]interface{}{
		{
			"id":               1,
			"name":             "Central Tennis Club",
			"address":          "123 Main St, City, State 12345",
			"phone":            "(555) 123-4567",
			"courts_available": 6,
			"court_types":      []string{"Hard Court", "Clay Court"},
			"amenities":        []string{"Pro Shop", "Locker Rooms", "Parking"},
		},
		{
			"id":               2,
			"name":             "Riverside Tennis Center",
			"address":          "456 River Rd, City, State 12345",
			"phone":            "(555) 987-6543",
			"courts_available": 8,
			"court_types":      []string{"Hard Court", "Grass Court"},
			"amenities":        []string{"Restaurant", "Locker Rooms", "Parking", "Pool"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(venues)
}

func courtsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Mock court slot data
	now := time.Now()
	courts := []map[string]interface{}{
		{
			"id":           1,
			"venue_id":     1,
			"venue_name":   "Central Tennis Club",
			"court_number": 1,
			"court_type":   "Hard Court",
			"date":         now.Format("2006-01-02"),
			"time_slot":    "09:00",
			"duration":     60,
			"price":        25.00,
			"available":    true,
		},
		{
			"id":           2,
			"venue_id":     1,
			"venue_name":   "Central Tennis Club",
			"court_number": 2,
			"court_type":   "Clay Court",
			"date":         now.Format("2006-01-02"),
			"time_slot":    "10:00",
			"duration":     60,
			"price":        30.00,
			"available":    true,
		},
		{
			"id":           3,
			"venue_id":     2,
			"venue_name":   "Riverside Tennis Center",
			"court_number": 1,
			"court_type":   "Hard Court",
			"date":         now.Format("2006-01-02"),
			"time_slot":    "11:00",
			"duration":     60,
			"price":        20.00,
			"available":    false,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courts)
}

func systemStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := map[string]interface{}{
		"status":           "active",
		"uptime":           "24h 30m",
		"scraper_status":   "running",
		"last_update":      time.Now().Add(-5 * time.Minute).UTC(),
		"venues_monitored": 2,
		"courts_tracked":   14,
		"bookings_today":   12,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func systemPauseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"message": "System scraping paused",
		"status":  "paused",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func systemResumeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"message": "System scraping resumed",
		"status":  "active",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func systemRestartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"message": "System restart initiated",
		"status":  "restarting",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
