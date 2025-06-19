package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"tennis-booker/internal/auth"
	"tennis-booker/internal/config"
	"tennis-booker/internal/database"
	"tennis-booker/internal/handlers"
	"tennis-booker/internal/logging"
	"tennis-booker/internal/middleware"
	"tennis-booker/internal/secrets"
)

// FallbackJWTProvider provides JWT secrets from environment variables as fallback
type FallbackJWTProvider struct{}

// GetJWTSecret returns a JWT secret from environment variables
func (f *FallbackJWTProvider) GetJWTSecret() (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET environment variable is required - cannot use insecure default")
	}
	return secret, nil
}

func main() {
	// Initialize structured logging
	logger := logging.New("tennis-server")
	
	// Load environment variables if .env file exists
	if err := godotenv.Load(); err != nil {
		logger.Debug("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", map[string]interface{}{"error": err.Error()})
	}

	// Initialize database connection with fallback
	var mongoDb database.Database
	var secretsManager *secrets.SecretsManager
	
	// Try to initialize secrets manager and database connection
	sm, err := secrets.NewSecretsManagerFromEnv()
	if err != nil {
		logger.Warn("Failed to create secrets manager", map[string]interface{}{"error": err.Error()})
		logger.Info("Using fallback database connection")
		
		// Fallback to direct database connection using config
		mongoURI := cfg.MongoDB.URI
		if mongoURI == "" {
			// Build URI from config components
			if cfg.MongoDB.Username != "" && cfg.MongoDB.Password != "" {
				mongoURI = fmt.Sprintf("mongodb://%s:%s@%s:%s?authSource=admin", 
					cfg.MongoDB.Username, cfg.MongoDB.Password, cfg.MongoDB.Host, cfg.MongoDB.Port)
			} else {
				mongoURI = fmt.Sprintf("mongodb://%s:%s", cfg.MongoDB.Host, cfg.MongoDB.Port)
			}
		}
		
		mongoDbInstance, err := database.InitDatabase(mongoURI, cfg.MongoDB.Database)
		mongoDb = database.NewMongoDB(mongoDbInstance)
		if err != nil {
			logger.Fatal("Failed to connect to database with fallback", map[string]interface{}{"error": err.Error()})
		}
		logger.ConnectionInfo("Connected to database using fallback credentials", "mongodb", cfg.MongoDB.Host)
	} else {
		secretsManager = sm
		defer secretsManager.Close()
		
		// Use connection manager for Vault-based connection
		connectionManager := database.NewConnectionManager(secretsManager)
		mongoDbInstance, err := connectionManager.ConnectWithFallback()
		if err != nil {
			logger.Fatal("Failed to connect to database", map[string]interface{}{"error": err.Error()})
		}
		mongoDb = database.NewMongoDB(mongoDbInstance)
		logger.ConnectionInfo("Connected to database using Vault credentials", "mongodb", cfg.MongoDB.Host)
	}

	// Initialize JWT service
	var jwtService *auth.JWTService
	if secretsManager != nil {
		jwtService = auth.NewJWTService(secretsManager, cfg.JWT.Issuer)
	} else {
		// Create a fallback JWT service using environment variables
		fallbackProvider := &FallbackJWTProvider{}
		jwtService = auth.NewJWTService(fallbackProvider, cfg.JWT.Issuer)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(jwtService, mongoDb)
	courtHandler := handlers.NewCourtHandler(mongoDb)
	userHandler := handlers.NewUserHandler(mongoDb, jwtService)
	systemHandler := handlers.NewSystemHandler(mongoDb)
	healthHandler := handlers.NewHealthHandler(secretsManager, mongoDb)

	// Setup router
	router := mux.NewRouter()

	// CORS middleware
	router.Use(middleware.CORSMiddleware())

	// Health endpoints
	router.HandleFunc("/api/health", healthHandler.Health).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/system/health", healthHandler.SystemHealth).Methods("GET", "OPTIONS")

	// Auth endpoints
	authRouter := router.PathPrefix("/api/auth").Subrouter()
	authRouter.HandleFunc("/login", authHandler.Login).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/register", authHandler.Register).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/refresh", authHandler.RefreshToken).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/logout", authHandler.Logout).Methods("POST", "OPTIONS")

	// Protected auth endpoints
	protectedAuthRouter := authRouter.PathPrefix("").Subrouter()
	protectedAuthRouter.Use(middleware.JWTMiddleware(jwtService))
	protectedAuthRouter.HandleFunc("/me", authHandler.GetCurrentUser).Methods("GET", "OPTIONS")

	// User endpoints
	userRouter := router.PathPrefix("/api/users").Subrouter()
	userRouter.Use(middleware.JWTMiddleware(jwtService))
	userRouter.HandleFunc("/preferences", userHandler.GetPreferences).Methods("GET", "OPTIONS")
	userRouter.HandleFunc("/preferences", userHandler.UpdatePreferences).Methods("PUT", "OPTIONS")
	
	// Court endpoints
	courtRouter := router.PathPrefix("/api").Subrouter()
	courtRouter.HandleFunc("/venues", courtHandler.GetVenues).Methods("GET", "OPTIONS")
	courtRouter.HandleFunc("/courts", courtHandler.GetCourtSlots).Methods("GET", "OPTIONS")
	courtRouter.HandleFunc("/dashboard/stats", courtHandler.GetDashboardStats).Methods("GET", "OPTIONS")

	// System endpoints
	systemRouter := router.PathPrefix("/api/system").Subrouter()
	systemRouter.HandleFunc("/status", systemHandler.GetStatus).Methods("GET", "OPTIONS")
	systemRouter.HandleFunc("/logs", systemHandler.GetScrapingLogs).Methods("GET", "OPTIONS")
	systemRouter.HandleFunc("/pause", systemHandler.PauseScraping).Methods("POST", "OPTIONS")
	systemRouter.HandleFunc("/resume", systemHandler.ResumeScraping).Methods("POST", "OPTIONS")
	systemRouter.HandleFunc("/restart", systemHandler.RestartSystem).Methods("POST", "OPTIONS")

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		logger.StartupInfo("🚀 Tennis Booker API Server starting", cfg.Server.Port, "production")
		logger.Info("API endpoints available", map[string]interface{}{
			"base_url": fmt.Sprintf("http://localhost:%s/api/", cfg.Server.Port),
			"cors_origins": cfg.CORS.AllowedOrigins,
		})
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", map[string]interface{}{"error": err.Error()})
		}
	}()

	// Wait for interrupt signal
	<-quit
	logger.ShutdownInfo("Shutting down server", "signal_received")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", map[string]interface{}{"error": err.Error()})
	}

	logger.Info("Server stopped gracefully")
} 