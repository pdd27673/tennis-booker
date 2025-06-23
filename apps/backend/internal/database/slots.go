package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"tennis-booker/internal/models"
)

// SlotsRepository handles operations on the slots collection
type SlotsRepository struct {
	collection *mongo.Collection
}

// NewSlotsRepository creates a new slots repository
func NewSlotsRepository(db *mongo.Database) *SlotsRepository {
	return &SlotsRepository{
		collection: db.Collection("slots"),
	}
}

// GetAvailableSlots retrieves available court slots
func (r *SlotsRepository) GetAvailableSlots(ctx context.Context, limit int64) ([]*models.CourtSlot, error) {
	filter := bson.M{
		"available": true,
		"date": bson.M{
			"$gte": time.Now().Format("2006-01-02"), // Today or later
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "date", Value: 1}, {Key: "start_time", Value: 1}}) // Sort by date and time

	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var slots []*models.CourtSlot
	for cursor.Next(ctx) {
		// Create a temporary struct that matches the database structure
		var dbSlot struct {
			ID         primitive.ObjectID `bson:"_id"`
			VenueID    primitive.ObjectID `bson:"venue_id"`
			VenueName  string             `bson:"venue_name"`
			CourtID    string             `bson:"court_id"`
			CourtName  string             `bson:"court_name"`
			Date       string             `bson:"date"`
			StartTime  string             `bson:"start_time"`
			EndTime    string             `bson:"end_time"`
			Price      float64            `bson:"price"`
			Currency   string             `bson:"currency"`
			Available  bool               `bson:"available"`
			BookingURL string             `bson:"booking_url"`
			ScrapedAt  time.Time          `bson:"scraped_at"`
			Platform   string             `bson:"platform"`
		}

		if err := cursor.Decode(&dbSlot); err != nil {
			continue // Skip invalid slots
		}

		// Convert to CourtSlot model
		slot := &models.CourtSlot{
			ID:          dbSlot.ID.Hex(),
			VenueID:     dbSlot.VenueID,
			VenueName:   dbSlot.VenueName,
			CourtID:     dbSlot.CourtID,
			CourtName:   dbSlot.CourtName,
			Date:        dbSlot.Date,
			StartTime:   dbSlot.StartTime,
			EndTime:     dbSlot.EndTime,
			Price:       dbSlot.Price,
			Currency:    dbSlot.Currency,
			Available:   dbSlot.Available,
			BookingURL:  dbSlot.BookingURL,
			Provider:    dbSlot.Platform,
			LastScraped: dbSlot.ScrapedAt,
		}
		slots = append(slots, slot)
	}

	return slots, cursor.Err()
}

// GetAvailableSlotsByVenue retrieves available court slots for a specific venue
func (r *SlotsRepository) GetAvailableSlotsByVenue(ctx context.Context, venueID primitive.ObjectID, limit int64) ([]*models.CourtSlot, error) {
	filter := bson.M{
		"venue_id":  venueID,
		"available": true,
		"date": bson.M{
			"$gte": time.Now().Format("2006-01-02"), // Today or later
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "date", Value: 1}, {Key: "start_time", Value: 1}}) // Sort by date and time

	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var slots []*models.CourtSlot
	for cursor.Next(ctx) {
		// Create a temporary struct that matches the database structure
		var dbSlot struct {
			ID         primitive.ObjectID `bson:"_id"`
			VenueID    primitive.ObjectID `bson:"venue_id"`
			VenueName  string             `bson:"venue_name"`
			CourtID    string             `bson:"court_id"`
			CourtName  string             `bson:"court_name"`
			Date       string             `bson:"date"`
			StartTime  string             `bson:"start_time"`
			EndTime    string             `bson:"end_time"`
			Price      float64            `bson:"price"`
			Currency   string             `bson:"currency"`
			Available  bool               `bson:"available"`
			BookingURL string             `bson:"booking_url"`
			ScrapedAt  time.Time          `bson:"scraped_at"`
			Platform   string             `bson:"platform"`
		}

		if err := cursor.Decode(&dbSlot); err != nil {
			continue // Skip invalid slots
		}

		// Convert to CourtSlot model
		slot := &models.CourtSlot{
			ID:          dbSlot.ID.Hex(),
			VenueID:     dbSlot.VenueID,
			VenueName:   dbSlot.VenueName,
			CourtID:     dbSlot.CourtID,
			CourtName:   dbSlot.CourtName,
			Date:        dbSlot.Date,
			StartTime:   dbSlot.StartTime,
			EndTime:     dbSlot.EndTime,
			Price:       dbSlot.Price,
			Currency:    dbSlot.Currency,
			Available:   dbSlot.Available,
			BookingURL:  dbSlot.BookingURL,
			Provider:    dbSlot.Platform,
			LastScraped: dbSlot.ScrapedAt,
		}
		slots = append(slots, slot)
	}

	return slots, cursor.Err()
}

