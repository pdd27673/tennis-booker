package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tennis-booker/internal/auth"
	"tennis-booker/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupTestUserHandler() (*UserHandler, models.UserService) {
	userService := models.NewInMemoryUserService()
	userHandler := NewUserHandler(userService)
	return userHandler, userService
}

func TestUserHandler_UpdatePreferences(t *testing.T) {
	userHandler, userService := setupTestUserHandler()

	// Create a test user
	ctx := context.Background()
	testUser, err := userService.CreateUser(ctx, "prefuser", "pref@example.com", "DEMO_PASSWORD")
	require.NoError(t, err)

	t.Run("successful preferences update", func(t *testing.T) {
		preferences := UserPreferences{
			PreferredCourts: []string{"court1", "court2"},
			PreferredDays:   []string{"monday", "wednesday", "friday"},
			PreferredTimes: []models.TimeRange{
				{Start: "09:00", End: "11:00"},
				{Start: "18:00", End: "20:00"},
			},
			NotifyBy: []string{"email", "sms"},
		}

		body, _ := json.Marshal(preferences)
		req := httptest.NewRequest(http.MethodPut, "/api/users/preferences", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Add user claims to context
		claims := &auth.AppClaims{
			UserID:   testUser.ID.Hex(),
			Username: testUser.Username,
		}
		ctx := auth.SetUserClaimsInContext(req.Context(), claims)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		userHandler.UpdatePreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response UserPreferences
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, preferences.PreferredCourts, response.PreferredCourts)
		assert.Equal(t, preferences.PreferredDays, response.PreferredDays)
		assert.Equal(t, preferences.PreferredTimes, response.PreferredTimes)
		assert.Equal(t, preferences.NotifyBy, response.NotifyBy)

		// Verify user was actually updated
		updatedUser, err := userService.FindByID(ctx, testUser.ID.Hex())
		require.NoError(t, err)
		assert.Equal(t, preferences.PreferredCourts, updatedUser.PreferredCourts)
		assert.Equal(t, preferences.PreferredDays, updatedUser.PreferredDays)
		assert.Equal(t, preferences.PreferredTimes, updatedUser.PreferredTimes)
		assert.Equal(t, preferences.NotifyBy, updatedUser.NotifyBy)
	})

	t.Run("partial preferences update", func(t *testing.T) {
		preferences := UserPreferences{
			PreferredDays: []string{"saturday", "sunday"},
			NotifyBy:      []string{"email"},
		}

		body, _ := json.Marshal(preferences)
		req := httptest.NewRequest(http.MethodPut, "/api/users/preferences", bytes.NewBuffer(body))

		claims := &auth.AppClaims{
			UserID:   testUser.ID.Hex(),
			Username: testUser.Username,
		}
		ctx := auth.SetUserClaimsInContext(req.Context(), claims)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		userHandler.UpdatePreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response UserPreferences
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, preferences.PreferredDays, response.PreferredDays)
		assert.Equal(t, preferences.NotifyBy, response.NotifyBy)
		// Other fields should be empty/nil since we only updated specific fields
		assert.Empty(t, response.PreferredCourts)
		assert.Empty(t, response.PreferredTimes)
	})

	t.Run("missing user claims", func(t *testing.T) {
		preferences := UserPreferences{
			PreferredDays: []string{"monday"},
		}

		body, _ := json.Marshal(preferences)
		req := httptest.NewRequest(http.MethodPut, "/api/users/preferences", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		userHandler.UpdatePreferences(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("user not found", func(t *testing.T) {
		preferences := UserPreferences{
			PreferredDays: []string{"monday"},
		}

		body, _ := json.Marshal(preferences)
		req := httptest.NewRequest(http.MethodPut, "/api/users/preferences", bytes.NewBuffer(body))

		// Add claims for non-existent user
		claims := &auth.AppClaims{
			UserID:   primitive.NewObjectID().Hex(),
			Username: "nonexistent",
		}
		ctx := auth.SetUserClaimsInContext(req.Context(), claims)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		userHandler.UpdatePreferences(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/users/preferences", bytes.NewBufferString("invalid json"))

		claims := &auth.AppClaims{
			UserID:   testUser.ID.Hex(),
			Username: testUser.Username,
		}
		ctx := auth.SetUserClaimsInContext(req.Context(), claims)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		userHandler.UpdatePreferences(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/users/preferences", nil)
		w := httptest.NewRecorder()

		userHandler.UpdatePreferences(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		assert.Equal(t, "PUT", w.Header().Get("Allow"))
	})
}

func TestUserHandler_ValidatePreferences(t *testing.T) {
	userHandler, _ := setupTestUserHandler()

	t.Run("valid preferences", func(t *testing.T) {
		preferences := &UserPreferences{
			PreferredCourts: []string{"court1", "court2"},
			PreferredDays:   []string{"monday", "wednesday", "friday"},
			PreferredTimes: []models.TimeRange{
				{Start: "09:00", End: "11:00"},
				{Start: "18:00", End: "20:00"},
			},
			NotifyBy: []string{"email", "sms"},
		}

		err := userHandler.validatePreferences(preferences)
		assert.NoError(t, err)
	})

	t.Run("invalid day", func(t *testing.T) {
		preferences := &UserPreferences{
			PreferredDays: []string{"monday", "invalidday"},
		}

		err := userHandler.validatePreferences(preferences)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid day")
	})

	t.Run("invalid notification method", func(t *testing.T) {
		preferences := &UserPreferences{
			NotifyBy: []string{"email", "invalidmethod"},
		}

		err := userHandler.validatePreferences(preferences)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid notification method")
	})

	t.Run("invalid time format", func(t *testing.T) {
		preferences := &UserPreferences{
			PreferredTimes: []models.TimeRange{
				{Start: "9:00", End: "11:00"}, // Invalid format (should be 09:00)
			},
		}

		err := userHandler.validatePreferences(preferences)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid time format")
	})

	t.Run("invalid time value", func(t *testing.T) {
		preferences := &UserPreferences{
			PreferredTimes: []models.TimeRange{
				{Start: "25:00", End: "11:00"}, // Invalid hour
			},
		}

		err := userHandler.validatePreferences(preferences)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid time value")
	})

	t.Run("start time after end time", func(t *testing.T) {
		preferences := &UserPreferences{
			PreferredTimes: []models.TimeRange{
				{Start: "18:00", End: "09:00"}, // Start after end
			},
		}

		err := userHandler.validatePreferences(preferences)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start time must be before end time")
	})
}

func TestNewUserHandler(t *testing.T) {
	userService := models.NewInMemoryUserService()
	handler := NewUserHandler(userService)

	assert.NotNil(t, handler)
	assert.Equal(t, userService, handler.userService)
}
