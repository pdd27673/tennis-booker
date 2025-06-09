package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"tennis-booking-bot/internal/database"
	"tennis-booking-bot/internal/handlers"
	"tennis-booking-bot/internal/models"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Get database configuration from environment
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:YOUR_PASSWORD@localhost:27017"
		log.Println("Using default MongoDB URI")
	}

	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "tennis_booking"
		log.Println("Using default database name: tennis_booking")
	}

	// Initialize MongoDB connection
	log.Println("Connecting to MongoDB...")
	db, err := database.InitDatabase(mongoURI, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB successfully")

	// Create all database indexes
	log.Println("Creating database indexes...")
	if err := database.CreateAllIndexes(db); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	// Initialize services
	preferenceService := models.NewPreferenceService(db)

	// Initialize handlers
	preferenceHandlers := handlers.NewPreferenceHandlers(preferenceService)
	systemHandlers := handlers.NewSystemHandlers(db)
	venueHandlers := handlers.NewVenueHandlers(db)
	alertHandlers := handlers.NewAlertHandlers(db)
	courtsHandlers := handlers.NewCourtsHandlers(db)

	// Initialize Gin router
	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add CORS middleware for development
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// System endpoints
	api := router.Group("/api")
	{
		api.GET("/health", systemHandlers.HealthCheck)
		api.GET("/metrics", systemHandlers.Metrics)
	}

	// API v1 routes
	v1 := api.Group("/v1")
	{
		// Preference routes
		preferences := v1.Group("/preferences")
		{
			preferences.GET("", preferenceHandlers.GetPreferences)
			preferences.PUT("", preferenceHandlers.UpdatePreferences)
			preferences.POST("/venues", preferenceHandlers.AddPreferredVenue)
			preferences.DELETE("/venues/:venueId", preferenceHandlers.RemovePreferredVenue)
		}

		// Venue routes
		venues := v1.Group("/venues")
		{
			venues.GET("", venueHandlers.GetVenues)
			venues.GET("/:id/slots", venueHandlers.GetVenueSlots)
		}

		// Alert routes
		alerts := v1.Group("/alerts")
		{
			alerts.GET("/history", alertHandlers.GetAlertHistory)
			alerts.GET("/stats", alertHandlers.GetAlertStats)
		}

		// Courts routes
		courts := v1.Group("/courts")
		{
			courts.GET("/available", courtsHandlers.GetAvailableCourts)
		}
	}

	// Create HTTP server
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Tennis Court API server starting on port %s", port)
		log.Printf("📋 Available endpoints:")
		log.Printf("  🏥 GET    /api/health")
		log.Printf("  📊 GET    /api/metrics")
		log.Printf("  ⚙️  GET    /api/v1/preferences?user_id=<user_id>")
		log.Printf("  ⚙️  PUT    /api/v1/preferences?user_id=<user_id>")
		log.Printf("  ⚙️  POST   /api/v1/preferences/venues?user_id=<user_id>")
		log.Printf("  ⚙️  DELETE /api/v1/preferences/venues/:venueId?user_id=<user_id>&list_type=<list_type>")
		log.Printf("  🏟️  GET    /api/v1/venues?provider=<provider>")
		log.Printf("  🏟️  GET    /api/v1/venues/<venue_id>/slots?date_from=<YYYY-MM-DD>&date_to=<YYYY-MM-DD>")
		log.Printf("  🔔 GET    /api/v1/alerts/history?user_id=<user_id>&limit=<limit>&offset=<offset>")
		log.Printf("  📈 GET    /api/v1/alerts/stats?user_id=<user_id>")
		log.Printf("  🎾 GET    /api/v1/courts/available?venue_ids=<ids>&provider=<provider>&date_from=<date>&time_from=<time>&price_min=<min>&limit=<limit>")
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
} 