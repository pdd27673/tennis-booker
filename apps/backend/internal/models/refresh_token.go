package models

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// RefreshToken represents a refresh token stored in the database
type RefreshToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	TokenHash string             `bson:"token_hash" json:"-"` // Never expose in JSON
	ExpiresAt time.Time          `bson:"expires_at" json:"expires_at"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	Revoked   bool               `bson:"revoked" json:"revoked"`
	RevokedAt *time.Time         `bson:"revoked_at,omitempty" json:"revoked_at,omitempty"`
}

// RefreshTokenService defines the interface for refresh token operations
type RefreshTokenService interface {
	// CreateRefreshToken creates and stores a new refresh token for a user
	CreateRefreshToken(ctx context.Context, userID primitive.ObjectID, token string, expiresAt time.Time) (*RefreshToken, error)
	
	// ValidateRefreshToken validates a refresh token and returns the associated token record
	ValidateRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	
	// RevokeRefreshToken marks a refresh token as revoked
	RevokeRefreshToken(ctx context.Context, token string) error
	
	// RevokeAllUserTokens revokes all refresh tokens for a specific user
	RevokeAllUserTokens(ctx context.Context, userID primitive.ObjectID) error
	
	// CleanupExpiredTokens removes expired tokens from the database
	CleanupExpiredTokens(ctx context.Context) error
}

// MongoRefreshTokenService implements RefreshTokenService using MongoDB
type MongoRefreshTokenService struct {
	collection *mongo.Collection
}

// NewMongoRefreshTokenService creates a new MongoDB-based refresh token service
func NewMongoRefreshTokenService(db *mongo.Database) *MongoRefreshTokenService {
	return &MongoRefreshTokenService{
		collection: db.Collection("refresh_tokens"),
	}
}

// hashToken creates a SHA-256 hash of the token for secure storage
func (s *MongoRefreshTokenService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CreateRefreshToken creates and stores a new refresh token for a user
func (s *MongoRefreshTokenService) CreateRefreshToken(ctx context.Context, userID primitive.ObjectID, token string, expiresAt time.Time) (*RefreshToken, error) {
	refreshToken := &RefreshToken{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		TokenHash: s.hashToken(token),
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	_, err := s.collection.InsertOne(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return refreshToken, nil
}

// ValidateRefreshToken validates a refresh token and returns the associated token record
func (s *MongoRefreshTokenService) ValidateRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	tokenHash := s.hashToken(token)
	
	var refreshToken RefreshToken
	filter := bson.M{
		"token_hash": tokenHash,
		"revoked":    false,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	err := s.collection.FindOne(ctx, filter).Decode(&refreshToken)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("invalid or expired refresh token")
		}
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	return &refreshToken, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (s *MongoRefreshTokenService) RevokeRefreshToken(ctx context.Context, token string) error {
	tokenHash := s.hashToken(token)
	now := time.Now()
	
	filter := bson.M{"token_hash": tokenHash}
	update := bson.M{
		"$set": bson.M{
			"revoked":    true,
			"revoked_at": now,
		},
	}

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("refresh token not found")
	}

	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a specific user
func (s *MongoRefreshTokenService) RevokeAllUserTokens(ctx context.Context, userID primitive.ObjectID) error {
	now := time.Now()
	
	filter := bson.M{
		"user_id": userID,
		"revoked": false,
	}
	update := bson.M{
		"$set": bson.M{
			"revoked":    true,
			"revoked_at": now,
		},
	}

	_, err := s.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}

	return nil
}

// CleanupExpiredTokens removes expired tokens from the database
func (s *MongoRefreshTokenService) CleanupExpiredTokens(ctx context.Context) error {
	filter := bson.M{
		"$or": []bson.M{
			{"expires_at": bson.M{"$lt": time.Now()}},
			{"revoked": true, "revoked_at": bson.M{"$lt": time.Now().AddDate(0, 0, -30)}}, // Remove revoked tokens older than 30 days
		},
	}

	_, err := s.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	return nil
} 