package database

import (
	"context"
	"testing"
	"time"

	"tennis-booking-bot/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBookingRepository_Create(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBookingRepository(db)
	ctx := context.Background()

	// Create a test booking
	booking := &models.Booking{
		UserID:    primitive.NewObjectID(),
		VenueID:   primitive.NewObjectID(),
		CourtID:   "court1",
		Date:      "2023-06-10",
		StartTime: "10:00",
		EndTime:   "11:00",
		Status:    models.BookingStatusPending,
		VenueName: "Test Tennis Club",
		CourtName: "Court 1",
		UserEmail: "test@example.com",
		UserName:  "Test User",
	}

	// Create the booking
	err := repo.Create(ctx, booking)
	if err != nil {
		t.Fatalf("Failed to create booking: %v", err)
	}

	// Verify booking ID is set
	if booking.ID.IsZero() {
		t.Errorf("Booking ID should not be zero")
	}

	// Verify timestamps are set
	if booking.CreatedAt.IsZero() || booking.UpdatedAt.IsZero() {
		t.Errorf("Booking timestamps should not be zero")
	}

	// Verify status is set
	if booking.Status != models.BookingStatusPending {
		t.Errorf("Expected status %s, got %s", models.BookingStatusPending, booking.Status)
	}
}

func TestBookingRepository_FindByID(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBookingRepository(db)
	ctx := context.Background()

	// Create a test booking
	booking := &models.Booking{
		UserID:    primitive.NewObjectID(),
		VenueID:   primitive.NewObjectID(),
		CourtID:   "court1",
		Date:      "2023-06-10",
		StartTime: "10:00",
		EndTime:   "11:00",
		Status:    models.BookingStatusPending,
		VenueName: "Test Tennis Club",
		CourtName: "Court 1",
		UserEmail: "test@example.com",
		UserName:  "Test User",
	}

	// Create the booking
	err := repo.Create(ctx, booking)
	if err != nil {
		t.Fatalf("Failed to create booking: %v", err)
	}

	// Find the booking by ID
	foundBooking, err := repo.FindByID(ctx, booking.ID)
	if err != nil {
		t.Fatalf("Failed to find booking by ID: %v", err)
	}

	// Verify booking fields
	if foundBooking.UserID != booking.UserID {
		t.Errorf("Expected user ID %s, got %s", booking.UserID.Hex(), foundBooking.UserID.Hex())
	}
	if foundBooking.VenueID != booking.VenueID {
		t.Errorf("Expected venue ID %s, got %s", booking.VenueID.Hex(), foundBooking.VenueID.Hex())
	}
	if foundBooking.CourtID != booking.CourtID {
		t.Errorf("Expected court ID %s, got %s", booking.CourtID, foundBooking.CourtID)
	}
	if foundBooking.Date != booking.Date {
		t.Errorf("Expected date %s, got %s", booking.Date, foundBooking.Date)
	}
	if foundBooking.StartTime != booking.StartTime {
		t.Errorf("Expected start time %s, got %s", booking.StartTime, foundBooking.StartTime)
	}
	if foundBooking.EndTime != booking.EndTime {
		t.Errorf("Expected end time %s, got %s", booking.EndTime, foundBooking.EndTime)
	}
	if foundBooking.Status != booking.Status {
		t.Errorf("Expected status %s, got %s", booking.Status, foundBooking.Status)
	}
}

