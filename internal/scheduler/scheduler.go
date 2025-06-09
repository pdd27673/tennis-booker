package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ScrapingTask represents a task to be queued for the Python scraper
type ScrapingTask struct {
	VenueID     string    `json:"venue_id"`
	VenueName   string    `json:"venue_name"`
	VenueURL    string    `json:"venue_url"`
	Provider    string    `json:"provider"`
	TaskID      string    `json:"task_id"`
	QueuedAt    time.Time `json:"queued_at"`
	Priority    string    `json:"priority,omitempty"`
	TargetDate  string    `json:"target_date,omitempty"`
}

// Venue represents a venue from MongoDB
type Venue struct {
	ID               string    `bson:"_id" json:"id"`
	Name             string    `bson:"name" json:"name"`
	Provider         string    `bson:"provider" json:"provider"`
	URL              string    `bson:"url" json:"url"`
	IsActive         bool      `bson:"is_active" json:"is_active"`
	ScrapingInterval int       `bson:"scraping_interval" json:"scraping_interval"` // Minutes
	LastScrapedAt    time.Time `bson:"last_scraped_at,omitempty" json:"last_scraped_at,omitempty"`
}

// SchedulerConfig holds configuration for the scheduler
type SchedulerConfig struct {
	MongoURI         string
	DatabaseName     string
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
	TaskQueueName    string
	DefaultInterval  string // Cron expression for default scraping
}

