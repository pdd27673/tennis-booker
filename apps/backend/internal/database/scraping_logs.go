package database

import (
	"context"
	"errors"
	"time"

	"github.com/yourusername/tennis-booking-system/apps/backend/internal/models"

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