package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"tennis-booker/internal/models"
)

// MockVenueRepository for testing
type MockVenueRepository struct {
	venues []*models.Venue
	err    error
}

func (m *MockVenueRepository) ListActive(ctx context.Context) ([]*models.Venue, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.venues, nil
}

// MockScrapingLogRepository for testing
type MockScrapingLogRepository struct {
	courtSlots []*models.CourtSlot
	err        error
}

func (m *MockScrapingLogRepository) GetAvailableCourtSlots(ctx context.Context, limit int64) ([]*models.CourtSlot, error) {
	if m.err != nil {
		return nil, m.err
	}

	// Apply limit if specified
	if limit > 0 && int64(len(m.courtSlots)) > limit {
		return m.courtSlots[:limit], nil
	}

	return m.courtSlots, nil
}

func (m *MockScrapingLogRepository) GetAvailableCourtSlotsByVenue(ctx context.Context, venueID primitive.ObjectID, limit int64) ([]*models.CourtSlot, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Filter slots by venue ID
	var filteredSlots []*models.CourtSlot
	for _, slot := range m.courtSlots {
		if slot.VenueID == venueID {
			filteredSlots = append(filteredSlots, slot)
		}
	}
	return filteredSlots, nil
}

func (m *MockScrapingLogRepository) GetAvailableCourtSlotsWithFilters(ctx context.Context, filter models.CourtSlotFilter, limit int64) ([]*models.CourtSlot, error) {
	if m.err != nil {
		return nil, m.err
	}

	// Apply filters to the mock data
	var filteredSlots []*models.CourtSlot
	for _, slot := range m.courtSlots {
		// Apply venue filter
		if filter.VenueID != nil && !filter.VenueID.IsZero() && slot.VenueID != *filter.VenueID {
			continue
		}

		// Apply date filter
		if filter.Date != nil && *filter.Date != "" && slot.Date != *filter.Date {
			continue
		}

		// Apply time range filters
		if (filter.StartTime != nil && *filter.StartTime != "") || (filter.EndTime != nil && *filter.EndTime != "") {
			filterStartTime := ""
			filterEndTime := ""
			if filter.StartTime != nil {
				filterStartTime = *filter.StartTime
			}
			if filter.EndTime != nil {
				filterEndTime = *filter.EndTime
			}
			if !isTimeInRangeMock(slot.StartTime, slot.EndTime, filterStartTime, filterEndTime) {
				continue
			}
		}

		// Apply provider filter
		if filter.Provider != nil && *filter.Provider != "" && slot.Provider != *filter.Provider {
			continue
		}

		// Apply price filters
		if filter.MinPrice != nil && slot.Price < *filter.MinPrice {
			continue
		}
		if filter.MaxPrice != nil && slot.Price > *filter.MaxPrice {
			continue
		}

		filteredSlots = append(filteredSlots, slot)
	}

	// Apply limit
	if limit > 0 && int64(len(filteredSlots)) > limit {
		filteredSlots = filteredSlots[:limit]
	}

	return filteredSlots, nil
}

// isTimeInRangeMock is a simplified version for testing
func isTimeInRangeMock(slotStart, slotEnd, filterStart, filterEnd string) bool {
	// If no time filters specified, include all slots
	if filterStart == "" && filterEnd == "" {
		return true
	}

	// For testing, we'll use simple string comparison
	// In real implementation, this would parse times properly

	// If only filterStart is specified, slot must start at or after filterStart
	if filterStart != "" && filterEnd == "" {
		return slotStart >= filterStart
	}

	// If only filterEnd is specified, slot must end at or before filterEnd
	if filterStart == "" && filterEnd != "" {
		return slotEnd <= filterEnd
	}

	// If both are specified, slot must overlap with the range
	if filterStart != "" && filterEnd != "" {
		// Slot overlaps if it doesn't end before filter starts or start after filter ends
		return !(slotEnd <= filterStart || slotStart >= filterEnd)
	}

	return true
}

