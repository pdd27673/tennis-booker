package database

import (
	"context"
	"errors"
	"time"

	"tennis-booker/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BookingRepository handles database operations for bookings
type BookingRepository struct {
	collection *mongo.Collection
}

// NewBookingRepository creates a new BookingRepository
func NewBookingRepository(db *mongo.Database) *BookingRepository {
	return &BookingRepository{
		collection: db.Collection("bookings"),
	}
}

// Create adds a new booking to the database
func (r *BookingRepository) Create(ctx context.Context, booking *models.Booking) error {
	if booking.UserID.IsZero() {
		return errors.New("user ID is required")
	}
	if booking.VenueID.IsZero() {
		return errors.New("venue ID is required")
	}
	if booking.CourtID == "" {
		return errors.New("court ID is required")
	}
	if booking.Date == "" {
		return errors.New("date is required")
	}
	if booking.StartTime == "" {
		return errors.New("start time is required")
	}
	if booking.EndTime == "" {
		return errors.New("end time is required")
	}

	// Set timestamps
	now := time.Now()
	booking.CreatedAt = now
	booking.UpdatedAt = now

	// Set default status if not provided
	if booking.Status == "" {
		booking.Status = models.BookingStatusPending
	}

	// Insert the booking
	result, err := r.collection.InsertOne(ctx, booking)
	if err != nil {
		return err
	}

	// Set the ID from the result
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		booking.ID = oid
	}

	return nil
}

// FindByID retrieves a booking by ID
func (r *BookingRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Booking, error) {
	var booking models.Booking
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&booking)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("booking not found")
		}
		return nil, err
	}
	return &booking, nil
}

// FindByUserID retrieves bookings for a specific user
func (r *BookingRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Booking, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []*models.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}

	return bookings, nil
}

// FindByVenueID retrieves bookings for a specific venue
func (r *BookingRepository) FindByVenueID(ctx context.Context, venueID primitive.ObjectID) ([]*models.Booking, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"venue_id": venueID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []*models.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}

	return bookings, nil
}

// FindByDateRange retrieves bookings within a date range
func (r *BookingRepository) FindByDateRange(ctx context.Context, startDate, endDate string) ([]*models.Booking, error) {
	filter := bson.M{
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []*models.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}

	return bookings, nil
}

// FindByStatus retrieves bookings with a specific status
func (r *BookingRepository) FindByStatus(ctx context.Context, status models.BookingStatus) ([]*models.Booking, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": status})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []*models.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}

	return bookings, nil
}

// Update updates an existing booking
func (r *BookingRepository) Update(ctx context.Context, booking *models.Booking) error {
	if booking.ID.IsZero() {
		return errors.New("booking ID is required")
	}

	// Update timestamp
	booking.UpdatedAt = time.Now()

	// Update the booking
	filter := bson.M{"_id": booking.ID}
	update := bson.M{"$set": booking}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// UpdateStatus updates the status of a booking
func (r *BookingRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.BookingStatus) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	// Add booked_at timestamp if status is confirmed
	if status == models.BookingStatusConfirmed {
		update["$set"].(bson.M)["booked_at"] = time.Now()
	}

	// Add cancelled_at timestamp if status is cancelled
	if status == models.BookingStatusCancelled {
		update["$set"].(bson.M)["cancelled_at"] = time.Now()
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// AddBookingAttempt adds a booking attempt to a booking
func (r *BookingRepository) AddBookingAttempt(ctx context.Context, id primitive.ObjectID, attempt models.BookingAttempt) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$push": bson.M{
			"booking_attempts": attempt,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete removes a booking from the database
func (r *BookingRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// List retrieves all bookings with optional pagination
func (r *BookingRepository) List(ctx context.Context, skip, limit int64) ([]*models.Booking, error) {
	opts := options.Find()
	if skip > 0 {
		opts.SetSkip(skip)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []*models.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}

	return bookings, nil
}

// CreateIndexes creates any necessary indexes for the bookings collection
func (r *BookingRepository) CreateIndexes(ctx context.Context) error {
	// Create an index on the user_id field
	userIDIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	}

	// Create an index on the venue_id field
	venueIDIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "venue_id", Value: 1}},
	}

	// Create an index on the date field
	dateIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "date", Value: 1}},
	}

	// Create an index on the status field
	statusIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}},
	}

	// Create a compound index on venue_id, court_id, date, start_time
	// This helps ensure we don't double-book the same court
	bookingConstraintIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "venue_id", Value: 1},
			{Key: "court_id", Value: 1},
			{Key: "date", Value: 1},
			{Key: "start_time", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}

	// Create indexes
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		userIDIndex,
		venueIDIndex,
		dateIndex,
		statusIndex,
		bookingConstraintIndex,
	})
	return err
} 