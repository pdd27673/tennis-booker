package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Slot struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	VenueID    primitive.ObjectID `bson:"venue_id" json:"venueId"`
	VenueName  string             `bson:"venue_name" json:"venueName"`
	CourtID    string             `bson:"court_id" json:"courtId"`
	CourtName  string             `bson:"court_name" json:"courtName"`
	Date       string             `bson:"date" json:"date"`
	StartTime  string             `bson:"start_time" json:"startTime"`
	EndTime    string             `bson:"end_time" json:"endTime"`
	Price      float64            `bson:"price" json:"price"`
	Platform   string             `bson:"platform" json:"platform"`
	BookingURL string             `bson:"booking_url" json:"bookingUrl"`
	ScrapedAt  time.Time          `bson:"scraped_at" json:"scrapedAt"`
}

type User struct {
	ID                  primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Email               string              `bson:"email" json:"email"`
	PreferredVenues     []string            `bson:"preferredVenues" json:"preferredVenues"`
	TimePreferences     TimePreferences     `bson:"timePreferences" json:"timePreferences"`
	MaxPrice            float64             `bson:"maxPrice" json:"maxPrice"`
	NotificationEnabled bool                `bson:"notificationEnabled" json:"notificationEnabled"`
}

type TimePreferences struct {
	WeekdaySlots []TimeSlot `bson:"weekdaySlots" json:"weekdaySlots"`
	WeekendSlots []TimeSlot `bson:"weekendSlots" json:"weekendSlots"`
}

type TimeSlot struct {
	Start string `bson:"start" json:"start"`
	End   string `bson:"end" json:"end"`
}

func main() {
	// Get environment variables
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:YOUR_PASSWORD@localhost:27017/tennis_booking?authSource=admin"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "tennis_booking"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		redisPassword = "password"
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database(dbName)

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	// Get user preferences
	var user User
	err = db.Collection("users").FindOne(context.TODO(), bson.M{"email": "demo@example.com"}).Decode(&user)
	if err != nil {
		log.Fatal("Error finding user:", err)
	}

	fmt.Printf("🔍 Finding slots matching preferences for %s...\n", user.Email)
	fmt.Printf("📍 Preferred venues: %v\n", user.PreferredVenues)
	fmt.Printf("⏰ Weekday slots: %v\n", user.TimePreferences.WeekdaySlots)
	fmt.Printf("🌅 Weekend slots: %v\n", user.TimePreferences.WeekendSlots)
	fmt.Printf("💰 Max price: £%.2f\n", user.MaxPrice)

	// Get all slots
	cursor, err := db.Collection("slots").Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal("Error finding slots:", err)
	}
	defer cursor.Close(context.TODO())

	var allSlots []Slot
	if err = cursor.All(context.TODO(), &allSlots); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n📊 Total slots in database: %d\n", len(allSlots))

	// Find matching slots
	var matchingSlots []Slot
	for _, slot := range allSlots {
		if matchesUserPreferences(slot, user) {
			matchingSlots = append(matchingSlots, slot)
		}
	}

	fmt.Printf("✅ Found %d matching slots:\n", len(matchingSlots))

	// Display and push matching slots to Redis
	for i, slot := range matchingSlots {
		fmt.Printf("  %d. %s %s-%s at %s (%s) - £%.2f\n",
			i+1, slot.Date, slot.StartTime, slot.EndTime,
			slot.VenueName, slot.CourtName, slot.Price)

		// Create notification message that matches the notification service's SlotData struct
		message := map[string]interface{}{
			"venueId":     slot.VenueID.Hex(),
			"venueName":   slot.VenueName,
			"platform":    slot.Platform,
			"courtId":     slot.CourtID,
			"courtName":   slot.CourtName,
			"date":        slot.Date,
			"startTime":   slot.StartTime,
			"endTime":     slot.EndTime,
			"price":       slot.Price,
			"isAvailable": true, // All slots we're sending are available
			"bookingUrl":  slot.BookingURL,
			"scrapedAt":   slot.ScrapedAt,
		}

		messageJSON, _ := json.Marshal(message)

		// Push to Redis queue
		err := rdb.LPush(context.TODO(), "court_slots", string(messageJSON)).Err()
		if err != nil {
			fmt.Printf("    ❌ Failed to push to Redis: %v\n", err)
		} else {
			fmt.Printf("    ✅ Pushed to Redis notification queue\n")
		}
	}

	if len(matchingSlots) > 0 {
		fmt.Printf("\n🔔 Pushed %d matching slots to Redis for immediate notification!\n", len(matchingSlots))
	} else {
		fmt.Printf("\n😔 No slots match your current preferences\n")
		fmt.Printf("💡 Try adjusting your time preferences or price range\n")
	}
}

func matchesUserPreferences(slot Slot, user User) bool {
	// Check venue preference
	venueMatches := false
	for _, prefVenue := range user.PreferredVenues {
		if slot.VenueName == prefVenue {
			venueMatches = true
			break
		}
	}
	if !venueMatches {
		return false
	}

	// Check price
	if slot.Price > user.MaxPrice {
		return false
	}

	// Parse date to determine if it's weekend
	date, err := time.Parse("2006-01-02", slot.Date)
	if err != nil {
		return false
	}

	// Parse start time
	startHour, err := parseTimeToHour(slot.StartTime)
	if err != nil {
		return false
	}

	// Check time preferences based on day of week
	isWeekend := date.Weekday() == time.Saturday || date.Weekday() == time.Sunday

	if isWeekend {
		// Check weekend time slots
		for _, timeSlot := range user.TimePreferences.WeekendSlots {
			startHourSlot, _ := parseTimeToHour(timeSlot.Start)
			endHourSlot, _ := parseTimeToHour(timeSlot.End)
			if startHour >= startHourSlot && startHour <= endHourSlot {
				return true
			}
		}
	} else {
		// Check weekday time slots
		for _, timeSlot := range user.TimePreferences.WeekdaySlots {
			startHourSlot, _ := parseTimeToHour(timeSlot.Start)
			endHourSlot, _ := parseTimeToHour(timeSlot.End)
			if startHour >= startHourSlot && startHour <= endHourSlot {
				return true
			}
		}
	}

	return false
}

func parseTimeToHour(timeStr string) (int, error) {
	// Handle formats like "19:00" or "07:30"
	parts := strings.Split(timeStr, ":")
	if len(parts) < 1 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}
	return strconv.Atoi(parts[0])
} 