func TestCourtHandler_ListVenues(t *testing.T) {
	// Create test venues
	testVenues := []*models.Venue{
		{
			ID:       primitive.NewObjectID(),
			Name:     "Test Tennis Club",
			Provider: "lta",
			URL:      "https://example.com",
			Location: models.Location{
				Address:  "123 Test Street",
				City:     "London",
				PostCode: "SW1A 1AA",
			},
			IsActive: true,
		},
		{
			ID:       primitive.NewObjectID(),
			Name:     "Another Tennis Club",
			Provider: "courtsides",
			URL:      "https://another.com",
			Location: models.Location{
				Address:  "456 Another Street",
				City:     "Manchester",
				PostCode: "M1 1AA",
			},
			IsActive: true,
		},
	}

	tests := []struct {
		name           string
		method         string
		mockVenueRepo  *MockVenueRepository
		expectedStatus int
		expectedCount  int
	}{
		{
			name:   "GET request returns venues successfully",
			method: http.MethodGet,
			mockVenueRepo: &MockVenueRepository{
				venues: testVenues,
				err:    nil,
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:   "GET request returns empty array when no venues",
			method: http.MethodGet,
			mockVenueRepo: &MockVenueRepository{
				venues: []*models.Venue{},
				err:    nil,
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:   "GET request returns 500 on database error",
			method: http.MethodGet,
			mockVenueRepo: &MockVenueRepository{
				venues: nil,
				err:    errors.New("database error"),
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  -1, // Not applicable
		},
		{
			name:           "POST request returns method not allowed",
			method:         http.MethodPost,
			mockVenueRepo:  &MockVenueRepository{venues: testVenues},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedCount:  -1,
		},
		{
			name:           "PUT request returns method not allowed",
			method:         http.MethodPut,
			mockVenueRepo:  &MockVenueRepository{venues: testVenues},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedCount:  -1,
		},
		{
			name:           "DELETE request returns method not allowed",
			method:         http.MethodDelete,
			mockVenueRepo:  &MockVenueRepository{venues: testVenues},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedCount:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler with mock repositories
			mockScrapingRepo := &MockScrapingLogRepository{}
			handler := NewCourtHandler(tt.mockVenueRepo, mockScrapingRepo)

			req := httptest.NewRequest(tt.method, "/api/venues", nil)
			w := httptest.NewRecorder()

			handler.ListVenues(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// For successful responses, check the JSON content
			if tt.expectedStatus == http.StatusOK {
				var venues []*models.Venue
				if err := json.Unmarshal(w.Body.Bytes(), &venues); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if len(venues) != tt.expectedCount {
					t.Errorf("Expected %d venues, got %d", tt.expectedCount, len(venues))
				}

				// Check content type
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}

				// If we have venues, verify the structure
				if len(venues) > 0 {
					venue := venues[0]
					if venue.Name == "" {
						t.Error("Expected venue to have a name")
					}
					if venue.Provider == "" {
						t.Error("Expected venue to have a provider")
					}
				}
			}
		})
	}
}

func TestCourtHandler_ListCourts(t *testing.T) {
	// Create test court slots
	testCourtSlots := []*models.CourtSlot{
		{
			ID:          "test_venue_001_court_1_2024-01-15_18:00",
			VenueID:     primitive.NewObjectID(),
			VenueName:   "Test Tennis Club",
			CourtID:     "court_1",
			CourtName:   "Court 1",
			Date:        "2024-01-15",
			StartTime:   "18:00",
			EndTime:     "19:00",
			Price:       15.00,
			Currency:    "GBP",
			Available:   true,
			BookingURL:  "https://example.com/book",
			Provider:    "lta",
			LastScraped: time.Now(),
		},
		{
			ID:          "test_venue_002_court_2_2024-01-15_19:00",
			VenueID:     primitive.NewObjectID(),
			VenueName:   "Another Tennis Club",
			CourtID:     "court_2",
			CourtName:   "Court 2",
			Date:        "2024-01-15",
			StartTime:   "19:00",
			EndTime:     "20:00",
			Price:       20.00,
			Currency:    "GBP",
			Available:   true,
			BookingURL:  "https://another.com/book",
			Provider:    "courtsides",
			LastScraped: time.Now(),
		},
	}

	tests := []struct {
		name             string
		method           string
		mockScrapingRepo *MockScrapingLogRepository
		expectedStatus   int
		expectedCount    int
	}{
		{
			name:   "GET request returns court slots successfully",
			method: http.MethodGet,
			mockScrapingRepo: &MockScrapingLogRepository{
				courtSlots: testCourtSlots,
				err:        nil,
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:   "GET request returns empty array when no court slots",
			method: http.MethodGet,
			mockScrapingRepo: &MockScrapingLogRepository{
				courtSlots: []*models.CourtSlot{},
				err:        nil,
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:   "GET request returns 500 on database error",
			method: http.MethodGet,
			mockScrapingRepo: &MockScrapingLogRepository{
				courtSlots: nil,
				err:        errors.New("database error"),
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  -1, // Not applicable
		},
		{
			name:             "POST request returns method not allowed",
			method:           http.MethodPost,
			mockScrapingRepo: &MockScrapingLogRepository{courtSlots: testCourtSlots},
			expectedStatus:   http.StatusMethodNotAllowed,
			expectedCount:    -1,
		},
		{
			name:             "PUT request returns method not allowed",
			method:           http.MethodPut,
			mockScrapingRepo: &MockScrapingLogRepository{courtSlots: testCourtSlots},
			expectedStatus:   http.StatusMethodNotAllowed,
			expectedCount:    -1,
		},
		{
			name:             "DELETE request returns method not allowed",
			method:           http.MethodDelete,
			mockScrapingRepo: &MockScrapingLogRepository{courtSlots: testCourtSlots},
			expectedStatus:   http.StatusMethodNotAllowed,
			expectedCount:    -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler with mock repositories
			mockVenueRepo := &MockVenueRepository{}
			handler := NewCourtHandler(mockVenueRepo, tt.mockScrapingRepo)

			req := httptest.NewRequest(tt.method, "/api/courts", nil)
			w := httptest.NewRecorder()

			handler.ListCourts(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// For successful responses, check the JSON content
			if tt.expectedStatus == http.StatusOK {
				var courtSlots []*models.CourtSlot
				if err := json.Unmarshal(w.Body.Bytes(), &courtSlots); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if len(courtSlots) != tt.expectedCount {
					t.Errorf("Expected %d court slots, got %d", tt.expectedCount, len(courtSlots))
				}

				// Check content type
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}

				// If we have court slots, verify the structure
				if len(courtSlots) > 0 {
					slot := courtSlots[0]
					if slot.ID == "" {
						t.Error("Expected court slot to have an ID")
					}
					if slot.VenueName == "" {
						t.Error("Expected court slot to have a venue name")
					}
					if slot.CourtName == "" {
						t.Error("Expected court slot to have a court name")
					}
					if slot.Date == "" {
						t.Error("Expected court slot to have a date")
					}
					if slot.StartTime == "" {
						t.Error("Expected court slot to have a start time")
					}
				}
			}
		})
	}
}

func TestNewCourtHandler(t *testing.T) {
	mockVenueRepo := &MockVenueRepository{
		venues: []*models.Venue{}, // Initialize with empty slice instead of nil
	}
	mockScrapingRepo := &MockScrapingLogRepository{
		courtSlots: []*models.CourtSlot{}, // Initialize with empty slice instead of nil
	}
	handler := NewCourtHandler(mockVenueRepo, mockScrapingRepo)

	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}

	// Test that the handler can call the repository methods
	venues, err := handler.venueRepo.ListActive(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if venues == nil {
		t.Error("Expected venues slice to be initialized, got nil")
	}

	courtSlots, err := handler.scrapingLogRepo.GetAvailableCourtSlots(context.Background(), 10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if courtSlots == nil {
		t.Error("Expected court slots slice to be initialized, got nil")
	}
}

// TestNewCourtHandlerWithDB tests the database integration
func TestNewCourtHandlerWithDB(t *testing.T) {
	// This test demonstrates how to use the handler with a real database repository
	// In a real integration test, you would:
	// 1. Set up a test database connection
	// 2. Create VenueRepository and ScrapingLogRepository with the test database
	// 3. Create the handler using NewCourtHandlerWithDB
	// 4. Test the actual database operations

	// For this unit test, we'll create mocks that implement the same interface
	// as the real repositories to verify the integration works

	// Create mocks that behave like real repositories
	mockVenueRepo := &MockVenueRepository{
		venues: []*models.Venue{
			{
				ID:       primitive.NewObjectID(),
				Name:     "Integration Test Club",
				Provider: "test",
				IsActive: true,
			},
		},
	}

	mockScrapingRepo := &MockScrapingLogRepository{
		courtSlots: []*models.CourtSlot{
			{
				ID:        "integration_test_slot",
				VenueName: "Integration Test Club",
				CourtName: "Court 1",
				Date:      "2024-01-15",
				StartTime: "18:00",
				EndTime:   "19:00",
				Available: true,
			},
		},
	}

	// Test that we can create a handler using the interfaces
	handler := NewCourtHandler(mockVenueRepo, mockScrapingRepo)

	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}

	// Test that the venues endpoint works correctly
	req := httptest.NewRequest(http.MethodGet, "/api/venues", nil)
	w := httptest.NewRecorder()

	handler.ListVenues(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var venues []*models.Venue
	if err := json.Unmarshal(w.Body.Bytes(), &venues); err != nil {
		t.Fatalf("Failed to unmarshal venues response: %v", err)
	}

	if len(venues) != 1 {
		t.Errorf("Expected 1 venue, got %d", len(venues))
	}

	if venues[0].Name != "Integration Test Club" {
		t.Errorf("Expected venue name 'Integration Test Club', got %s", venues[0].Name)
	}

	// Test that the courts endpoint works correctly
	req = httptest.NewRequest(http.MethodGet, "/api/courts", nil)
	w = httptest.NewRecorder()

	handler.ListCourts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var courtSlots []*models.CourtSlot
	if err := json.Unmarshal(w.Body.Bytes(), &courtSlots); err != nil {
		t.Fatalf("Failed to unmarshal court slots response: %v", err)
	}

	if len(courtSlots) != 1 {
		t.Errorf("Expected 1 court slot, got %d", len(courtSlots))
	}

	if courtSlots[0].VenueName != "Integration Test Club" {
		t.Errorf("Expected court slot venue name 'Integration Test Club', got %s", courtSlots[0].VenueName)
	}
}

func TestCourtHandler_ListCourtsWithFilters(t *testing.T) {
	// Create test court slots with different venues, dates, and times
	venue1ID := primitive.NewObjectID()
	venue2ID := primitive.NewObjectID()

	testCourtSlots := []*models.CourtSlot{
		{
			ID:          "venue1_court1_2024-01-15_18:00",
			VenueID:     venue1ID,
			VenueName:   "Tennis Club 1",
			CourtID:     "court_1",
			CourtName:   "Court 1",
			Date:        "2024-01-15",
			StartTime:   "18:00",
			EndTime:     "19:00",
			Price:       15.00,
			Currency:    "GBP",
			Available:   true,
			BookingURL:  "https://example1.com/book",
			Provider:    "lta",
			LastScraped: time.Now(),
		},
		{
			ID:          "venue1_court2_2024-01-15_19:00",
			VenueID:     venue1ID,
			VenueName:   "Tennis Club 1",
			CourtID:     "court_2",
			CourtName:   "Court 2",
			Date:        "2024-01-15",
			StartTime:   "19:00",
			EndTime:     "20:00",
			Price:       15.00,
			Currency:    "GBP",
			Available:   true,
			BookingURL:  "https://example1.com/book",
			Provider:    "lta",
			LastScraped: time.Now(),
		},
		{
			ID:          "venue2_court1_2024-01-15_18:00",
			VenueID:     venue2ID,
			VenueName:   "Tennis Club 2",
			CourtID:     "court_1",
			CourtName:   "Court 1",
			Date:        "2024-01-15",
			StartTime:   "18:00",
			EndTime:     "19:00",
			Price:       20.00,
			Currency:    "GBP",
			Available:   true,
			BookingURL:  "https://example2.com/book",
			Provider:    "courtsides",
			LastScraped: time.Now(),
		},
		{
			ID:          "venue1_court1_2024-01-16_18:00",
			VenueID:     venue1ID,
			VenueName:   "Tennis Club 1",
			CourtID:     "court_1",
			CourtName:   "Court 1",
			Date:        "2024-01-16",
			StartTime:   "18:00",
			EndTime:     "19:00",
			Price:       15.00,
			Currency:    "GBP",
			Available:   true,
			BookingURL:  "https://example1.com/book",
			Provider:    "lta",
			LastScraped: time.Now(),
		},
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedCount  int
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Filter by venue ID - venue 1",
			queryParams:    "venueId=" + venue1ID.Hex(),
			expectedCount:  3, // 3 slots for venue 1
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Filter by venue ID - venue 2",
			queryParams:    "venueId=" + venue2ID.Hex(),
			expectedCount:  1, // 1 slot for venue 2
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Filter by date - 2024-01-15",
			queryParams:    "date=2024-01-15",
			expectedCount:  3, // 3 slots on this date
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Filter by date - 2024-01-16",
			queryParams:    "date=2024-01-16",
			expectedCount:  1, // 1 slot on this date
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Filter by start time",
			queryParams:    "startTime=18:30",
			expectedCount:  1, // Only the 19:00-20:00 slot should match
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Filter by end time",
			queryParams:    "endTime=18:30",
			expectedCount:  0, // No slots end at or before 18:30 (all end at 19:00 or 20:00)
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Combined filters - venue and date",
			queryParams:    "venueId=" + venue1ID.Hex() + "&date=2024-01-15",
			expectedCount:  2, // 2 slots for venue 1 on 2024-01-15
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Combined filters - venue, date, and time",
			queryParams:    "venueId=" + venue1ID.Hex() + "&date=2024-01-15&startTime=18:30",
			expectedCount:  1, // 1 slot for venue 1 on 2024-01-15 after 18:30
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Filter with limit",
			queryParams:    "limit=2",
			expectedCount:  2, // Should limit to 2 results
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No matching results",
			queryParams:    "date=2024-01-17",
			expectedCount:  0, // No slots on this date
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid venue ID format",
			queryParams:    "venueId=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid venueId format",
		},
		{
			name:           "Invalid date format",
			queryParams:    "date=invalid-date",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid date format. Use YYYY-MM-DD",
		},
		{
			name:           "Invalid start time format",
			queryParams:    "startTime=invalid-time",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid startTime format. Use HH:MM",
		},
		{
			name:           "Invalid end time format",
			queryParams:    "endTime=25:00",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid endTime format. Use HH:MM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler with mock repositories
			mockVenueRepo := &MockVenueRepository{}
			mockScrapingRepo := &MockScrapingLogRepository{
				courtSlots: testCourtSlots,
			}
			handler := NewCourtHandler(mockVenueRepo, mockScrapingRepo)

			// Build URL with query parameters
			url := "/api/courts"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			handler.ListCourts(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				// For successful responses, check the JSON content
				var courtSlots []*models.CourtSlot
				if err := json.Unmarshal(w.Body.Bytes(), &courtSlots); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if len(courtSlots) != tt.expectedCount {
					t.Errorf("Expected %d court slots, got %d", tt.expectedCount, len(courtSlots))
				}

				// Check content type
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}
			} else if tt.expectedError != "" {
				// For error responses, check the error message
				responseBody := w.Body.String()
				if !strings.Contains(responseBody, tt.expectedError) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedError, responseBody)
				}
			}
		})
	}
}
