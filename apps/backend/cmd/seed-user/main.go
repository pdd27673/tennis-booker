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
)

// User represents user preferences for notifications
type User struct {
	ID                 primitive.ObjectID `bson:"_id"`
	Email              string             `bson:"email"`
	Name               string             `bson:"name"`
	PreferredVenues    []string           `bson:"preferredVenues"`
	TimePreferences    TimePreferences    `bson:"timePreferences"`
	MaxPrice           float64            `bson:"maxPrice"`
	NotificationEnabled bool              `bson:"notificationEnabled"`
	CreatedAt          time.Time          `bson:"createdAt"`
	UpdatedAt          time.Time          `bson:"updatedAt"`
}

type TimePreferences struct {
	WeekdaySlots []TimeSlot `bson:"weekdaySlots"`
	WeekendSlots []TimeSlot `bson:"weekendSlots"`
}

type TimeSlot struct {
	Start string `bson:"start"`
	End   string `bson:"end"`
}

func main() {
	log.Println("Starting user seeding process...")

	// Get MongoDB URI from environment or use default
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:YOUR_PASSWORD@localhost:27017"
	}

	dbName := os.Getenv("DB_NAME")
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

	// Clear existing users
	_, err = usersCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Printf("Warning: Failed to clear users collection: %v", err)
	} else {
		log.Println("Cleared existing users")
	}

	// Create user with your preferences
	user := User{
		ID:    primitive.NewObjectID(),
		Email: "demo@example.com",
		Name:  "Tennis Player",
		PreferredVenues: []string{
			"Victoria Park",
			"Stratford Park", 
			"Ropemakers Field",
		},
		TimePreferences: TimePreferences{
			WeekdaySlots: []TimeSlot{
				{Start: "19:00", End: "22:00"},
			},
			WeekendSlots: []TimeSlot{
				{Start: "10:00", End: "20:00"},
			},
		},
		MaxPrice:           1000.0, // No price limit
		NotificationEnabled: true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Insert user
	result, err := usersCollection.InsertOne(ctx, user)
	if err != nil {
		log.Fatalf("Failed to insert user: %v", err)
	}

	log.Printf("✅ Successfully seeded user profile:")
	log.Printf("   ID: %v", result.InsertedID)
	log.Printf("   Email: %s", user.Email)
	log.Printf("   Preferred venues: %v", user.PreferredVenues)
	log.Printf("   Weekday slots: %v", user.TimePreferences.WeekdaySlots)
	log.Printf("   Weekend slots: %v", user.TimePreferences.WeekendSlots)
	log.Printf("   Max price: £%.2f", user.MaxPrice)
	log.Printf("   Notifications enabled: %v", user.NotificationEnabled)
} 