func TestBookingRepository_FindByUserID(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBookingRepository(db)
	ctx := context.Background()

	// Create a user ID
	userID := primitive.NewObjectID()

	// Create multiple test bookings for the user
	for i := 0; i < 3; i++ {
		booking := &models.Booking{
			UserID:    userID,
			VenueID:   primitive.NewObjectID(),
			CourtID:   "court1",
			Date:      "2023-06-10",
			StartTime: "10:00",
			EndTime:   "11:00",
			Status:    models.BookingStatusPending,
		}
		err := repo.Create(ctx, booking)
		if err != nil {
			t.Fatalf("Failed to create booking: %v", err)
		}
	}

	// Create a booking for another user
	otherBooking := &models.Booking{
		UserID:    primitive.NewObjectID(),
		VenueID:   primitive.NewObjectID(),
		CourtID:   "court1",
		Date:      "2023-06-10",
		StartTime: "10:00",
		EndTime:   "11:00",
		Status:    models.BookingStatusPending,
	}
	err := repo.Create(ctx, otherBooking)
	if err != nil {
		t.Fatalf("Failed to create booking: %v", err)
	}

	// Find bookings by user ID
	bookings, err := repo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to find bookings by user ID: %v", err)
	}

	// Verify booking count
	if len(bookings) != 3 {
		t.Errorf("Expected 3 bookings, got %d", len(bookings))
	}

	// Verify all bookings are for the correct user
	for _, booking := range bookings {
		if booking.UserID != userID {
			t.Errorf("Expected user ID %s, got %s", userID.Hex(), booking.UserID.Hex())
		}
	}
}

func TestBookingRepository_FindByVenueID(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBookingRepository(db)
	ctx := context.Background()

	// Create a venue ID
	venueID := primitive.NewObjectID()

	// Create multiple test bookings for the venue
	for i := 0; i < 3; i++ {
		booking := &models.Booking{
			UserID:    primitive.NewObjectID(),
			VenueID:   venueID,
			CourtID:   "court1",
			Date:      "2023-06-10",
			StartTime: "10:00",
			EndTime:   "11:00",
			Status:    models.BookingStatusPending,
		}
		err := repo.Create(ctx, booking)
		if err != nil {
			t.Fatalf("Failed to create booking: %v", err)
		}
	}

	// Create a booking for another venue
	otherBooking := &models.Booking{
		UserID:    primitive.NewObjectID(),
		VenueID:   primitive.NewObjectID(),
		CourtID:   "court1",
		Date:      "2023-06-10",
		StartTime: "10:00",
		EndTime:   "11:00",
		Status:    models.BookingStatusPending,
	}
	err := repo.Create(ctx, otherBooking)
	if err != nil {
		t.Fatalf("Failed to create booking: %v", err)
	}

	// Find bookings by venue ID
	bookings, err := repo.FindByVenueID(ctx, venueID)
	if err != nil {
		t.Fatalf("Failed to find bookings by venue ID: %v", err)
	}

	// Verify booking count
	if len(bookings) != 3 {
		t.Errorf("Expected 3 bookings, got %d", len(bookings))
	}

	// Verify all bookings are for the correct venue
	for _, booking := range bookings {
		if booking.VenueID != venueID {
			t.Errorf("Expected venue ID %s, got %s", venueID.Hex(), booking.VenueID.Hex())
		}
	}
}

func TestBookingRepository_FindByDateRange(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBookingRepository(db)
	ctx := context.Background()

	// Create bookings with different dates
	dates := []string{"2023-06-01", "2023-06-05", "2023-06-10", "2023-06-15", "2023-06-20"}
	for _, date := range dates {
		booking := &models.Booking{
			UserID:    primitive.NewObjectID(),
			VenueID:   primitive.NewObjectID(),
			CourtID:   "court1",
			Date:      date,
			StartTime: "10:00",
			EndTime:   "11:00",
			Status:    models.BookingStatusPending,
		}
		err := repo.Create(ctx, booking)
		if err != nil {
			t.Fatalf("Failed to create booking: %v", err)
		}
	}

	// Find bookings by date range
	bookings, err := repo.FindByDateRange(ctx, "2023-06-05", "2023-06-15")
	if err != nil {
		t.Fatalf("Failed to find bookings by date range: %v", err)
	}

	// Verify booking count
	if len(bookings) != 3 {
		t.Errorf("Expected 3 bookings, got %d", len(bookings))
	}

	// Verify all bookings are within the date range
	for _, booking := range bookings {
		if booking.Date < "2023-06-05" || booking.Date > "2023-06-15" {
			t.Errorf("Expected date between 2023-06-05 and 2023-06-15, got %s", booking.Date)
		}
	}
}