// Scheduler manages cron jobs and task queuing
type Scheduler struct {
	config       *SchedulerConfig
	cron         *cron.Cron
	redisClient  *redis.Client
	mongoClient  *mongo.Client
	db           *mongo.Database
	logger       *log.Logger
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewScheduler creates a new scheduler instance
func NewScheduler(config *SchedulerConfig) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Scheduler{
		config: config,
		cron:   cron.New(cron.WithSeconds()),
		logger: log.New(log.Writer(), "[SCHEDULER] ", log.LstdFlags|log.Lshortfile),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start initializes connections and starts the scheduler
func (s *Scheduler) Start() error {
	s.logger.Println("Starting Tennis Court Scraper Scheduler...")

	// Initialize Redis connection
	if err := s.initRedis(); err != nil {
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Initialize MongoDB connection
	if err := s.initMongoDB(); err != nil {
		return fmt.Errorf("failed to initialize MongoDB: %w", err)
	}

	// Add cron jobs
	if err := s.setupCronJobs(); err != nil {
		return fmt.Errorf("failed to setup cron jobs: %w", err)
	}

	// Start the cron scheduler
	s.cron.Start()
	s.logger.Println("✅ Scheduler started successfully")

	return nil
}

// Stop gracefully shuts down the scheduler
func (s *Scheduler) Stop() error {
	s.logger.Println("Stopping scheduler...")

	// Cancel context
	s.cancel()

	// Stop cron scheduler
	cronCtx := s.cron.Stop()
	<-cronCtx.Done()

	// Close Redis connection
	if s.redisClient != nil {
		if err := s.redisClient.Close(); err != nil {
			s.logger.Printf("Error closing Redis connection: %v", err)
		}
	}

	// Close MongoDB connection
	if s.mongoClient != nil {
		if err := s.mongoClient.Disconnect(context.Background()); err != nil {
			s.logger.Printf("Error closing MongoDB connection: %v", err)
		}
	}

	s.logger.Println("✅ Scheduler stopped gracefully")
	return nil
}

// initRedis establishes connection to Redis
func (s *Scheduler) initRedis() error {
	s.logger.Printf("Connecting to Redis at %s...", s.config.RedisAddr)

	s.redisClient = redis.NewClient(&redis.Options{
		Addr:     s.config.RedisAddr,
		Password: s.config.RedisPassword,
		DB:       s.config.RedisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	_, err := s.redisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis ping failed: %w", err)
	}

	s.logger.Println("✅ Connected to Redis")
	return nil
}

// initMongoDB establishes connection to MongoDB
func (s *Scheduler) initMongoDB() error {
	s.logger.Printf("Connecting to MongoDB at %s...", s.config.MongoURI)

	clientOptions := options.Client().ApplyURI(s.config.MongoURI)

	var err error
	s.mongoClient, err = mongo.Connect(s.ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("MongoDB connection failed: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	err = s.mongoClient.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("MongoDB ping failed: %w", err)
	}

	s.db = s.mongoClient.Database(s.config.DatabaseName)
	s.logger.Println("✅ Connected to MongoDB")
	return nil
}

// setupCronJobs adds all scheduled jobs
func (s *Scheduler) setupCronJobs() error {
	s.logger.Println("Setting up cron jobs...")

	// Primary scraping job - runs every 30 minutes
	_, err := s.cron.AddFunc("0 */30 * * * *", func() {
		s.logger.Println("🎯 Running scheduled venue scraping...")
		if err := s.scheduleVenueScraping(); err != nil {
			s.logger.Printf("❌ Error in scheduled scraping: %v", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to add primary scraping job: %w", err)
	}

	// High-priority venues scraping - runs every 15 minutes
	_, err = s.cron.AddFunc("0 */15 * * * *", func() {
		s.logger.Println("⚡ Running high-priority venue scraping...")
		if err := s.scheduleHighPriorityVenueScraping(); err != nil {
			s.logger.Printf("❌ Error in high-priority scraping: %v", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to add high-priority scraping job: %w", err)
	}

	// Health check job - runs every 5 minutes
	_, err = s.cron.AddFunc("0 */5 * * * *", func() {
		s.performHealthCheck()
	})
	if err != nil {
		return fmt.Errorf("failed to add health check job: %w", err)
	}

	s.logger.Printf("✅ Added %d cron jobs", len(s.cron.Entries()))
	return nil
}

// scheduleVenueScraping fetches venues and queues scraping tasks
func (s *Scheduler) scheduleVenueScraping() error {
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	// Fetch active venues from MongoDB
	venues, err := s.fetchActiveVenues(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch venues: %w", err)
	}

	if len(venues) == 0 {
		s.logger.Println("⚠️  No active venues found")
		return nil
	}

	s.logger.Printf("📋 Found %d active venues to schedule", len(venues))

	// Queue tasks for each venue
	tasksQueued := 0
	for _, venue := range venues {
		// Check if venue needs scraping based on its interval
		if s.shouldScrapeVenue(venue) {
			if err := s.queueScrapingTask(venue, "normal"); err != nil {
				s.logger.Printf("❌ Failed to queue task for venue %s: %v", venue.Name, err)
				continue
			}
			tasksQueued++
		} else {
			s.logger.Printf("⏭️  Skipping %s (not due for scraping)", venue.Name)
		}
	}

	s.logger.Printf("✅ Queued %d scraping tasks", tasksQueued)
	return nil
}

// scheduleHighPriorityVenueScraping handles frequent scraping for high-priority venues
func (s *Scheduler) scheduleHighPriorityVenueScraping() error {
	ctx, cancel := context.WithTimeout(s.ctx, 15*time.Second)
	defer cancel()

	// Fetch venues that need frequent scraping (e.g., scraping_interval <= 15 minutes)
	filter := bson.M{
		"is_active": true,
		"scraping_interval": bson.M{"$lte": 15},
	}

	cursor, err := s.db.Collection("venues").Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to query high-priority venues: %w", err)
	}
	defer cursor.Close(ctx)

	var venues []Venue
	if err := cursor.All(ctx, &venues); err != nil {
		return fmt.Errorf("failed to decode venues: %w", err)
	}

	if len(venues) == 0 {
		s.logger.Println("📊 No high-priority venues found")
		return nil
	}

	s.logger.Printf("⚡ Found %d high-priority venues", len(venues))

	tasksQueued := 0
	for _, venue := range venues {
		if s.shouldScrapeVenue(venue) {
			if err := s.queueScrapingTask(venue, "high"); err != nil {
				s.logger.Printf("❌ Failed to queue high-priority task for venue %s: %v", venue.Name, err)
				continue
			}
			tasksQueued++
		}
	}

	s.logger.Printf("✅ Queued %d high-priority tasks", tasksQueued)
	return nil
}

// fetchActiveVenues retrieves all active venues from MongoDB
func (s *Scheduler) fetchActiveVenues(ctx context.Context) ([]Venue, error) {
	filter := bson.M{"is_active": true}
	
	cursor, err := s.db.Collection("venues").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var venues []Venue
	if err := cursor.All(ctx, &venues); err != nil {
		return nil, err
	}

	return venues, nil
}

// shouldScrapeVenue determines if a venue needs scraping based on its interval
func (s *Scheduler) shouldScrapeVenue(venue Venue) bool {
	if venue.ScrapingInterval <= 0 {
		venue.ScrapingInterval = 30 // Default to 30 minutes
	}

	// If never scraped, scrape it
	if venue.LastScrapedAt.IsZero() {
		return true
	}

	// Check if enough time has passed since last scrape
	timeSinceLastScrape := time.Since(venue.LastScrapedAt)
	requiredInterval := time.Duration(venue.ScrapingInterval) * time.Minute

	return timeSinceLastScrape >= requiredInterval
}

// queueScrapingTask adds a scraping task to the Redis queue
func (s *Scheduler) queueScrapingTask(venue Venue, priority string) error {
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	task := ScrapingTask{
		VenueID:    venue.ID,
		VenueName:  venue.Name,
		VenueURL:   venue.URL,
		Provider:   venue.Provider,
		TaskID:     fmt.Sprintf("%s_%d", venue.ID, time.Now().Unix()),
		QueuedAt:   time.Now(),
		Priority:   priority,
		TargetDate: time.Now().AddDate(0, 0, 1).Format("2006-01-02"), // Tomorrow
	}

	taskJSON, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Use different queues for different priorities
	queueName := s.config.TaskQueueName
	if priority == "high" {
		queueName = s.config.TaskQueueName + ":high"
	}

	// Push task to Redis list (queue)
	err = s.redisClient.LPush(ctx, queueName, taskJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to queue task: %w", err)
	}

	s.logger.Printf("📤 Queued %s task for venue: %s (ID: %s)", priority, venue.Name, task.TaskID)
	return nil
}

// performHealthCheck checks the health of Redis and MongoDB connections
func (s *Scheduler) performHealthCheck() {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	// Check Redis
	if _, err := s.redisClient.Ping(ctx).Result(); err != nil {
		s.logger.Printf("🚨 Redis health check failed: %v", err)
	}

	// Check MongoDB
	if err := s.mongoClient.Ping(ctx, nil); err != nil {
		s.logger.Printf("🚨 MongoDB health check failed: %v", err)
	}

	// Check queue lengths
	queueLength, err := s.redisClient.LLen(ctx, s.config.TaskQueueName).Result()
	if err != nil {
		s.logger.Printf("⚠️  Failed to check queue length: %v", err)
	} else {
		s.logger.Printf("📊 Queue length: %d tasks", queueLength)
	}
}

// GetQueueStats returns statistics about the task queues
func (s *Scheduler) GetQueueStats() (map[string]int64, error) {
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	stats := make(map[string]int64)

	// Normal priority queue
	normalLength, err := s.redisClient.LLen(ctx, s.config.TaskQueueName).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get normal queue length: %w", err)
	}
	stats["normal"] = normalLength

	// High priority queue
	highLength, err := s.redisClient.LLen(ctx, s.config.TaskQueueName+":high").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get high priority queue length: %w", err)
	}
	stats["high"] = highLength

	return stats, nil
} 