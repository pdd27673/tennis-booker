package database

import (
	"context"
	"testing"
	"time"

	"tennis-booking-bot/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestDB(t *testing.T) (*mongo.Client, *mongo.Database, func()) {
	// Connect to MongoDB
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://admin:YOUR_PASSWORD@localhost:27017"))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to ping MongoDB: %v", err)
	}

	// Use a test database
	db := client.Database("tennis_booking_test")

	// Return client, database, and cleanup function
	cleanup := func() {
		// Drop the test database
		err := db.Drop(ctx)
		if err != nil {
			t.Logf("Failed to drop test database: %v", err)
		}

		// Disconnect from MongoDB
		err = client.Disconnect(ctx)
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
	time.Sleep(1 * time.Millisecond) // Ensure updated time is different
	user.Name = "Updated User"
	user.Phone = "0987654321"
	err = repo.Update(ctx, user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Find the user by ID to verify update
	foundUser, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to find user by ID: %v", err)
	}

	// Verify user fields
	if foundUser.Name != "Updated User" {
		t.Errorf("Expected updated name 'Updated User', got %s", foundUser.Name)
	}
	if foundUser.Phone != "0987654321" {
		t.Errorf("Expected updated phone '0987654321', got %s", foundUser.Phone)
	}
	if !foundUser.UpdatedAt.After(originalUpdateTime) {
		t.Errorf("Expected updated time to be after original update time")
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

	// Try to find the user by ID
	_, err = repo.FindByID(ctx, user.ID)
	if err == nil {
		t.Errorf("Expected error when finding deleted user, got nil")
	}
}

func TestUserRepository_List(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create multiple test users
	for i := 0; i < 5; i++ {
		user := &models.User{
			Email: primitive.NewObjectID().Hex() + "@example.com",
			Name:  "Test User " + primitive.NewObjectID().Hex(),
		}
		err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// List all users
	users, err := repo.List(ctx, 0, 0)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	// Verify user count
	if len(users) != 5 {
		t.Errorf("Expected 5 users, got %d", len(users))
	}

	// Test pagination
	users, err = repo.List(ctx, 2, 2)
	if err != nil {
		t.Fatalf("Failed to list users with pagination: %v", err)
	}

	// Verify paginated user count
	if len(users) != 2 {
		t.Errorf("Expected 2 users with pagination, got %d", len(users))
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

	// Create a test user
	user1 := &models.User{
		Email: "test@example.com",
		Name:  "Test User 1",
	}
	err = repo.Create(ctx, user1)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Try to create another user with the same email
	user2 := &models.User{
		Email: "test@example.com",
		Name:  "Test User 2",
	}
	err = repo.Create(ctx, user2)
	if err == nil {
		t.Errorf("Expected error when creating user with duplicate email, got nil")
	}
} 