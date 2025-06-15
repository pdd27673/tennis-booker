package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tennis-booker/internal/auth"
	"tennis-booker/internal/config"
	"tennis-booker/internal/database"
	"tennis-booker/internal/handlers"
	"tennis-booker/internal/middleware"
	"tennis-booker/internal/secrets"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize secrets manager
	secretsManager, err := secrets.NewSecretsManagerFromEnv()
	if err != nil {
		log.Fatalf("Failed to create secrets manager: %v", err)
	}
	defer secretsManager.Close()

	// Initialize database
	mongoDb, err := database.InitDatabase(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	// Wrap in our Database interface
	db := database.NewMongoDB(mongoDb)

	// Initialize JWT service
	jwtService := auth.NewJWTService(secretsManager, cfg.JWT.Issuer)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(jwtService, db)
	courtHandler := handlers.NewCourtHandler(db)
	userHandler := handlers.NewUserHandler(db, jwtService)
	systemHandler := handlers.NewSystemHandler(db)
	healthHandler := handlers.NewHealthHandler(secretsManager, db)

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
		log.Printf("🚀 Tennis Booker API Server starting on port %s", cfg.Server.Port)
		log.Printf("🌐 CORS enabled for origins: %v", cfg.CORS.AllowedOrigins)
		log.Printf("📋 API endpoints available at http://localhost:%s/api/", cfg.Server.Port)
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("🛑 Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
} 