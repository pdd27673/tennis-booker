package database

import (
	"context"
	"os"
	"testing"
	"time"

	"tennis-booker/internal/models"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestDB(t *testing.T) (*mongo.Client, *mongo.Database, func()) {
	// Skip integration tests if MongoDB is not available
	mongoURI := os.Getenv("MONGODB_TEST_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:password@localhost:27017"
	}

	// Check if we should skip MongoDB tests
	if os.Getenv("SKIP_MONGODB_TESTS") == "true" {
		t.Skip("Skipping MongoDB integration tests - SKIP_MONGODB_TESTS=true")
	}

	// Connect to MongoDB with a short timeout to fail fast
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Skipf("Skipping MongoDB integration tests - failed to connect: %v", err)
	}

	// Ping the database with short timeout
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer pingCancel()

	err = client.Ping(pingCtx, nil)
	if err != nil {
		client.Disconnect(context.Background())
		t.Skipf("Skipping MongoDB integration tests - failed to ping: %v", err)
	}

	// Use a test database
	db := client.Database("tennis_booking_test")

	// Return client, database, and cleanup function
	cleanup := func() {
		// Drop the test database
		err := db.Drop(context.Background())
		if err != nil {
			t.Logf("Failed to drop test database: %v", err)
		}

		// Disconnect from MongoDB
		err = client.Disconnect(context.Background())
		if err != nil {
			t.Logf("Failed to disconnect from MongoDB: %v", err)
		}
	}

	return client, db, cleanup
}

func TestUserRepository_Create(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Email:           "test@example.com",
		Name:            "Test User",
		Phone:           "1234567890",
		PreferredCourts: []string{"Court 1", "Court 2"},
		PreferredDays:   []string{"Monday", "Wednesday", "Friday"},
		PreferredTimes: []models.TimeRange{
			{Start: "08:00", End: "10:00"},
			{Start: "18:00", End: "20:00"},
		},
		NotifyBy: []string{"email", "sms"},
	}

	// Create the user
	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Verify user ID is set
	if user.ID.IsZero() {
		t.Errorf("User ID should not be zero")
	}

	// Verify timestamps are set
	if user.CreatedAt.IsZero() || user.UpdatedAt.IsZero() {
		t.Errorf("User timestamps should not be zero")
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Email:           "test@example.com",
		Name:            "Test User",
		Phone:           "1234567890",
		PreferredCourts: []string{"Court 1", "Court 2"},
		PreferredDays:   []string{"Monday", "Wednesday", "Friday"},
		PreferredTimes: []models.TimeRange{
			{Start: "08:00", End: "10:00"},
			{Start: "18:00", End: "20:00"},
		},
		NotifyBy: []string{"email", "sms"},
	}

	// Create the user
	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Find the user by ID
	foundUser, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to find user by ID: %v", err)
	}

	// Verify user fields
	if foundUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, foundUser.Email)
	}
	if foundUser.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, foundUser.Name)
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Email:           "test@example.com",
		Name:            "Test User",
		Phone:           "1234567890",
		PreferredCourts: []string{"Court 1", "Court 2"},
		PreferredDays:   []string{"Monday", "Wednesday", "Friday"},
		PreferredTimes: []models.TimeRange{
			{Start: "08:00", End: "10:00"},
			{Start: "18:00", End: "20:00"},
		},
		NotifyBy: []string{"email", "sms"},
	}

	// Create the user
	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Find the user by email
	foundUser, err := repo.FindByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("Failed to find user by email: %v", err)
	}

	// Verify user fields
	if foundUser.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID.Hex(), foundUser.ID.Hex())
	}
	if foundUser.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, foundUser.Name)
	}
}

