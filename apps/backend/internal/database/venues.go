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

// VenueRepository handles database operations for venues
type VenueRepository struct {
	collection *mongo.Collection
}

// NewVenueRepository creates a new VenueRepository
func NewVenueRepository(db *mongo.Database) *VenueRepository {
	return &VenueRepository{
		collection: db.Collection("venues"),
	}
}

// Create adds a new venue to the database
func (r *VenueRepository) Create(ctx context.Context, venue *models.Venue) error {
	if venue.Name == "" {
		return errors.New("venue name is required")
	}

	// Set timestamps
	now := time.Now()
	venue.CreatedAt = now
	venue.UpdatedAt = now

	// Insert the venue
	result, err := r.collection.InsertOne(ctx, venue)
	if err != nil {
		return err
	}

	// Set the ID from the result
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		venue.ID = oid
	}

	return nil
}

// FindByID retrieves a venue by ID
func (r *VenueRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Venue, error) {
	var venue models.Venue
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&venue)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("venue not found")
		}
		return nil, err
	}
	return &venue, nil
}

// FindByName retrieves a venue by name
func (r *VenueRepository) FindByName(ctx context.Context, name string) (*models.Venue, error) {
	var venue models.Venue
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&venue)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("venue not found")
		}
		return nil, err
	}
	return &venue, nil
}

// FindByProvider retrieves venues by provider
func (r *VenueRepository) FindByProvider(ctx context.Context, provider string) ([]*models.Venue, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"provider": provider})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var venues []*models.Venue
	if err := cursor.All(ctx, &venues); err != nil {
		return nil, err
	}

	return venues, nil
}

// Update updates an existing venue
func (r *VenueRepository) Update(ctx context.Context, venue *models.Venue) error {
	if venue.ID.IsZero() {
		return errors.New("venue ID is required")
	}

	// Update timestamp
	venue.UpdatedAt = time.Now()

	// Update the venue
	filter := bson.M{"_id": venue.ID}
	update := bson.M{"$set": venue}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// UpdateLastScraped updates the last_scraped_at field for a venue
func (r *VenueRepository) UpdateLastScraped(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"last_scraped_at": time.Now()}}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete removes a venue from the database
func (r *VenueRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// List retrieves all venues with optional pagination
func (r *VenueRepository) List(ctx context.Context, skip, limit int64) ([]*models.Venue, error) {
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

	var venues []*models.Venue
	if err := cursor.All(ctx, &venues); err != nil {
		return nil, err
	}

	return venues, nil
}

// ListActive retrieves all active venues
func (r *VenueRepository) ListActive(ctx context.Context) ([]*models.Venue, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"is_active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var venues []*models.Venue
	if err := cursor.All(ctx, &venues); err != nil {
		return nil, err
	}

	return venues, nil
}

// CreateIndexes creates any necessary indexes for the venues collection
func (r *VenueRepository) CreateIndexes(ctx context.Context) error {
	// Create a unique index on the name field
	nameIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	// Create an index on the provider field
	providerIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "provider", Value: 1}},
	}

	// Create an index on the is_active field
	activeIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "is_active", Value: 1}},
	}

	// Create indexes
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		nameIndex,
		providerIndex,
		activeIndex,
	})
	return err
}