func TestBookingRepository_FindByStatus(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBookingRepository(db)
	ctx := context.Background()

	// Create bookings with different statuses
	statuses := []models.BookingStatus{
		models.BookingStatusPending,
		models.BookingStatusConfirmed,
		models.BookingStatusPending,
		models.BookingStatusFailed,
		models.BookingStatusPending,
	}
	for _, status := range statuses {
		booking := &models.Booking{
			UserID:    primitive.NewObjectID(),
			VenueID:   primitive.NewObjectID(),
			CourtID:   "court1",
			Date:      "2023-06-10",
			StartTime: "10:00",
			EndTime:   "11:00",
			Status:    status,
		}
		err := repo.Create(ctx, booking)
		if err != nil {
			t.Fatalf("Failed to create booking: %v", err)
		}
	}

	// Find bookings by status
	pendingBookings, err := repo.FindByStatus(ctx, models.BookingStatusPending)
	if err != nil {
		t.Fatalf("Failed to find bookings by status: %v", err)
	}

	// Verify booking count
	if len(pendingBookings) != 3 {
		t.Errorf("Expected 3 pending bookings, got %d", len(pendingBookings))
	}

	// Verify all bookings have the correct status
	for _, booking := range pendingBookings {
		if booking.Status != models.BookingStatusPending {
			t.Errorf("Expected status %s, got %s", models.BookingStatusPending, booking.Status)
		}
	}
}

func TestBookingRepository_Update(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBookingRepository(db)
	ctx := context.Background()

	// Create a test booking
	booking := &models.Booking{
		UserID:    primitive.NewObjectID(),
		VenueID:   primitive.NewObjectID(),
		CourtID:   "court1",
		Date:      "2023-06-10",
		StartTime: "10:00",
		EndTime:   "11:00",
		Status:    models.BookingStatusPending,
		VenueName: "Test Tennis Club",
		CourtName: "Court 1",
		UserEmail: "test@example.com",
		UserName:  "Test User",
	}

	// Create the booking
	err := repo.Create(ctx, booking)
	if err != nil {
		t.Fatalf("Failed to create booking: %v", err)
	}

	// Update the booking
	originalUpdateTime := booking.UpdatedAt
	time.Sleep(1 * time.Millisecond) // Ensure updated time is different
	booking.Status = models.BookingStatusConfirmed
	booking.BookingRef = "REF123456"
	booking.Price = 25.50
	booking.Currency = "GBP"
	err = repo.Update(ctx, booking)
	if err != nil {
		t.Fatalf("Failed to update booking: %v", err)
	}

	// Find the booking by ID to verify update
	foundBooking, err := repo.FindByID(ctx, booking.ID)
	if err != nil {
		t.Fatalf("Failed to find booking by ID: %v", err)
	}

	// Verify booking fields
	if foundBooking.Status != models.BookingStatusConfirmed {
		t.Errorf("Expected status %s, got %s", models.BookingStatusConfirmed, foundBooking.Status)
	}
	if foundBooking.BookingRef != "REF123456" {
		t.Errorf("Expected booking ref REF123456, got %s", foundBooking.BookingRef)
	}
	if foundBooking.Price != 25.50 {
		t.Errorf("Expected price 25.50, got %f", foundBooking.Price)
	}
	if foundBooking.Currency != "GBP" {
		t.Errorf("Expected currency GBP, got %s", foundBooking.Currency)
	}
	if !foundBooking.UpdatedAt.After(originalUpdateTime) {
		t.Errorf("Expected updated time to be after original update time")
	}
}

