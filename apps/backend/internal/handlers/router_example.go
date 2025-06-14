package handlers

import (
	"net/http"

	"tennis-booker/internal/auth"
	"tennis-booker/internal/database"
	"tennis-booker/internal/models"
	"tennis-booker/internal/ratelimit"
)

// SetupRoutes demonstrates how to wire up all application routes with rate limiting
// This is an example function showing how to integrate all handlers with appropriate rate limiting
func SetupRoutes(mux *http.ServeMux, userService models.UserService, jwtService *auth.JWTService, refreshTokenService models.RefreshTokenService, venueRepo *database.VenueRepository, scrapingLogRepo *database.ScrapingLogRepository, rateLimiter *ratelimit.Limiter, version string) {
	// Create handlers
	authHandler := NewAuthHandler(userService, jwtService, refreshTokenService)
	systemHandler := NewSystemHandler(version)
	userHandler := NewUserHandler(userService)
	courtHandler := NewCourtHandlerWithDB(venueRepo, scrapingLogRepo)

	// Create JWT middleware
	jwtMiddleware := auth.JWTMiddleware(jwtService)

	// Create rate limiting middleware instances
	ipRateLimit := ratelimit.IPRateLimitMiddleware(rateLimiter)
	authRateLimit := ratelimit.AuthRateLimitMiddleware(rateLimiter)
	dataRateLimit := ratelimit.DataRateLimitMiddleware(rateLimiter)
	sensitiveRateLimit := ratelimit.SensitiveRateLimitMiddleware(rateLimiter)
	combinedRateLimit := ratelimit.CombinedRateLimitMiddleware(rateLimiter)

	// Public routes with IP-based rate limiting
	// Health endpoint - basic IP limiting
	mux.Handle("/api/health", ipRateLimit(http.HandlerFunc(systemHandler.Health)))

	// Authentication endpoints - strict IP-based rate limiting to prevent brute force
	mux.Handle("/auth/register", authRateLimit(http.HandlerFunc(authHandler.Register)))
	mux.Handle("/auth/login", authRateLimit(http.HandlerFunc(authHandler.Login)))
	mux.Handle("/auth/refresh", authRateLimit(http.HandlerFunc(authHandler.RefreshToken)))
	mux.Handle("/auth/logout", authRateLimit(http.HandlerFunc(authHandler.Logout)))

	// Protected routes with combined rate limiting (both IP and user-based)
	// User profile endpoints - moderate limits
	mux.Handle("/api/users/me", combinedRateLimit(jwtMiddleware(http.HandlerFunc(authHandler.Me))))
	mux.Handle("/api/users/preferences", combinedRateLimit(jwtMiddleware(http.HandlerFunc(userHandler.UpdatePreferences))))

	// System management endpoints - sensitive operations with strict limits
	mux.Handle("/api/system/status", sensitiveRateLimit(jwtMiddleware(http.HandlerFunc(systemHandler.Status))))
	mux.Handle("/api/system/pause", sensitiveRateLimit(jwtMiddleware(http.HandlerFunc(systemHandler.Pause))))
	mux.Handle("/api/system/resume", sensitiveRateLimit(jwtMiddleware(http.HandlerFunc(systemHandler.Resume))))

	// Data-intensive endpoints - moderate rate limiting for data access
	mux.Handle("/api/venues", dataRateLimit(jwtMiddleware(http.HandlerFunc(courtHandler.ListVenues))))
	mux.Handle("/api/courts", dataRateLimit(jwtMiddleware(http.HandlerFunc(courtHandler.ListCourts))))

	// Example of how to protect other routes with different rate limiting strategies:
	// mux.Handle("/api/protected", userRateLimit(jwtMiddleware(http.HandlerFunc(someOtherHandler))))
	// mux.Handle("/api/bulk-export", sensitiveRateLimit(jwtMiddleware(http.HandlerFunc(bulkExportHandler))))
}

