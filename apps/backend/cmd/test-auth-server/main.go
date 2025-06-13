package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"tennis-booker/internal/auth"
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

	// Set up routes
	http.HandleFunc("/login", loginHandler(jwtService))
	http.HandleFunc("/refresh", refreshHandler(jwtService))
	
	// Protected route with middleware
	protectedHandler := auth.JWTMiddleware(jwtService)(http.HandlerFunc(protectedRouteHandler))
	http.Handle("/protected", protectedHandler)

	// Health check
	http.HandleFunc("/health", healthHandler(secretsManager))

	fmt.Println("🚀 Test Auth Server starting on :8080")
	fmt.Println("📋 Available endpoints:")
	fmt.Println("  POST /login - Login with username/password")
	fmt.Println("  POST /refresh - Refresh access token")
	fmt.Println("  GET /protected - Protected route (requires Bearer token)")
	fmt.Println("  GET /health - Health check")
	fmt.Println()
	fmt.Println("🧪 Test commands:")
	fmt.Println("  curl -X POST http://localhost:8080/login -H 'Content-Type: application/json' -d '{\"username\":\"testuser\",\"password\":\"testpass\"}'")
	fmt.Println("  curl -H 'Authorization: Bearer <token>' http://localhost:8080/protected")

	log.Fatal(http.ListenAndServe(":8080", nil))
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