func TestBookingRepository_UpdateStatus(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBookingRepository(db)
	ctx := context.Background()

	// Create a test booking
	booking := &models.Booking{
		UserID:    primitive.NewObjectID(),
		VenueID:   primitive.NewObjectID(),
		CourtID:   "court1",
		Date:      "2023-06-10",
		StartTime: "10:00",
		EndTime:   "11:00",
		Status:    models.BookingStatusPending,
	}

	// Create the booking
	err := repo.Create(ctx, booking)
	if err != nil {
		t.Fatalf("Failed to create booking: %v", err)
	}

	// Update the status to confirmed
	err = repo.UpdateStatus(ctx, booking.ID, models.BookingStatusConfirmed)
	if err != nil {
		t.Fatalf("Failed to update booking status: %v", err)
	}

	// Find the booking by ID to verify update
	foundBooking, err := repo.FindByID(ctx, booking.ID)
	if err != nil {
		t.Fatalf("Failed to find booking by ID: %v", err)
	}

	// Verify booking status
	if foundBooking.Status != models.BookingStatusConfirmed {
		t.Errorf("Expected status %s, got %s", models.BookingStatusConfirmed, foundBooking.Status)
	}

	// Verify booked_at is set
	if foundBooking.BookedAt.IsZero() {
		t.Errorf("Expected booked_at to be set")
	}

	// Update the status to cancelled
	err = repo.UpdateStatus(ctx, booking.ID, models.BookingStatusCancelled)
	if err != nil {
		t.Fatalf("Failed to update booking status: %v", err)
	}

	// Find the booking by ID to verify update
	foundBooking, err = repo.FindByID(ctx, booking.ID)
	if err != nil {
		t.Fatalf("Failed to find booking by ID: %v", err)
	}

	// Verify booking status
	if foundBooking.Status != models.BookingStatusCancelled {
		t.Errorf("Expected status %s, got %s", models.BookingStatusCancelled, foundBooking.Status)
	}

	// Verify cancelled_at is set
	if foundBooking.CancelledAt.IsZero() {
		t.Errorf("Expected cancelled_at to be set")
	}
}

func TestBookingRepository_AddBookingAttempt(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBookingRepository(db)
	ctx := context.Background()

	// Create a test booking
	booking := &models.Booking{
		UserID:    primitive.NewObjectID(),
		VenueID:   primitive.NewObjectID(),
		CourtID:   "court1",
		Date:      "2023-06-10",
		StartTime: "10:00",
		EndTime:   "11:00",
		Status:    models.BookingStatusPending,
	}

	// Create the booking
	err := repo.Create(ctx, booking)
	if err != nil {
		t.Fatalf("Failed to create booking: %v", err)
	}

	// Add a booking attempt
	attempt := models.BookingAttempt{
		Timestamp: time.Now(),
		Success:   false,
		Error:     "Failed to connect to booking system",
		Duration:  1500,
	}
	err = repo.AddBookingAttempt(ctx, booking.ID, attempt)
	if err != nil {
		t.Fatalf("Failed to add booking attempt: %v", err)
	}

	// Find the booking by ID to verify update
	foundBooking, err := repo.FindByID(ctx, booking.ID)
	if err != nil {
		t.Fatalf("Failed to find booking by ID: %v", err)
	}

	// Verify booking attempts
	if len(foundBooking.BookingAttempts) != 1 {
		t.Errorf("Expected 1 booking attempt, got %d", len(foundBooking.BookingAttempts))
	}
	if foundBooking.BookingAttempts[0].Success != false {
		t.Errorf("Expected success false, got %v", foundBooking.BookingAttempts[0].Success)
	}
	if foundBooking.BookingAttempts[0].Error != "Failed to connect to booking system" {
		t.Errorf("Expected error message 'Failed to connect to booking system', got %s", foundBooking.BookingAttempts[0].Error)
	}
	if foundBooking.BookingAttempts[0].Duration != 1500 {
		t.Errorf("Expected duration 1500, got %d", foundBooking.BookingAttempts[0].Duration)
	}

	// Add another booking attempt
	attempt2 := models.BookingAttempt{
		Timestamp: time.Now(),
		Success:   true,
		Duration:  1200,
	}
	err = repo.AddBookingAttempt(ctx, booking.ID, attempt2)
	if err != nil {
		t.Fatalf("Failed to add booking attempt: %v", err)
	}

	// Find the booking by ID to verify update
	foundBooking, err = repo.FindByID(ctx, booking.ID)
	if err != nil {
		t.Fatalf("Failed to find booking by ID: %v", err)
	}

	// Verify booking attempts
	if len(foundBooking.BookingAttempts) != 2 {
		t.Errorf("Expected 2 booking attempts, got %d", len(foundBooking.BookingAttempts))
	}
	if foundBooking.BookingAttempts[1].Success != true {
		t.Errorf("Expected success true, got %v", foundBooking.BookingAttempts[1].Success)
	}
	if foundBooking.BookingAttempts[1].Duration != 1200 {
		t.Errorf("Expected duration 1200, got %d", foundBooking.BookingAttempts[1].Duration)
	}
}

