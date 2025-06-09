package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"tennis-booking-bot/internal/scheduler"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Create scheduler configuration
	config := &scheduler.SchedulerConfig{
		MongoURI:         getEnvWithDefault("MONGO_URI", "mongodb://admin:YOUR_PASSWORD@localhost:27017"),
		DatabaseName:     getEnvWithDefault("MONGO_DATABASE", "tennis_booking"),
		RedisAddr:        getEnvWithDefault("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    getEnvWithDefault("REDIS_PASSWORD", ""),
		RedisDB:          0,
		TaskQueueName:    getEnvWithDefault("TASK_QUEUE_NAME", "tennis_scraping_tasks"),
		DefaultInterval:  getEnvWithDefault("DEFAULT_SCRAPING_INTERVAL", "0 */30 * * * *"), // Every 30 minutes
	}

	// Create and start scheduler
	s := scheduler.NewScheduler(config)

	if err := s.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}

	// Wait for interrupt signal to gracefully shut down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	log.Println("🚀 Tennis Court Booking Scheduler is running")
	log.Println("📊 Schedules:")
	log.Println("   • Regular scraping: Every 30 minutes")
	log.Println("   • High-priority scraping: Every 15 minutes") 
	log.Println("   • Health checks: Every 5 minutes")
	log.Println("Press Ctrl+C to stop...")

	<-quit
	log.Println("\n🛑 Shutdown signal received...")

	// Gracefully stop the scheduler
	if err := s.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("✅ Scheduler shutdown complete")
}

// getEnvWithDefault returns the environment variable value or a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 