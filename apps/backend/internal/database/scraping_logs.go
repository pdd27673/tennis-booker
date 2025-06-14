package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"tennis-booker/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ScrapingLogRepository handles database operations for scraping logs
type ScrapingLogRepository struct {
	collection *mongo.Collection
}

// NewScrapingLogRepository creates a new ScrapingLogRepository
func NewScrapingLogRepository(db *mongo.Database) *ScrapingLogRepository {
	return &ScrapingLogRepository{
		collection: db.Collection("scraping_logs"),
	}
}

// Create adds a new scraping log to the database
func (r *ScrapingLogRepository) Create(ctx context.Context, log *models.ScrapingLog) error {
	if log.VenueID.IsZero() {
		return errors.New("venue ID is required")
	}

	// Set timestamps
	now := time.Now()
	log.CreatedAt = now
	if log.ScrapeTimestamp.IsZero() {
		log.ScrapeTimestamp = now
	}

	// Insert the log
	result, err := r.collection.InsertOne(ctx, log)
	if err != nil {
		return err
	}

	// Set the ID from the result
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		log.ID = oid
	}

	return nil
}

// FindByID retrieves a scraping log by ID
func (r *ScrapingLogRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.ScrapingLog, error) {
	var log models.ScrapingLog
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&log)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("scraping log not found")
		}
		return nil, err
	}
	return &log, nil
}

// FindByVenueID retrieves scraping logs for a specific venue with pagination
func (r *ScrapingLogRepository) FindByVenueID(ctx context.Context, venueID primitive.ObjectID, skip, limit int64) ([]*models.ScrapingLog, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "scrape_timestamp", Value: -1}}) // Sort by scrape_timestamp descending

	if skip > 0 {
		opts.SetSkip(skip)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, bson.M{"venue_id": venueID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*models.ScrapingLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// FindByTimeRange retrieves scraping logs within a time range
func (r *ScrapingLogRepository) FindByTimeRange(ctx context.Context, startTime, endTime time.Time, skip, limit int64) ([]*models.ScrapingLog, error) {
	filter := bson.M{
		"scrape_timestamp": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "scrape_timestamp", Value: -1}}) // Sort by scrape_timestamp descending

	if skip > 0 {
		opts.SetSkip(skip)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*models.ScrapingLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// FindSuccessfulByVenue retrieves successful scraping logs for a specific venue
func (r *ScrapingLogRepository) FindSuccessfulByVenue(ctx context.Context, venueID primitive.ObjectID, skip, limit int64) ([]*models.ScrapingLog, error) {
	filter := bson.M{
		"venue_id": venueID,
		"success":  true,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "scrape_timestamp", Value: -1}}) // Sort by scrape_timestamp descending

	if skip > 0 {
		opts.SetSkip(skip)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*models.ScrapingLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// FindErrorsByVenue retrieves failed scraping logs for a specific venue
func (r *ScrapingLogRepository) FindErrorsByVenue(ctx context.Context, venueID primitive.ObjectID, skip, limit int64) ([]*models.ScrapingLog, error) {
	filter := bson.M{
		"venue_id": venueID,
		"success":  false,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "scrape_timestamp", Value: -1}}) // Sort by scrape_timestamp descending

	if skip > 0 {
		opts.SetSkip(skip)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*models.ScrapingLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// FindByRunID retrieves scraping logs for a specific run ID
func (r *ScrapingLogRepository) FindByRunID(ctx context.Context, runID string) ([]*models.ScrapingLog, error) {
	filter := bson.M{"run_id": runID}

	opts := options.Find().
		SetSort(bson.D{{Key: "scrape_timestamp", Value: -1}}) // Sort by scrape_timestamp descending

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*models.ScrapingLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// CountByVenue counts the number of scraping logs for a specific venue
func (r *ScrapingLogRepository) CountByVenue(ctx context.Context, venueID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"venue_id": venueID})
}

// Delete removes a scraping log from the database
func (r *ScrapingLogRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// DeleteOlderThan removes scraping logs older than a specified time
func (r *ScrapingLogRepository) DeleteOlderThan(ctx context.Context, olderThan time.Time) (int64, error) {
	filter := bson.M{
		"scrape_timestamp": bson.M{
			"$lt": olderThan,
		},
	}
	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// List retrieves all scraping logs with optional pagination
func (r *ScrapingLogRepository) List(ctx context.Context, skip, limit int64) ([]*models.ScrapingLog, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "scrape_timestamp", Value: -1}}) // Sort by scrape_timestamp descending

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

	var logs []*models.ScrapingLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// CreateIndexes creates any necessary indexes for the scraping_logs collection
func (r *ScrapingLogRepository) CreateIndexes(ctx context.Context) error {
	// Create an index on the venue_id field
	venueIDIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "venue_id", Value: 1}},
	}

	// Create an index on the scrape_timestamp field
	timestampIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "scrape_timestamp", Value: -1}},
	}

	// Create a compound index on venue_id and scrape_timestamp
	venueTimestampIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "venue_id", Value: 1},
			{Key: "scrape_timestamp", Value: -1},
		},
	}

	// Create an index on the success field
	successIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "success", Value: 1}},
	}

	// Create an index on the run_id field
	runIDIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "run_id", Value: 1}},
	}

	// Create a compound index on provider and scrape_timestamp
	providerTimestampIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "provider", Value: 1},
			{Key: "scrape_timestamp", Value: -1},
		},
	}

	// Create a TTL index on createdAt to automatically delete old logs
	// Expires logs after 30 days
	ttlIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "created_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(30 * 24 * 60 * 60), // 30 days in seconds
	}

	// Create indexes
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		venueIDIndex,
		timestampIndex,
		venueTimestampIndex,
		successIndex,
		runIDIndex,
		providerTimestampIndex,
		ttlIndex,
	})
	return err
}

