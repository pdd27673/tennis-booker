package models

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupRefreshTokenTest(t *testing.T) (*mongo.Database, *MongoRefreshTokenService, func()) {
	// Skip integration tests if MongoDB is not available
	if os.Getenv("SKIP_MONGODB_TESTS") == "true" {
		t.Skip("Skipping MongoDB integration tests - SKIP_MONGODB_TESTS=true")
	}

	mongoURI := os.Getenv("MONGODB_TEST_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:password@localhost:27017"
	}

	// Use a unique database name for this test
	dbName := "tennis_booker_refresh_token_test_" + primitive.NewObjectID().Hex()

	// Use a short timeout to fail fast if MongoDB is not available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to connect first to see if MongoDB is available
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Skipf("Skipping MongoDB integration tests - MongoDB not available: %v", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		t.Skipf("Skipping MongoDB integration tests - MongoDB ping failed: %v", err)
	}

	db := client.Database(dbName)
	service := NewMongoRefreshTokenService(db)

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		db.Drop(ctx)
		client.Disconnect(ctx)
	}

	return db, service, cleanup
}

func TestMongoRefreshTokenService_CreateRefreshToken(t *testing.T) {
	_, service, cleanup := setupRefreshTokenTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := primitive.NewObjectID()
	token := "test-refresh-token-123"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Test creating a refresh token
	refreshToken, err := service.CreateRefreshToken(ctx, userID, token, expiresAt)
	require.NoError(t, err)
	assert.NotNil(t, refreshToken)
	assert.Equal(t, userID, refreshToken.UserID)
	assert.NotEmpty(t, refreshToken.TokenHash)
	assert.NotEqual(t, token, refreshToken.TokenHash) // Should be hashed
	assert.Equal(t, expiresAt.Unix(), refreshToken.ExpiresAt.Unix())
	assert.False(t, refreshToken.Revoked)
	assert.Nil(t, refreshToken.RevokedAt)
	assert.WithinDuration(t, time.Now(), refreshToken.CreatedAt, time.Second)
}

func TestMongoRefreshTokenService_ValidateRefreshToken(t *testing.T) {
	_, service, cleanup := setupRefreshTokenTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := primitive.NewObjectID()
	token := "test-refresh-token-456"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Create a refresh token first
	_, err := service.CreateRefreshToken(ctx, userID, token, expiresAt)
	require.NoError(t, err)

	// Test validating the token
	validatedToken, err := service.ValidateRefreshToken(ctx, token)
	require.NoError(t, err)
	assert.NotNil(t, validatedToken)
	assert.Equal(t, userID, validatedToken.UserID)
	assert.False(t, validatedToken.Revoked)

	// Test validating an invalid token
	_, err = service.ValidateRefreshToken(ctx, "invalid-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired refresh token")
}

