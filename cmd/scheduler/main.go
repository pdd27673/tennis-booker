package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
)

func main() {
	// Create a new cron scheduler
	c := cron.New(cron.WithSeconds())

	// Add a simple job that runs every minute
	_, err := c.AddFunc("0 * * * * *", func() {
		log.Println("Running scheduled task - this will be replaced with real scraping tasks")
	})
	if err != nil {
		log.Fatalf("Error adding cron job: %v", err)
	}

	// Start the scheduler
	c.Start()
	log.Println("Scheduler service started")

	// Wait for interrupt signal to gracefully shut down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down scheduler...")

	// Create a context with timeout for cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Stop the scheduler (context not used here but would be for other cleanup)
	c.Stop()
	_ = ctx // Just to avoid unused variable warning

	log.Println("Scheduler stopped gracefully")
} 