// GetAvailableSlotsByDate retrieves available court slots for a specific date
func (r *SlotsRepository) GetAvailableSlotsByDate(ctx context.Context, date string, limit int64) ([]*models.CourtSlot, error) {
	filter := bson.M{
		"available": true,
		"date":      date,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "start_time", Value: 1}}) // Sort by time

	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var slots []*models.CourtSlot
	for cursor.Next(ctx) {
		// Create a temporary struct that matches the database structure
		var dbSlot struct {
			ID         primitive.ObjectID `bson:"_id"`
			VenueID    primitive.ObjectID `bson:"venue_id"`
			VenueName  string             `bson:"venue_name"`
			CourtID    string             `bson:"court_id"`
			CourtName  string             `bson:"court_name"`
			Date       string             `bson:"date"`
			StartTime  string             `bson:"start_time"`
			EndTime    string             `bson:"end_time"`
			Price      float64            `bson:"price"`
			Currency   string             `bson:"currency"`
			Available  bool               `bson:"available"`
			BookingURL string             `bson:"booking_url"`
			ScrapedAt  time.Time          `bson:"scraped_at"`
			Platform   string             `bson:"platform"`
		}

		if err := cursor.Decode(&dbSlot); err != nil {
			continue // Skip invalid slots
		}

		// Convert to CourtSlot model
		slot := &models.CourtSlot{
			ID:          dbSlot.ID.Hex(),
			VenueID:     dbSlot.VenueID,
			VenueName:   dbSlot.VenueName,
			CourtID:     dbSlot.CourtID,
			CourtName:   dbSlot.CourtName,
			Date:        dbSlot.Date,
			StartTime:   dbSlot.StartTime,
			EndTime:     dbSlot.EndTime,
			Price:       dbSlot.Price,
			Currency:    dbSlot.Currency,
			Available:   dbSlot.Available,
			BookingURL:  dbSlot.BookingURL,
			Provider:    dbSlot.Platform,
			LastScraped: dbSlot.ScrapedAt,
		}
		slots = append(slots, slot)
	}

	return slots, cursor.Err()
}

// CountAvailableSlots counts the total number of available slots
func (r *SlotsRepository) CountAvailableSlots(ctx context.Context) (int64, error) {
	filter := bson.M{
		"available": true,
		"date": bson.M{
			"$gte": time.Now().Format("2006-01-02"), // Today or later
		},
	}

	return r.collection.CountDocuments(ctx, filter)
}

// CountSlotsByDate counts slots for a specific date
func (r *SlotsRepository) CountSlotsByDate(ctx context.Context, date string) (int64, error) {
	filter := bson.M{
		"available": true,
		"date":      date,
	}

	return r.collection.CountDocuments(ctx, filter)
}

// CountSlotsByDateRange counts slots within a date range
func (r *SlotsRepository) CountSlotsByDateRange(ctx context.Context, startDate, endDate string) (int64, error) {
	filter := bson.M{
		"available": true,
		"date": bson.M{
			"$gte": startDate,
			"$lt":  endDate,
		},
	}

	return r.collection.CountDocuments(ctx, filter)
}

// GetActivePlatforms returns the list of unique platforms that have available slots
func (r *SlotsRepository) GetActivePlatforms(ctx context.Context) ([]string, error) {
	filter := bson.M{
		"available": true,
		"date": bson.M{
			"$gte": time.Now().Format("2006-01-02"), // Today or later
		},
	}

	platforms, err := r.collection.Distinct(ctx, "platform", filter)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, platform := range platforms {
		if str, ok := platform.(string); ok {
			result = append(result, str)
		}
	}

	return result, nil
}
