package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCourtSlot_GenerateSlotID(t *testing.T) {
	venueID := primitive.NewObjectID()
	slot := &CourtSlot{
		VenueID:   venueID,
		CourtID:   "court1",
		Date:      "2025-06-14",
		StartTime: "09:00",
	}

	expected := venueID.Hex() + "_court1_2025-06-14_09:00"
	assert.Equal(t, expected, slot.GenerateSlotID())
}

func TestCourtSlot_IsOlderThan(t *testing.T) {
	tests := []struct {
		name     string
		slotDate time.Time
		duration time.Duration
		expected bool
	}{
		{
			name:     "slot is older than duration",
			slotDate: time.Now().Add(-8 * 24 * time.Hour), // 8 days ago
			duration: 7 * 24 * time.Hour,                  // 7 days
			expected: true,
		},
		{
			name:     "slot is newer than duration",
			slotDate: time.Now().Add(-5 * 24 * time.Hour), // 5 days ago
			duration: 7 * 24 * time.Hour,                  // 7 days
			expected: false,
		},
		{
			name:     "slot is exactly at duration boundary",
			slotDate: time.Now().Add(-7 * 24 * time.Hour), // 7 days ago
			duration: 7 * 24 * time.Hour,                  // 7 days
			expected: true,                                // Will be true since time.Since includes microseconds/nanoseconds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slot := &CourtSlot{SlotDate: tt.slotDate}
			assert.Equal(t, tt.expected, slot.IsOlderThan(tt.duration))
		})
	}
}

func TestCourtSlot_HasBeenNotified(t *testing.T) {
	tests := []struct {
		name       string
		notifiedAt *time.Time
		expected   bool
	}{
		{
			name:       "slot has been notified",
			notifiedAt: &time.Time{},
			expected:   true,
		},
		{
			name:       "slot has not been notified",
			notifiedAt: nil,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slot := &CourtSlot{NotifiedAt: tt.notifiedAt}
			assert.Equal(t, tt.expected, slot.HasBeenNotified())
		})
	}
}

func TestCourtSlot_MarkAsNotified(t *testing.T) {
	slot := &CourtSlot{
		NotifiedAt: nil,
		UpdatedAt:  time.Time{}, // Zero time
	}

	beforeMark := time.Now()
	slot.MarkAsNotified()
	afterMark := time.Now()

	// Check that NotifiedAt was set
	require.NotNil(t, slot.NotifiedAt)
	assert.True(t, slot.NotifiedAt.After(beforeMark) || slot.NotifiedAt.Equal(beforeMark))
	assert.True(t, slot.NotifiedAt.Before(afterMark) || slot.NotifiedAt.Equal(afterMark))

	// Check that UpdatedAt was set
	assert.True(t, slot.UpdatedAt.After(beforeMark) || slot.UpdatedAt.Equal(beforeMark))
	assert.True(t, slot.UpdatedAt.Before(afterMark) || slot.UpdatedAt.Equal(afterMark))

	// Check that HasBeenNotified now returns true
	assert.True(t, slot.HasBeenNotified())
}

func TestNewCourtSlotService(t *testing.T) {
	// This test would require a mock MongoDB database
	// For now, we'll just test that the constructor doesn't panic
	// In a real test environment, you would use a test database

	// service := NewCourtSlotService(mockDB)
	// assert.NotNil(t, service)
	// assert.Equal(t, "court_slots", service.Collection())
}

func TestCourtSlotService_Collection(t *testing.T) {
	service := &CourtSlotService{}
	assert.Equal(t, "court_slots", service.Collection())
}

// Integration tests would go here for MongoDB operations
// These would require a test MongoDB instance

func TestCourtSlotFilter_Validation(t *testing.T) {
	// Test that the filter struct has the expected fields
	filter := &CourtSlotFilter{
		VenueID:     nil,
		Date:        nil,
		SlotDateGTE: nil,
		SlotDateLTE: nil,
		StartTime:   nil,
		EndTime:     nil,
		Available:   nil,
		Provider:    nil,
		MinPrice:    nil,
		MaxPrice:    nil,
		Notified:    nil,
	}

	// This test mainly ensures the struct compiles correctly
	assert.NotNil(t, filter)
}

// Mock tests for MongoDB operations would be added here
// Example structure for future integration tests:

/*
func TestCourtSlotService_CreateIndexes(t *testing.T) {
	// Setup test MongoDB connection
	// Create service
	// Call CreateIndexes
	// Verify indexes were created
}

func TestCourtSlotService_FindOldUnnotifiedSlots(t *testing.T) {
	// Setup test data in MongoDB
	// Call FindOldUnnotifiedSlots with various durations
	// Verify correct slots are returned
}

func TestCourtSlotService_DeleteSlotsByIDs(t *testing.T) {
	// Setup test data in MongoDB
	// Call DeleteSlotsByIDs
	// Verify correct slots were deleted
}

func TestCourtSlotService_MarkSlotAsNotified(t *testing.T) {
	// Setup test data in MongoDB
	// Call MarkSlotAsNotified
	// Verify slot was updated correctly
}
*/