// GetAvailableCourtSlots retrieves available court slots from recent scraping logs
func (r *ScrapingLogRepository) GetAvailableCourtSlots(ctx context.Context, limit int64) ([]*models.CourtSlot, error) {
	// Query for recent successful scraping logs with available slots
	filter := bson.M{
		"success":     true,
		"slots_found": bson.M{"$gt": 0},
		"scrape_timestamp": bson.M{
			"$gte": time.Now().Add(-24 * time.Hour), // Only last 24 hours
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "scrape_timestamp", Value: -1}}) // Most recent first

	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var courtSlots []*models.CourtSlot

	for cursor.Next(ctx) {
		var log models.ScrapingLog
		if err := cursor.Decode(&log); err != nil {
			continue // Skip invalid logs
		}

		// Convert each slot in the log to a CourtSlot
		for _, slot := range log.SlotsFound {
			if !slot.Available {
				continue // Skip unavailable slots
			}

			// Parse time range from slot.Time (format: "HH:MM-HH:MM")
			startTime, endTime := parseTimeRange(slot.Time)

			courtSlot := &models.CourtSlot{
				VenueID:       log.VenueID,
				VenueName:     log.VenueName,
				CourtID:       slot.CourtID,
				CourtName:     slot.Court,
				Date:          slot.Date,
				StartTime:     startTime,
				EndTime:       endTime,
				Price:         slot.Price,
				Currency:      "GBP", // Default currency, could be made configurable
				Available:     slot.Available,
				BookingURL:    slot.URL,
				Provider:      log.Provider,
				LastScraped:   log.ScrapeTimestamp,
				ScrapingLogID: log.ID,
			}

			// Generate unique ID for the slot
			courtSlot.ID = courtSlot.GenerateSlotID()

			courtSlots = append(courtSlots, courtSlot)
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return courtSlots, nil
}

// GetAvailableCourtSlotsByVenue retrieves available court slots for a specific venue
func (r *ScrapingLogRepository) GetAvailableCourtSlotsByVenue(ctx context.Context, venueID primitive.ObjectID, limit int64) ([]*models.CourtSlot, error) {
	filter := bson.M{
		"venue_id":    venueID,
		"success":     true,
		"slots_found": bson.M{"$gt": 0},
		"scrape_timestamp": bson.M{
			"$gte": time.Now().Add(-24 * time.Hour), // Only last 24 hours
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "scrape_timestamp", Value: -1}}) // Most recent first

	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var courtSlots []*models.CourtSlot

	for cursor.Next(ctx) {
		var log models.ScrapingLog
		if err := cursor.Decode(&log); err != nil {
			continue // Skip invalid logs
		}

		// Convert each slot in the log to a CourtSlot
		for _, slot := range log.SlotsFound {
			if !slot.Available {
				continue // Skip unavailable slots
			}

			// Parse time range from slot.Time (format: "HH:MM-HH:MM")
			startTime, endTime := parseTimeRange(slot.Time)

			courtSlot := &models.CourtSlot{
				VenueID:       log.VenueID,
				VenueName:     log.VenueName,
				CourtID:       slot.CourtID,
				CourtName:     slot.Court,
				Date:          slot.Date,
				StartTime:     startTime,
				EndTime:       endTime,
				Price:         slot.Price,
				Currency:      "GBP", // Default currency, could be made configurable
				Available:     slot.Available,
				BookingURL:    slot.URL,
				Provider:      log.Provider,
				LastScraped:   log.ScrapeTimestamp,
				ScrapingLogID: log.ID,
			}

			// Generate unique ID for the slot
			courtSlot.ID = courtSlot.GenerateSlotID()

			courtSlots = append(courtSlots, courtSlot)
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return courtSlots, nil
}

// parseTimeRange parses a time range string like "18:00-19:00" into start and end times
func parseTimeRange(timeRange string) (startTime, endTime string) {
	parts := strings.Split(timeRange, "-")
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	// Fallback: assume 1-hour slot if format is unexpected
	return timeRange, timeRange
}

// GetAvailableCourtSlotsWithFilters retrieves available court slots with optional filtering
func (r *ScrapingLogRepository) GetAvailableCourtSlotsWithFilters(ctx context.Context, filter models.CourtSlotFilter, limit int64) ([]*models.CourtSlot, error) {
	// Build MongoDB filter
	mongoFilter := bson.M{
		"success":     true,
		"slots_found": bson.M{"$gt": 0},
		"scrape_timestamp": bson.M{
			"$gte": time.Now().Add(-24 * time.Hour), // Only last 24 hours
		},
	}

	// Add venue filter if specified
	if filter.VenueID != nil && !filter.VenueID.IsZero() {
		mongoFilter["venue_id"] = *filter.VenueID
	}

	// Add date filter if specified
	if filter.Date != nil && *filter.Date != "" {
		// Parse date and create date range for the entire day
		if parsedDate, err := time.Parse("2006-01-02", *filter.Date); err == nil {
			startOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, parsedDate.Location())
			endOfDay := startOfDay.Add(24 * time.Hour)
			mongoFilter["scrape_timestamp"] = bson.M{
				"$gte": startOfDay,
				"$lt":  endOfDay,
			}
		}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "scrape_timestamp", Value: -1}}) // Most recent first

	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var courtSlots []*models.CourtSlot

	for cursor.Next(ctx) {
		var log models.ScrapingLog
		if err := cursor.Decode(&log); err != nil {
			continue // Skip invalid logs
		}

		// Convert each slot in the log to a CourtSlot
		for _, slot := range log.SlotsFound {
			if !slot.Available {
				continue // Skip unavailable slots
			}

			// Parse time range from slot.Time (format: "HH:MM-HH:MM")
			startTime, endTime := parseTimeRange(slot.Time)

			// Apply time range filters if specified
			if (filter.StartTime != nil && *filter.StartTime != "") || (filter.EndTime != nil && *filter.EndTime != "") {
				filterStartTime := ""
				filterEndTime := ""
				if filter.StartTime != nil {
					filterStartTime = *filter.StartTime
				}
				if filter.EndTime != nil {
					filterEndTime = *filter.EndTime
				}
				if !isTimeInRange(startTime, endTime, filterStartTime, filterEndTime) {
					continue
				}
			}

			// Apply date filter to slot date if specified (in addition to scrape timestamp filter)
			if filter.Date != nil && *filter.Date != "" && slot.Date != *filter.Date {
				continue
			}

			// Apply provider filter if specified
			if filter.Provider != nil && *filter.Provider != "" && log.Provider != *filter.Provider {
				continue
			}

			courtSlot := &models.CourtSlot{
				VenueID:       log.VenueID,
				VenueName:     log.VenueName,
				CourtID:       slot.CourtID,
				CourtName:     slot.Court,
				Date:          slot.Date,
				StartTime:     startTime,
				EndTime:       endTime,
				Price:         slot.Price,
				Currency:      "GBP", // Default currency, could be made configurable
				Available:     slot.Available,
				BookingURL:    slot.URL,
				Provider:      log.Provider,
				LastScraped:   log.ScrapeTimestamp,
				ScrapingLogID: log.ID,
			}

			// Apply price filters if specified
			if filter.MinPrice != nil && courtSlot.Price < *filter.MinPrice {
				continue
			}
			if filter.MaxPrice != nil && courtSlot.Price > *filter.MaxPrice {
				continue
			}

			// Generate unique ID for the slot
			courtSlot.ID = courtSlot.GenerateSlotID()

			courtSlots = append(courtSlots, courtSlot)
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return courtSlots, nil
}

// isTimeInRange checks if a time slot overlaps with the specified time range
func isTimeInRange(slotStart, slotEnd, filterStart, filterEnd string) bool {
	// If no time filters specified, include all slots
	if filterStart == "" && filterEnd == "" {
		return true
	}

	// Parse times for comparison (assuming HH:MM format)
	parseTime := func(timeStr string) (int, int) {
		if timeStr == "" {
			return -1, -1
		}
		var hour, minute int
		if _, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &minute); err != nil {
			return -1, -1
		}
		return hour, minute
	}

	slotStartHour, slotStartMin := parseTime(slotStart)
	slotEndHour, slotEndMin := parseTime(slotEnd)
	filterStartHour, filterStartMin := parseTime(filterStart)
	filterEndHour, filterEndMin := parseTime(filterEnd)

	// Convert to minutes for easier comparison
	toMinutes := func(hour, minute int) int {
		if hour == -1 || minute == -1 {
			return -1
		}
		return hour*60 + minute
	}

	slotStartMinutes := toMinutes(slotStartHour, slotStartMin)
	slotEndMinutes := toMinutes(slotEndHour, slotEndMin)
	filterStartMinutes := toMinutes(filterStartHour, filterStartMin)
	filterEndMinutes := toMinutes(filterEndHour, filterEndMin)

	// If parsing failed, include the slot
	if slotStartMinutes == -1 || slotEndMinutes == -1 {
		return true
	}

	// Check if slot overlaps with filter range
	if filterStartMinutes != -1 && slotEndMinutes <= filterStartMinutes {
		return false // Slot ends before filter starts
	}
	if filterEndMinutes != -1 && slotStartMinutes >= filterEndMinutes {
		return false // Slot starts after filter ends
	}

	return true // Slot overlaps with filter range
}
