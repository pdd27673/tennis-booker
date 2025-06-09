package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email           string             `bson:"email" json:"email"`
	Name            string             `bson:"name" json:"name"`
	Phone           string             `bson:"phone,omitempty" json:"phone,omitempty"`
	PreferredCourts []string           `bson:"preferred_courts,omitempty" json:"preferred_courts,omitempty"`
	PreferredDays   []string           `bson:"preferred_days,omitempty" json:"preferred_days,omitempty"`
	PreferredTimes  []TimeRange        `bson:"preferred_times,omitempty" json:"preferred_times,omitempty"`
	NotifyBy        []string           `bson:"notify_by,omitempty" json:"notify_by,omitempty"` // "email", "sms"
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

// TimeRange represents a time range preference
type TimeRange struct {
	Start string `bson:"start" json:"start"` // Format: "HH:MM" in 24-hour format
	End   string `bson:"end" json:"end"`     // Format: "HH:MM" in 24-hour format
}

// UserService provides methods for interacting with users
type UserService struct {
	// Will be implemented later with MongoDB connection
}

// Collection returns the name of the MongoDB collection for users
func (UserService) Collection() string {
	return "users"
} 