func TestMongoRefreshTokenService_ValidateExpiredToken(t *testing.T) {
	_, service, cleanup := setupRefreshTokenTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := primitive.NewObjectID()
	token := "test-expired-token"
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago

	// Create an expired refresh token
	_, err := service.CreateRefreshToken(ctx, userID, token, expiresAt)
	require.NoError(t, err)

	// Test validating the expired token
	_, err = service.ValidateRefreshToken(ctx, token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired refresh token")
}

func TestMongoRefreshTokenService_RevokeRefreshToken(t *testing.T) {
	_, service, cleanup := setupRefreshTokenTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := primitive.NewObjectID()
	token := "test-revoke-token"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Create a refresh token first
	_, err := service.CreateRefreshToken(ctx, userID, token, expiresAt)
	require.NoError(t, err)

	// Verify token is valid before revocation
	validatedToken, err := service.ValidateRefreshToken(ctx, token)
	require.NoError(t, err)
	assert.False(t, validatedToken.Revoked)

	// Test revoking the token
	err = service.RevokeRefreshToken(ctx, token)
	require.NoError(t, err)

	// Verify token is no longer valid after revocation
	_, err = service.ValidateRefreshToken(ctx, token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired refresh token")

	// Test revoking a non-existent token
	err = service.RevokeRefreshToken(ctx, "non-existent-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh token not found")
}

func TestMongoRefreshTokenService_RevokeAllUserTokens(t *testing.T) {
	_, service, cleanup := setupRefreshTokenTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Create multiple tokens for the user
	token1 := "user-token-1"
	token2 := "user-token-2"
	otherToken := "other-user-token"

	_, err := service.CreateRefreshToken(ctx, userID, token1, expiresAt)
	require.NoError(t, err)
	_, err = service.CreateRefreshToken(ctx, userID, token2, expiresAt)
	require.NoError(t, err)
	_, err = service.CreateRefreshToken(ctx, otherUserID, otherToken, expiresAt)
	require.NoError(t, err)

	// Verify all tokens are valid before revocation
	_, err = service.ValidateRefreshToken(ctx, token1)
	require.NoError(t, err)
	_, err = service.ValidateRefreshToken(ctx, token2)
	require.NoError(t, err)
	_, err = service.ValidateRefreshToken(ctx, otherToken)
	require.NoError(t, err)

	// Revoke all tokens for the specific user
	err = service.RevokeAllUserTokens(ctx, userID)
	require.NoError(t, err)

	// Verify user's tokens are revoked
	_, err = service.ValidateRefreshToken(ctx, token1)
	assert.Error(t, err)
	_, err = service.ValidateRefreshToken(ctx, token2)
	assert.Error(t, err)

	// Verify other user's token is still valid
	_, err = service.ValidateRefreshToken(ctx, otherToken)
	require.NoError(t, err)
}

func TestMongoRefreshTokenService_CleanupExpiredTokens(t *testing.T) {
	_, service, cleanup := setupRefreshTokenTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := primitive.NewObjectID()

	// Create an expired token
	expiredToken := "expired-token"
	expiredTime := time.Now().Add(-2 * time.Hour)
	_, err := service.CreateRefreshToken(ctx, userID, expiredToken, expiredTime)
	require.NoError(t, err)

	// Create a valid token
	validToken := "valid-token"
	validTime := time.Now().Add(7 * 24 * time.Hour)
	_, err = service.CreateRefreshToken(ctx, userID, validToken, validTime)
	require.NoError(t, err)

	// Create an old revoked token (older than 30 days)
	oldRevokedToken := "old-revoked-token"
	_, err = service.CreateRefreshToken(ctx, userID, oldRevokedToken, validTime)
	require.NoError(t, err)
	err = service.RevokeRefreshToken(ctx, oldRevokedToken)
	require.NoError(t, err)

	// Manually update the revoked_at time to be older than 30 days
	// This simulates an old revoked token
	filter := map[string]interface{}{"token_hash": service.hashToken(oldRevokedToken)}
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"revoked_at": time.Now().AddDate(0, 0, -31),
		},
	}
	_, err = service.collection.UpdateOne(ctx, filter, update)
	require.NoError(t, err)

	// Run cleanup
	err = service.CleanupExpiredTokens(ctx)
	require.NoError(t, err)

	// Verify expired and old revoked tokens are removed
	_, err = service.ValidateRefreshToken(ctx, expiredToken)
	assert.Error(t, err) // Should be removed

	_, err = service.ValidateRefreshToken(ctx, oldRevokedToken)
	assert.Error(t, err) // Should be removed

	// Verify valid token still exists
	_, err = service.ValidateRefreshToken(ctx, validToken)
	require.NoError(t, err) // Should still be valid
}

func TestMongoRefreshTokenService_TokenHashing(t *testing.T) {
	_, service, cleanup := setupRefreshTokenTest(t)
	defer cleanup()

	// Test that the same token produces the same hash
	token := "test-token-for-hashing"
	hash1 := service.hashToken(token)
	hash2 := service.hashToken(token)
	assert.Equal(t, hash1, hash2)

	// Test that different tokens produce different hashes
	differentToken := "different-token"
	hash3 := service.hashToken(differentToken)
	assert.NotEqual(t, hash1, hash3)

	// Test that the hash is not the original token
	assert.NotEqual(t, token, hash1)
	assert.NotEmpty(t, hash1)
}
