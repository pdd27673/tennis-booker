package handlers

import (
	"net/http"

	"tennis-booker/internal/auth"
	"tennis-booker/internal/models"
)

// SetupAuthRoutes demonstrates how to wire up authentication routes
// This is an example function showing how to integrate the auth handlers
func SetupAuthRoutes(mux *http.ServeMux, userService models.UserService, jwtService *auth.JWTService, refreshTokenService models.RefreshTokenService) {
	// Create auth handler
	authHandler := NewAuthHandler(userService, jwtService, refreshTokenService)

	// Create JWT middleware
	jwtMiddleware := auth.JWTMiddleware(jwtService)

	// Public routes (no authentication required)
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.HandleFunc("/auth/refresh", authHandler.RefreshToken)
	mux.HandleFunc("/auth/logout", authHandler.Logout)

	// Protected routes (authentication required)
	mux.Handle("/api/me", jwtMiddleware(http.HandlerFunc(authHandler.Me)))

	// Example of how to protect other routes
	// mux.Handle("/api/protected", jwtMiddleware(http.HandlerFunc(someOtherHandler)))
}

// Example usage in main.go or server setup:
//
// func main() {
//     // Setup services
//     userService := models.NewInMemoryUserService()
//     secretsProvider := &vault.VaultSecretsProvider{...}
//     jwtService := auth.NewJWTService(secretsProvider, "your-issuer")
//     refreshTokenService := models.NewMongoRefreshTokenService(database)
//
//     // Setup routes
//     mux := http.NewServeMux()
//     SetupAuthRoutes(mux, userService, jwtService, refreshTokenService)
//
//     // Start server
//     log.Println("Server starting on :8080")
//     log.Fatal(http.ListenAndServe(":8080", mux))
// } 