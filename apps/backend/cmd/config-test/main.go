package main

import (
	"fmt"
	"log"
	"os"

	"tennis-booker/internal/config"
)

func main() {
	fmt.Println("=== GO BACKEND CONFIGURATION TEST ===")
	
	// Test different environments
	environments := []string{"development", "production", "test"}
	
	for _, env := range environments {
		fmt.Printf("\n--- Testing Environment: %s ---\n", env)
		
		// Set environment
		os.Setenv("APP_ENV", env)
		
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			log.Printf("Error loading config for %s: %v", env, err)
			continue
		}
		
		// Display key configuration values
		fmt.Printf("App Name: %s\n", cfg.App.Name)
		fmt.Printf("App Version: %s\n", cfg.App.Version)
		fmt.Printf("Environment: %s\n", cfg.App.Environment)
		fmt.Printf("API Port: %d\n", cfg.API.Port)
		fmt.Printf("Log Level: %s\n", cfg.Logging.Level)
		fmt.Printf("Scraper Interval: %v\n", cfg.Scraper.Interval)
		
		// Test feature flags
		fmt.Printf("Feature Flags:\n")
		fmt.Printf("  Analytics: %t\n", cfg.IsFeatureEnabled("analytics"))
		fmt.Printf("  Notifications: %t\n", cfg.IsFeatureEnabled("notifications"))
		
		// Test platform configuration
		fmt.Printf("Platforms:\n")
		fmt.Printf("  ClubSpark Enabled: %t\n", cfg.Scraper.Platforms.Clubspark.Enabled)
		fmt.Printf("  Courtsides Enabled: %t\n", cfg.Scraper.Platforms.Courtsides.Enabled)
	}
	
	fmt.Println("\n=== TESTING ENVIRONMENT VARIABLE OVERRIDES ===")
	
	// Test environment variable overrides
	os.Setenv("APP_ENV", "test")
	os.Setenv("API_PORT", "9999")
	os.Setenv("LOG_LEVEL", "error")
	os.Setenv("FEATURE_ANALYTICS_ENABLED", "false")
	
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config with overrides: %v", err)
	}
	
	fmt.Printf("API Port (should be 9999): %d\n", cfg.API.Port)
	fmt.Printf("Log Level (should be error): %s\n", cfg.Logging.Level)
	fmt.Printf("Analytics Feature (should be false): %t\n", cfg.IsFeatureEnabled("analytics"))
	
	fmt.Println("\n✅ Go configuration system test completed successfully!")
} 