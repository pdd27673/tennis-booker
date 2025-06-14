package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPreferenceService_IsActivePreference(t *testing.T) {
	service := &PreferenceService{}

	tests := []struct {
		name     string
		pref     *UserPreferences
		expected bool
	}{
		{
			name: "active preference with times",
			pref: &UserPreferences{
				Times: []TimeRange{{Start: "09:00", End: "11:00"}},
				NotificationSettings: NotificationSettings{
					Unsubscribed: false,
				},
			},
			expected: true,
		},
		{
			name: "active preference with preferred venues",
			pref: &UserPreferences{
				PreferredVenues: []string{"venue1", "venue2"},
				NotificationSettings: NotificationSettings{
					Unsubscribed: false,
				},
			},
			expected: true,
		},
		{
			name: "active preference with excluded venues",
			pref: &UserPreferences{
				ExcludedVenues: []string{"venue3"},
				NotificationSettings: NotificationSettings{
					Unsubscribed: false,
				},
			},
			expected: true,
		},
		{
			name: "active preference with preferred days",
			pref: &UserPreferences{
				PreferredDays: []string{"monday", "wednesday"},
				NotificationSettings: NotificationSettings{
					Unsubscribed: false,
				},
			},
			expected: true,
		},
		{
			name: "active preference with max price",
			pref: &UserPreferences{
				MaxPrice: 50.0,
				NotificationSettings: NotificationSettings{
					Unsubscribed: false,
				},
			},
			expected: true,
		},
		{
			name: "inactive preference - unsubscribed",
			pref: &UserPreferences{
				Times: []TimeRange{{Start: "09:00", End: "11:00"}},
				NotificationSettings: NotificationSettings{
					Unsubscribed: true,
				},
			},
			expected: false,
		},
		{
			name: "inactive preference - no meaningful preferences",
			pref: &UserPreferences{
				Times:           []TimeRange{},
				PreferredVenues: []string{},
				ExcludedVenues:  []string{},
				PreferredDays:   []string{},
				MaxPrice:        0,
				NotificationSettings: NotificationSettings{
					Unsubscribed: false,
				},
			},
			expected: false,
		},
		{
			name: "inactive preference - empty with zero max price",
			pref: &UserPreferences{
				MaxPrice: 0,
				NotificationSettings: NotificationSettings{
					Unsubscribed: false,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.IsActivePreference(tt.pref)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPreferenceService_Collection(t *testing.T) {
	service := &PreferenceService{}
	assert.Equal(t, "user_preferences", service.Collection())
}

func TestTimeRange_Validation(t *testing.T) {
	// Test that TimeRange struct has the expected fields
	timeRange := TimeRange{
		Start: "09:00",
		End:   "11:00",
	}

	assert.Equal(t, "09:00", timeRange.Start)
	assert.Equal(t, "11:00", timeRange.End)
}

func TestUserPreferences_Structure(t *testing.T) {
	// Test that UserPreferences struct has all expected fields
	userID := primitive.NewObjectID()
	now := time.Now()

	pref := UserPreferences{
		ID:              primitive.NewObjectID(),
		UserID:          userID,
		Times:           []TimeRange{{Start: "09:00", End: "11:00"}},
		MaxPrice:        50.0,
		PreferredVenues: []string{"venue1"},
		ExcludedVenues:  []string{"venue2"},
		PreferredDays:   []string{"monday"},
		NotificationSettings: NotificationSettings{
			Email:                true,
			EmailAddress:         "test@example.com",
			InstantAlerts:        true,
			MaxAlertsPerHour:     10,
			MaxAlertsPerDay:      50,
			AlertTimeWindowStart: "07:00",
			AlertTimeWindowEnd:   "22:00",
			Unsubscribed:         false,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Verify all fields are accessible
	assert.NotEqual(t, primitive.NilObjectID, pref.ID)
	assert.Equal(t, userID, pref.UserID)
	assert.Len(t, pref.Times, 1)
	assert.Equal(t, 50.0, pref.MaxPrice)
	assert.Len(t, pref.PreferredVenues, 1)
	assert.Len(t, pref.ExcludedVenues, 1)
	assert.Len(t, pref.PreferredDays, 1)
	assert.True(t, pref.NotificationSettings.Email)
	assert.Equal(t, "test@example.com", pref.NotificationSettings.EmailAddress)
	assert.False(t, pref.NotificationSettings.Unsubscribed)
	assert.Equal(t, now, pref.CreatedAt)
	assert.Equal(t, now, pref.UpdatedAt)
}

func TestNotificationSettings_Structure(t *testing.T) {
	// Test NotificationSettings struct
	settings := NotificationSettings{
		Email:                true,
		EmailAddress:         "user@example.com",
		InstantAlerts:        true,
		MaxAlertsPerHour:     15,
		MaxAlertsPerDay:      100,
		AlertTimeWindowStart: "08:00",
		AlertTimeWindowEnd:   "20:00",
		Unsubscribed:         false,
	}

	assert.True(t, settings.Email)
	assert.Equal(t, "user@example.com", settings.EmailAddress)
	assert.True(t, settings.InstantAlerts)
	assert.Equal(t, 15, settings.MaxAlertsPerHour)
	assert.Equal(t, 100, settings.MaxAlertsPerDay)
	assert.Equal(t, "08:00", settings.AlertTimeWindowStart)
	assert.Equal(t, "20:00", settings.AlertTimeWindowEnd)
	assert.False(t, settings.Unsubscribed)
}

func TestPreferenceRequest_Structure(t *testing.T) {
	// Test PreferenceRequest struct
	maxPrice := 75.0
	notificationSettings := &NotificationSettings{
		Email:         true,
		InstantAlerts: true,
	}

	req := PreferenceRequest{
		Times:                []TimeRange{{Start: "10:00", End: "12:00"}},
		MaxPrice:             &maxPrice,
		PreferredVenues:      []string{"venue1", "venue2"},
		ExcludedVenues:       []string{"venue3"},
		PreferredDays:        []string{"tuesday", "thursday"},
		NotificationSettings: notificationSettings,
	}

	assert.Len(t, req.Times, 1)
	assert.Equal(t, 75.0, *req.MaxPrice)
	assert.Len(t, req.PreferredVenues, 2)
	assert.Len(t, req.ExcludedVenues, 1)
	assert.Len(t, req.PreferredDays, 2)
	assert.NotNil(t, req.NotificationSettings)
	assert.True(t, req.NotificationSettings.Email)
}

// Integration tests would go here for MongoDB operations
// These would require a test MongoDB instance

/*
func TestPreferenceService_GetActiveUserPreferences(t *testing.T) {
	// Setup test MongoDB connection
	// Create test data with various preference configurations
	// Call GetActiveUserPreferences
	// Verify only active preferences are returned
}

func TestPreferenceService_GetActiveUserPreferencesCount(t *testing.T) {
	// Setup test MongoDB connection
	// Create test data with active and inactive preferences
	// Call GetActiveUserPreferencesCount
	// Verify correct count is returned
}

func TestPreferenceService_CreateIndexes(t *testing.T) {
	// Setup test MongoDB connection
	// Call CreateIndexes
	// Verify all indexes were created correctly
}
*/