// SetupRoutesWithCustomLimits demonstrates how to apply custom rate limits to specific endpoints
func SetupRoutesWithCustomLimits(mux *http.ServeMux, userService models.UserService, jwtService *auth.JWTService, refreshTokenService models.RefreshTokenService, venueRepo *database.VenueRepository, scrapingLogRepo *database.ScrapingLogRepository, rateLimiter *ratelimit.Limiter, version string) {
	// Create handlers
	authHandler := NewAuthHandler(userService, jwtService, refreshTokenService)
	systemHandler := NewSystemHandler(version)
	userHandler := NewUserHandler(userService)
	courtHandler := NewCourtHandlerWithDB(venueRepo, scrapingLogRepo)

	// Create JWT middleware
	jwtMiddleware := auth.JWTMiddleware(jwtService)

	// Create custom rate limiting middleware for specific endpoints

	// Very strict rate limiting for password reset (2 requests per hour)
	passwordResetLimit := ratelimit.CustomRateLimitMiddleware(rateLimiter, ratelimit.RateLimit{
		Requests: 2,
		Window:   3600, // 1 hour
	}, "password-reset")

	// Moderate rate limiting for data export (10 requests per hour)
	dataExportLimit := ratelimit.CustomRateLimitMiddleware(rateLimiter, ratelimit.RateLimit{
		Requests: 10,
		Window:   3600, // 1 hour
	}, "data-export")

	// Relaxed rate limiting for read-only operations (1000 requests per minute)
	readOnlyLimit := ratelimit.CustomRateLimitMiddleware(rateLimiter, ratelimit.RateLimit{
		Requests: 1000,
		Window:   60, // 1 minute
	}, "read-only")

	// Apply custom limits to specific endpoints
	// Example: Password reset with very strict limits
	mux.Handle("/auth/password-reset", passwordResetLimit(http.HandlerFunc(authHandler.Register))) // Using Register as placeholder

	// Example: Data export with moderate limits
	mux.Handle("/api/export/venues", dataExportLimit(jwtMiddleware(http.HandlerFunc(courtHandler.ListVenues)))) // Using ListVenues as placeholder

	// Example: Read-only operations with relaxed limits
	mux.Handle("/api/venues/search", readOnlyLimit(jwtMiddleware(http.HandlerFunc(courtHandler.ListVenues)))) // Using ListVenues as placeholder

	// Standard routes with default rate limiting
	setupStandardRoutes(mux, authHandler, systemHandler, userHandler, courtHandler, jwtMiddleware, rateLimiter)
}

// setupStandardRoutes is a helper function to set up routes with standard rate limiting
func setupStandardRoutes(mux *http.ServeMux, authHandler *AuthHandler, systemHandler *SystemHandler, userHandler *UserHandler, courtHandler *CourtHandler, jwtMiddleware func(http.Handler) http.Handler, rateLimiter *ratelimit.Limiter) {
	// Create standard rate limiting middleware instances
	ipRateLimit := ratelimit.IPRateLimitMiddleware(rateLimiter)
	authRateLimit := ratelimit.AuthRateLimitMiddleware(rateLimiter)
	combinedRateLimit := ratelimit.CombinedRateLimitMiddleware(rateLimiter)
	dataRateLimit := ratelimit.DataRateLimitMiddleware(rateLimiter)
	sensitiveRateLimit := ratelimit.SensitiveRateLimitMiddleware(rateLimiter)

	// Public routes
	mux.Handle("/api/health", ipRateLimit(http.HandlerFunc(systemHandler.Health)))

	// Authentication endpoints
	mux.Handle("/auth/register", authRateLimit(http.HandlerFunc(authHandler.Register)))
	mux.Handle("/auth/login", authRateLimit(http.HandlerFunc(authHandler.Login)))
	mux.Handle("/auth/refresh", authRateLimit(http.HandlerFunc(authHandler.RefreshToken)))
	mux.Handle("/auth/logout", authRateLimit(http.HandlerFunc(authHandler.Logout)))

	// Protected routes
	mux.Handle("/api/users/me", combinedRateLimit(jwtMiddleware(http.HandlerFunc(authHandler.Me))))
	mux.Handle("/api/users/preferences", combinedRateLimit(jwtMiddleware(http.HandlerFunc(userHandler.UpdatePreferences))))

	// System management
	mux.Handle("/api/system/status", sensitiveRateLimit(jwtMiddleware(http.HandlerFunc(systemHandler.Status))))
	mux.Handle("/api/system/pause", sensitiveRateLimit(jwtMiddleware(http.HandlerFunc(systemHandler.Pause))))
	mux.Handle("/api/system/resume", sensitiveRateLimit(jwtMiddleware(http.HandlerFunc(systemHandler.Resume))))

	// Data endpoints
	mux.Handle("/api/venues", dataRateLimit(jwtMiddleware(http.HandlerFunc(courtHandler.ListVenues))))
	mux.Handle("/api/courts", dataRateLimit(jwtMiddleware(http.HandlerFunc(courtHandler.ListCourts))))
}

// Example usage in main.go or server setup:
//
// func main() {
//     // Setup database connection
//     db, err := database.InitDatabase("mongodb://localhost:27017", "tennis_booking")
//     if err != nil {
//         log.Fatal(err)
//     }
//
//     // Setup services
//     userService := models.NewInMemoryUserService()
//     secretsProvider := &vault.VaultSecretsProvider{...}
//     jwtService := auth.NewJWTService(secretsProvider, "your-issuer")
//     refreshTokenService := models.NewMongoRefreshTokenService(database)
//     venueRepo := database.NewVenueRepository(db)
//     scrapingLogRepo := database.NewScrapingLogRepository(db)
//
//     // Setup rate limiter
//     rateLimitConfig := ratelimit.DefaultConfig()
//     rateLimiter, err := ratelimit.NewLimiter(rateLimitConfig)
//     if err != nil {
//         log.Fatal("Failed to create rate limiter:", err)
//     }
//     defer rateLimiter.Close()
//
//     // Setup routes with rate limiting
//     mux := http.NewServeMux()
//     SetupRoutes(mux, userService, jwtService, refreshTokenService, venueRepo, scrapingLogRepo, rateLimiter, "1.0.0")
//
//     // Start server
//     log.Println("Server starting on :8080")
//     log.Fatal(http.ListenAndServe(":8080", mux))
// }