func TestBookingRepository_Delete(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBookingRepository(db)
	ctx := context.Background()

	// Create a test booking
	booking := &models.Booking{
		UserID:    primitive.NewObjectID(),
		VenueID:   primitive.NewObjectID(),
		CourtID:   "court1",
		Date:      "2023-06-10",
		StartTime: "10:00",
		EndTime:   "11:00",
		Status:    models.BookingStatusPending,
	}

	// Create the booking
	err := repo.Create(ctx, booking)
	if err != nil {
		t.Fatalf("Failed to create booking: %v", err)
	}

	// Delete the booking
	err = repo.Delete(ctx, booking.ID)
	if err != nil {
		t.Fatalf("Failed to delete booking: %v", err)
	}

	// Try to find the booking by ID
	_, err = repo.FindByID(ctx, booking.ID)
	if err == nil {
		t.Errorf("Expected error when finding deleted booking, got nil")
	}
}

func TestBookingRepository_List(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBookingRepository(db)
	ctx := context.Background()

	// Create multiple test bookings
	for i := 0; i < 5; i++ {
		booking := &models.Booking{
			UserID:    primitive.NewObjectID(),
			VenueID:   primitive.NewObjectID(),
			CourtID:   "court1",
			Date:      "2023-06-10",
			StartTime: "10:00",
			EndTime:   "11:00",
			Status:    models.BookingStatusPending,
		}
		err := repo.Create(ctx, booking)
		if err != nil {
			t.Fatalf("Failed to create booking: %v", err)
		}
	}

	// List all bookings
	bookings, err := repo.List(ctx, 0, 0)
	if err != nil {
		t.Fatalf("Failed to list bookings: %v", err)
	}

	// Verify booking count
	if len(bookings) != 5 {
		t.Errorf("Expected 5 bookings, got %d", len(bookings))
	}

	// Test pagination
	bookings, err = repo.List(ctx, 2, 2)
	if err != nil {
		t.Fatalf("Failed to list bookings with pagination: %v", err)
	}

	// Verify paginated booking count
	if len(bookings) != 2 {
		t.Errorf("Expected 2 bookings with pagination, got %d", len(bookings))
	}
}

func TestBookingRepository_CreateIndexes(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBookingRepository(db)
	ctx := context.Background()

	// Create indexes
	err := repo.CreateIndexes(ctx)
	if err != nil {
		t.Fatalf("Failed to create indexes: %v", err)
	}

	// Create a test booking
	booking1 := &models.Booking{
		UserID:    primitive.NewObjectID(),
		VenueID:   primitive.NewObjectID(),
		CourtID:   "court1",
		Date:      "2023-06-10",
		StartTime: "10:00",
		EndTime:   "11:00",
		Status:    models.BookingStatusPending,
	}
	err = repo.Create(ctx, booking1)
	if err != nil {
		t.Fatalf("Failed to create booking: %v", err)
	}

	// Try to create another booking with the same venue, court, date, and start time
	booking2 := &models.Booking{
		UserID:    primitive.NewObjectID(),
		VenueID:   booking1.VenueID,
		CourtID:   booking1.CourtID,
		Date:      booking1.Date,
		StartTime: booking1.StartTime,
		EndTime:   "12:00",
		Status:    models.BookingStatusPending,
	}
	err = repo.Create(ctx, booking2)
	if err == nil {
		t.Errorf("Expected error when creating booking with duplicate venue, court, date, and start time, got nil")
	}
} 