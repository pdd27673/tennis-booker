package retention

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"tennis-booker/internal/models"
)

func TestDoesSlotMatchActivePreferences(t *testing.T) {
	// Create test slot
	venueID := primitive.NewObjectID()
	slot := models.CourtSlot{
		ID:        "test-slot-1",
		VenueID:   venueID,
		VenueName: "Test Tennis Club",
		CourtID:   "court1",
		CourtName: "Court 1",
		Date:      "2025-06-16", // Monday
		SlotDate:  time.Date(2025, 6, 16, 10, 0, 0, 0, time.UTC),
		StartTime: "10:00",
		EndTime:   "11:00",
		Price:     25.0,
		Currency:  "GBP",
	}

	tests := []struct {
		name        string
		preferences []models.UserPreferences
		expected    bool
		expectError bool
	}{
		{
			name:        "no preferences - no match",
			preferences: []models.UserPreferences{},
			expected:    false,
			expectError: false,
		},
		{
			name: "single matching preference - venue match",
			preferences: []models.UserPreferences{
				{
					ID:              primitive.NewObjectID(),
					UserID:          primitive.NewObjectID(),
					PreferredVenues: []string{venueID.Hex()},
					NotificationSettings: models.NotificationSettings{
						Unsubscribed: false,
					},
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "single matching preference - day match",
			preferences: []models.UserPreferences{
				{
					ID:            primitive.NewObjectID(),
					UserID:        primitive.NewObjectID(),
					PreferredDays: []string{"monday"},
					NotificationSettings: models.NotificationSettings{
						Unsubscribed: false,
					},
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "single matching preference - time match",
			preferences: []models.UserPreferences{
				{
					ID:     primitive.NewObjectID(),
					UserID: primitive.NewObjectID(),
					Times: []models.TimeRange{
						{Start: "09:00", End: "12:00"},
					},
					NotificationSettings: models.NotificationSettings{
						Unsubscribed: false,
					},
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "single matching preference - price match",
			preferences: []models.UserPreferences{
				{
					ID:       primitive.NewObjectID(),
					UserID:   primitive.NewObjectID(),
					MaxPrice: 30.0,
					NotificationSettings: models.NotificationSettings{
						Unsubscribed: false,
					},
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "excluded venue - no match",
			preferences: []models.UserPreferences{
				{
					ID:             primitive.NewObjectID(),
					UserID:         primitive.NewObjectID(),
					ExcludedVenues: []string{venueID.Hex()},
					NotificationSettings: models.NotificationSettings{
						Unsubscribed: false,
					},
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name: "wrong day - no match",
			preferences: []models.UserPreferences{
				{
					ID:            primitive.NewObjectID(),
					UserID:        primitive.NewObjectID(),
					PreferredDays: []string{"tuesday", "wednesday"},
					NotificationSettings: models.NotificationSettings{
						Unsubscribed: false,
					},
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name: "time no overlap - no match",
			preferences: []models.UserPreferences{
				{
					ID:     primitive.NewObjectID(),
					UserID: primitive.NewObjectID(),
					Times: []models.TimeRange{
						{Start: "14:00", End: "16:00"},
					},
					NotificationSettings: models.NotificationSettings{
						Unsubscribed: false,
					},
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name: "price too high - no match",
			preferences: []models.UserPreferences{
				{
					ID:       primitive.NewObjectID(),
					UserID:   primitive.NewObjectID(),
					MaxPrice: 20.0,
					NotificationSettings: models.NotificationSettings{
						Unsubscribed: false,
					},
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name: "multiple preferences - first matches",
			preferences: []models.UserPreferences{
				{
					ID:            primitive.NewObjectID(),
					UserID:        primitive.NewObjectID(),
					PreferredDays: []string{"monday"},
					NotificationSettings: models.NotificationSettings{
						Unsubscribed: false,
					},
				},
				{
					ID:            primitive.NewObjectID(),
					UserID:        primitive.NewObjectID(),
					PreferredDays: []string{"tuesday"},
					NotificationSettings: models.NotificationSettings{
						Unsubscribed: false,
					},
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "multiple preferences - second matches",
			preferences: []models.UserPreferences{
				{
					ID:            primitive.NewObjectID(),
					UserID:        primitive.NewObjectID(),
					PreferredDays: []string{"tuesday"},
					NotificationSettings: models.NotificationSettings{
						Unsubscribed: false,
					},
				},
				{
					ID:            primitive.NewObjectID(),
					UserID:        primitive.NewObjectID(),
					PreferredDays: []string{"monday"},
					NotificationSettings: models.NotificationSettings{
						Unsubscribed: false,
					},
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "complex matching - all criteria match",
			preferences: []models.UserPreferences{
				{
					ID:              primitive.NewObjectID(),
					UserID:          primitive.NewObjectID(),
					PreferredVenues: []string{venueID.Hex()},
					PreferredDays:   []string{"monday", "wednesday"},
					Times: []models.TimeRange{
						{Start: "09:00", End: "12:00"},
					},
					MaxPrice: 30.0,
					NotificationSettings: models.NotificationSettings{
						Unsubscribed: false,
					},
				},
			},
			expected:    true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DoesSlotMatchActivePreferences(slot, tt.preferences)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestMatchesVenuePreferences(t *testing.T) {
	venueID := primitive.NewObjectID()
	slot := models.CourtSlot{
		VenueID:   venueID,
		VenueName: "Test Tennis Club",
	}

	tests := []struct {
		name     string
		pref     models.UserPreferences
		expected bool
	}{
		{
			name:     "no venue preferences - matches",
			pref:     models.UserPreferences{},
			expected: true,
		},
		{
			name: "preferred venue by ID - matches",
			pref: models.UserPreferences{
				PreferredVenues: []string{venueID.Hex()},
			},
			expected: true,
		},
		{
			name: "preferred venue by name - matches",
			pref: models.UserPreferences{
				PreferredVenues: []string{"Test Tennis Club"},
			},
			expected: true,
		},
		{
			name: "excluded venue by ID - no match",
			pref: models.UserPreferences{
				ExcludedVenues: []string{venueID.Hex()},
			},
			expected: false,
		},
		{
			name: "excluded venue by name - no match",
			pref: models.UserPreferences{
				ExcludedVenues: []string{"Test Tennis Club"},
			},
			expected: false,
		},
		{
			name: "different preferred venue - no match",
			pref: models.UserPreferences{
				PreferredVenues: []string{"Other Tennis Club"},
			},
			expected: false,
		},
		{
			name: "excluded takes precedence over preferred",
			pref: models.UserPreferences{
				PreferredVenues: []string{venueID.Hex()},
				ExcludedVenues:  []string{venueID.Hex()},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesVenuePreferences(slot, tt.pref)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchesDayPreferences(t *testing.T) {
	// Monday slot
	slot := models.CourtSlot{
		Date:     "2025-06-16", // Monday
		SlotDate: time.Date(2025, 6, 16, 10, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name     string
		pref     models.UserPreferences
		expected bool
	}{
		{
			name:     "no day preferences - matches",
			pref:     models.UserPreferences{},
			expected: true,
		},
		{
			name: "preferred day matches - matches",
			pref: models.UserPreferences{
				PreferredDays: []string{"monday"},
			},
			expected: true,
		},
		{
			name: "preferred day matches (case insensitive) - matches",
			pref: models.UserPreferences{
				PreferredDays: []string{"MONDAY"},
			},
			expected: true,
		},
		{
			name: "multiple preferred days, one matches - matches",
			pref: models.UserPreferences{
				PreferredDays: []string{"tuesday", "monday", "friday"},
			},
			expected: true,
		},
		{
			name: "preferred day doesn't match - no match",
			pref: models.UserPreferences{
				PreferredDays: []string{"tuesday"},
			},
			expected: false,
		},
		{
			name: "multiple preferred days, none match - no match",
			pref: models.UserPreferences{
				PreferredDays: []string{"tuesday", "wednesday", "friday"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesDayPreferences(slot, tt.pref)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchesTimePreferences(t *testing.T) {
	slot := models.CourtSlot{
		StartTime: "10:00",
		EndTime:   "11:00",
	}

	tests := []struct {
		name        string
		pref        models.UserPreferences
		expected    bool
		expectError bool
	}{
		{
			name:        "no time preferences - matches",
			pref:        models.UserPreferences{},
			expected:    true,
			expectError: false,
		},
		{
			name: "exact time match - matches",
			pref: models.UserPreferences{
				Times: []models.TimeRange{
					{Start: "10:00", End: "11:00"},
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "overlapping time - matches",
			pref: models.UserPreferences{
				Times: []models.TimeRange{
					{Start: "09:00", End: "10:30"},
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "slot within preference range - matches",
			pref: models.UserPreferences{
				Times: []models.TimeRange{
					{Start: "09:00", End: "12:00"},
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "preference within slot range - matches",
			pref: models.UserPreferences{
				Times: []models.TimeRange{
					{Start: "10:15", End: "10:45"},
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "no overlap - no match",
			pref: models.UserPreferences{
				Times: []models.TimeRange{
					{Start: "14:00", End: "16:00"},
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name: "adjacent times - no match",
			pref: models.UserPreferences{
				Times: []models.TimeRange{
					{Start: "11:00", End: "12:00"},
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name: "multiple time ranges, one matches - matches",
			pref: models.UserPreferences{
				Times: []models.TimeRange{
					{Start: "07:00", End: "08:00"},
					{Start: "09:30", End: "10:30"},
					{Start: "14:00", End: "16:00"},
				},
			},
			expected:    true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := matchesTimePreferences(slot, tt.pref)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestMatchesPricePreferences(t *testing.T) {
	slot := models.CourtSlot{
		Price: 25.0,
	}

	tests := []struct {
		name     string
		pref     models.UserPreferences
		expected bool
	}{
		{
			name:     "no price preference - matches",
			pref:     models.UserPreferences{MaxPrice: 0},
			expected: true,
		},
		{
			name:     "negative price preference - matches",
			pref:     models.UserPreferences{MaxPrice: -1},
			expected: true,
		},
		{
			name:     "price within budget - matches",
			pref:     models.UserPreferences{MaxPrice: 30.0},
			expected: true,
		},
		{
			name:     "exact price match - matches",
			pref:     models.UserPreferences{MaxPrice: 25.0},
			expected: true,
		},
		{
			name:     "price too high - no match",
			pref:     models.UserPreferences{MaxPrice: 20.0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesPricePreferences(slot, tt.pref)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseTimeString(t *testing.T) {
	tests := []struct {
		name        string
		timeStr     string
		expected    int
		expectError bool
	}{
		{
			name:        "valid time - 10:30",
			timeStr:     "10:30",
			expected:    630, // 10*60 + 30
			expectError: false,
		},
		{
			name:        "invalid format",
			timeStr:     "1030",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTimeString(tt.timeStr)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestTimesOverlap(t *testing.T) {
	tests := []struct {
		name     string
		start1   int
		end1     int
		start2   int
		end2     int
		expected bool
	}{
		{
			name:     "exact overlap",
			start1:   600, // 10:00
			end1:     660, // 11:00
			start2:   600, // 10:00
			end2:     660, // 11:00
			expected: true,
		},
		{
			name:     "partial overlap - start",
			start1:   600, // 10:00
			end1:     660, // 11:00
			start2:   570, // 09:30
			end2:     630, // 10:30
			expected: true,
		},
		{
			name:     "partial overlap - end",
			start1:   600, // 10:00
			end1:     660, // 11:00
			start2:   630, // 10:30
			end2:     690, // 11:30
			expected: true,
		},
		{
			name:     "one contains the other",
			start1:   600, // 10:00
			end1:     660, // 11:00
			start2:   540, // 09:00
			end2:     720, // 12:00
			expected: true,
		},
		{
			name:     "no overlap - before",
			start1:   600, // 10:00
			end1:     660, // 11:00
			start2:   480, // 08:00
			end2:     540, // 09:00
			expected: false,
		},
		{
			name:     "no overlap - after",
			start1:   600, // 10:00
			end1:     660, // 11:00
			start2:   720, // 12:00
			end2:     780, // 13:00
			expected: false,
		},
		{
			name:     "adjacent - no overlap",
			start1:   600, // 10:00
			end1:     660, // 11:00
			start2:   660, // 11:00
			end2:     720, // 12:00
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := timesOverlap(tt.start1, tt.end1, tt.start2, tt.end2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetWeekdayFromSlot(t *testing.T) {
	tests := []struct {
		name     string
		slot     models.CourtSlot
		expected string
	}{
		{
			name: "SlotDate set - Monday",
			slot: models.CourtSlot{
				SlotDate: time.Date(2025, 6, 16, 10, 0, 0, 0, time.UTC), // Monday
			},
			expected: "monday",
		},
		{
			name: "SlotDate set - Friday",
			slot: models.CourtSlot{
				SlotDate: time.Date(2025, 6, 20, 10, 0, 0, 0, time.UTC), // Friday
			},
			expected: "friday",
		},
		{
			name: "Date string fallback - Tuesday",
			slot: models.CourtSlot{
				Date: "2025-06-17", // Tuesday
			},
			expected: "tuesday",
		},
		{
			name: "Date string fallback - Sunday",
			slot: models.CourtSlot{
				Date: "2025-06-15", // Sunday
			},
			expected: "sunday",
		},
		{
			name: "invalid date string",
			slot: models.CourtSlot{
				Date: "invalid-date",
			},
			expected: "",
		},
		{
			name:     "no date information",
			slot:     models.CourtSlot{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getWeekdayFromSlot(tt.slot)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDoesSlotMatchActivePreferencesDetailed(t *testing.T) {
	venueID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	slot := models.CourtSlot{
		VenueID:   venueID,
		VenueName: "Test Tennis Club",
		Date:      "2025-06-16", // Monday
		SlotDate:  time.Date(2025, 6, 16, 10, 0, 0, 0, time.UTC),
		StartTime: "10:00",
		EndTime:   "11:00",
		Price:     25.0,
	}

	preferences := []models.UserPreferences{
		{
			ID:              primitive.NewObjectID(),
			UserID:          userID,
			PreferredVenues: []string{venueID.Hex()},
			PreferredDays:   []string{"monday"},
			MaxPrice:        30.0,
			NotificationSettings: models.NotificationSettings{
				Unsubscribed: false,
			},
		},
	}

	result := DoesSlotMatchActivePreferencesDetailed(slot, preferences)

	assert.True(t, result.Matches)
	assert.Equal(t, userID.Hex(), result.MatchedUserID)
	assert.Contains(t, result.MatchReason, "preferred venue")
	assert.Contains(t, result.MatchReason, "preferred day")
	assert.Contains(t, result.MatchReason, "within budget")
	assert.NoError(t, result.Error)
}

func TestBuildMatchReason(t *testing.T) {
	venueID := primitive.NewObjectID()

	slot := models.CourtSlot{
		VenueID:   venueID,
		VenueName: "Test Tennis Club",
		StartTime: "10:00",
		EndTime:   "11:00",
		Price:     25.0,
	}

	pref := models.UserPreferences{
		PreferredVenues: []string{venueID.Hex()},
		MaxPrice:        30.0,
	}

	reason := buildMatchReason(slot, pref)

	assert.Contains(t, reason, "preferred venue: Test Tennis Club")
	assert.Contains(t, reason, "within budget: £25.00 <= £30.00")
}
