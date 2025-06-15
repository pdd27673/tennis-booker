package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user account for authentication
type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	Username       string             `bson:"username"`
	Email          string             `bson:"email"`
	HashedPassword string             `bson:"hashed_password"`
	Name           string             `bson:"name"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
}

// UserPreferences represents user preferences for notifications
type UserPreferences struct {
	ID                   primitive.ObjectID   `bson:"_id,omitempty"`
	UserID               primitive.ObjectID   `bson:"user_id"`
	Times                []TimeRange          `bson:"times,omitempty"`
	MaxPrice             float64              `bson:"max_price,omitempty"`
	PreferredVenues      []string             `bson:"preferred_venues,omitempty"`
	ExcludedVenues       []string             `bson:"excluded_venues,omitempty"`
	PreferredDays        []string             `bson:"preferred_days,omitempty"`
	NotificationSettings NotificationSettings `bson:"notification_settings,omitempty"`
	CreatedAt            time.Time            `bson:"created_at"`
	UpdatedAt            time.Time            `bson:"updated_at"`
}

type TimeRange struct {
	Start string `bson:"start"`
	End   string `bson:"end"`
}

type NotificationSettings struct {
	Email                bool   `bson:"email"`
	EmailAddress         string `bson:"email_address,omitempty"`
	InstantAlerts        bool   `bson:"instant_alerts"`
	MaxAlertsPerHour     int    `bson:"max_alerts_per_hour,omitempty"`
	MaxAlertsPerDay      int    `bson:"max_alerts_per_day,omitempty"`
	AlertTimeWindowStart string `bson:"alert_time_window_start,omitempty"`
	AlertTimeWindowEnd   string `bson:"alert_time_window_end,omitempty"`
	Unsubscribed         bool   `bson:"unsubscribed,omitempty"`
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func main() {
	log.Println("Starting user seeding process...")

	// Get MongoDB URI from environment or use default
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:YOUR_PASSWORD@localhost:27017"
	}

	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = "tennis_booking"
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	db := client.Database(dbName)
	usersCollection := db.Collection("users")
	preferencesCollection := db.Collection("user_preferences")

	// Clear existing users and preferences
	_, err = usersCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Printf("Warning: Failed to clear users collection: %v", err)
	} else {
		log.Println("Cleared existing users")
	}

	_, err = preferencesCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Printf("Warning: Failed to clear user_preferences collection: %v", err)
	} else {
		log.Println("Cleared existing user preferences")
	}

	// Create demo users
	demoUsers := []struct {
		email    string
		password string
		name     string
		username string
	}{
		{"demo@example.com", "DEMO_PASSWORD", "Demo User", "demo"},
		{"test@example.com", "test123", "Test User", "test"},
		{"demo@example.com", "DEMO_PASSWORD", "Tennis Player", "mvgnum"},
	}

	for _, demoUser := range demoUsers {
		// Hash password
		hashedPassword, err := hashPassword(demoUser.password)
		if err != nil {
			log.Fatalf("Failed to hash password for %s: %v", demoUser.email, err)
		}

		// Create user
		user := User{
			ID:             primitive.NewObjectID(),
			Username:       demoUser.username,
			Email:          demoUser.email,
			HashedPassword: hashedPassword,
			Name:           demoUser.name,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// Insert user
		result, err := usersCollection.InsertOne(ctx, user)
		if err != nil {
			log.Fatalf("Failed to insert user %s: %v", demoUser.email, err)
		}

		userID := result.InsertedID.(primitive.ObjectID)

		// Create user preferences
		preferences := UserPreferences{
			ID:     primitive.NewObjectID(),
			UserID: userID,
			Times: []TimeRange{
				{Start: "18:00", End: "20:00"},
				{Start: "09:00", End: "11:00"},
			},
			MaxPrice: 100.0,
			PreferredVenues: []string{
				"Victoria Park",
				"Stratford Park",
				"Ropemakers Field",
			},
			ExcludedVenues: []string{},
			PreferredDays:  []string{"monday", "tuesday", "wednesday", "thursday", "friday"},
			NotificationSettings: NotificationSettings{
				Email:                true,
				EmailAddress:         demoUser.email,
				InstantAlerts:        true,
				MaxAlertsPerHour:     10,
				MaxAlertsPerDay:      50,
				AlertTimeWindowStart: "07:00",
				AlertTimeWindowEnd:   "22:00",
				Unsubscribed:         false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Insert preferences
		_, err = preferencesCollection.InsertOne(ctx, preferences)
		if err != nil {
			log.Fatalf("Failed to insert preferences for %s: %v", demoUser.email, err)
		}

		log.Printf("✅ Successfully seeded user: %s (password: %s)", demoUser.email, demoUser.password)
	}

	log.Println("\n🎾 Demo credentials for testing:")
	log.Println("   Email: demo@example.com, Password: DEMO_PASSWORD")
	log.Println("   Email: test@example.com, Password: test123")
	log.Println("   Email: demo@example.com, Password: DEMO_PASSWORD")
}
