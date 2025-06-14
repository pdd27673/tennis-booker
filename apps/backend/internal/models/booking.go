package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BookingStatus represents the status of a booking
type BookingStatus string

const (
	// BookingStatusPending represents a booking that is pending
	BookingStatusPending BookingStatus = "pending"
	// BookingStatusConfirmed represents a booking that is confirmed
	BookingStatusConfirmed BookingStatus = "confirmed"
	// BookingStatusFailed represents a booking that failed
	BookingStatusFailed BookingStatus = "failed"
	// BookingStatusCancelled represents a booking that was cancelled
	BookingStatusCancelled BookingStatus = "cancelled"
)

// Booking represents a tennis court booking
type Booking struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	VenueID         primitive.ObjectID `bson:"venue_id" json:"venue_id"`
	CourtID         string             `bson:"court_id" json:"court_id"`
	Date            string             `bson:"date" json:"date"`             // Format: "YYYY-MM-DD"
	StartTime       string             `bson:"start_time" json:"start_time"` // Format: "HH:MM"
	EndTime         string             `bson:"end_time" json:"end_time"`     // Format: "HH:MM"
	Status          BookingStatus      `bson:"status" json:"status"`
	BookingRef      string             `bson:"booking_ref,omitempty" json:"booking_ref,omitempty"`
	Price           float64            `bson:"price,omitempty" json:"price,omitempty"`
	Currency        string             `bson:"currency,omitempty" json:"currency,omitempty"`
	PaymentStatus   string             `bson:"payment_status,omitempty" json:"payment_status,omitempty"`
	Notes           string             `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
	BookedAt        time.Time          `bson:"booked_at,omitempty" json:"booked_at,omitempty"`
	CancelledAt     time.Time          `bson:"cancelled_at,omitempty" json:"cancelled_at,omitempty"`
	BookingAttempts []BookingAttempt   `bson:"booking_attempts,omitempty" json:"booking_attempts,omitempty"`
	VenueName       string             `bson:"venue_name" json:"venue_name"`
	CourtName       string             `bson:"court_name" json:"court_name"`
	UserEmail       string             `bson:"user_email" json:"user_email"`
	UserName        string             `bson:"user_name" json:"user_name"`
}

// BookingAttempt represents an attempt to book a court
type BookingAttempt struct {
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	Success   bool      `bson:"success" json:"success"`
	Error     string    `bson:"error,omitempty" json:"error,omitempty"`
	Duration  int       `bson:"duration" json:"duration"` // Duration in milliseconds
}

// BookingService provides methods for interacting with bookings
type BookingService struct {
	// Will be implemented later with MongoDB connection
}

// Collection returns the name of the MongoDB collection for bookings
func (BookingService) Collection() string {
	return "bookings"
}