func TestUserRepository_Update(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Email:           "test@example.com",
		Name:            "Test User",
		Phone:           "1234567890",
		PreferredCourts: []string{"Court 1", "Court 2"},
		PreferredDays:   []string{"Monday", "Wednesday", "Friday"},
		PreferredTimes: []models.TimeRange{
			{Start: "08:00", End: "10:00"},
			{Start: "18:00", End: "20:00"},
		},
		NotifyBy: []string{"email", "sms"},
	}

	// Create the user
	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Update the user
	originalUpdateTime := user.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Ensure timestamp difference

	user.Name = "Updated Test User"
	user.Phone = "0987654321"
	user.PreferredCourts = []string{"Court 3", "Court 4"}

	err = repo.Update(ctx, user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Verify UpdatedAt timestamp changed
	if !user.UpdatedAt.After(originalUpdateTime) {
		t.Errorf("UpdatedAt should be updated")
	}

	// Find the updated user
	foundUser, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to find updated user: %v", err)
	}

	// Verify updated fields
	if foundUser.Name != "Updated Test User" {
		t.Errorf("Expected name 'Updated Test User', got %s", foundUser.Name)
	}
	if foundUser.Phone != "0987654321" {
		t.Errorf("Expected phone '0987654321', got %s", foundUser.Phone)
	}
	if len(foundUser.PreferredCourts) != 2 || foundUser.PreferredCourts[0] != "Court 3" {
		t.Errorf("Expected preferred courts to be updated")
	}
}

func TestUserRepository_Delete(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Email:           "test@example.com",
		Name:            "Test User",
		Phone:           "1234567890",
		PreferredCourts: []string{"Court 1", "Court 2"},
		PreferredDays:   []string{"Monday", "Wednesday", "Friday"},
		PreferredTimes: []models.TimeRange{
			{Start: "08:00", End: "10:00"},
			{Start: "18:00", End: "20:00"},
		},
		NotifyBy: []string{"email", "sms"},
	}

	// Create the user
	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Delete the user
	err = repo.Delete(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Try to find the deleted user
	_, err = repo.FindByID(ctx, user.ID)
	if err == nil {
		t.Errorf("Expected error when finding deleted user")
	}
	if err.Error() != "user not found" {
		t.Errorf("Expected 'user not found', got %v", err)
	}
}

func TestUserRepository_List(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create test users
	users := []*models.User{
		{
			Email:           "user1@example.com",
			Name:            "User 1",
			Phone:           "1111111111",
			PreferredCourts: []string{"Court 1"},
			PreferredDays:   []string{"Monday"},
			PreferredTimes:  []models.TimeRange{{Start: "08:00", End: "10:00"}},
			NotifyBy:        []string{"email"},
		},
		{
			Email:           "user2@example.com",
			Name:            "User 2",
			Phone:           "2222222222",
			PreferredCourts: []string{"Court 2"},
			PreferredDays:   []string{"Tuesday"},
			PreferredTimes:  []models.TimeRange{{Start: "10:00", End: "12:00"}},
			NotifyBy:        []string{"sms"},
		},
	}

	// Create the users
	for _, user := range users {
		err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// List users
	foundUsers, err := repo.List(ctx, 0, 100)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	// Should find at least the 2 users we created
	if len(foundUsers) < 2 {
		t.Errorf("Expected at least 2 users, got %d", len(foundUsers))
	}
}

func TestUserRepository_CreateIndexes(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create indexes
	err := repo.CreateIndexes(ctx)
	if err != nil {
		t.Fatalf("Failed to create indexes: %v", err)
	}

	// Verify indexes were created by listing them
	collection := db.Collection("users")
	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		t.Fatalf("Failed to list indexes: %v", err)
	}
	defer cursor.Close(ctx)

	var indexes []interface{}
	err = cursor.All(ctx, &indexes)
	if err != nil {
		t.Fatalf("Failed to decode indexes: %v", err)
	}

	// Should have at least the default _id index plus our custom indexes
	if len(indexes) < 2 {
		t.Errorf("Expected at least 2 indexes, got %d", len(indexes))
	}
}
