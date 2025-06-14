package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"tennis-booker/internal/auth"
	"tennis-booker/internal/models"
)

// MockJWTSecretsProvider implements the JWTSecretsProvider interface for testing
type MockJWTSecretsProvider struct {
	secret string
}

func (m *MockJWTSecretsProvider) GetJWTSecret() (string, error) {
	return m.secret, nil
}

// IntegrationTestSuite provides a comprehensive test environment
type IntegrationTestSuite struct {
	handler        *CourtHandler
	jwtService     *auth.JWTService
	testVenues     []*models.Venue
	testCourtSlots []*models.CourtSlot
	validToken     string
	expiredToken   string
}

// SetupIntegrationTestSuite creates a test environment with mock data
func SetupIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	// Create mock JWT secrets provider
	mockSecretsProvider := &MockJWTSecretsProvider{secret: "test-secret-key"}

	// Create JWT service for testing
	jwtService := auth.NewJWTService(mockSecretsProvider, "test-issuer")

	// Create test venues
	venue1ID := primitive.NewObjectID()
	venue2ID := primitive.NewObjectID()

	testVenues := []*models.Venue{
		{
			ID:       venue1ID,
			Name:     "Premium Tennis Club",
			Provider: "lta",
			URL:      "https://premium-tennis.com",
			Location: models.Location{
				Address:  "123 Premium Street",
				City:     "London",
				PostCode: "SW1A 1AA",
			},
			Courts: []models.Court{
				{
					ID:   "court_1",
					Name: "Court 1",
				},
				{
					ID:   "court_2",
					Name: "Court 2",
				},
			},
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:       venue2ID,
			Name:     "Community Sports Center",
			Provider: "courtsides",
			URL:      "https://community-sports.com",
			Location: models.Location{
				Address:  "456 Community Road",
				City:     "Manchester",
				PostCode: "M1 2BB",
			},
			Courts: []models.Court{
				{
					ID:   "court_1",
					Name: "Court 1",
				},
			},
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Create test court slots
	testCourtSlots := []*models.CourtSlot{
		{
			ID:            "premium_court1_2024-01-15_18:00",
			VenueID:       venue1ID,
			VenueName:     "Premium Tennis Club",
			CourtID:       "court_1",
			CourtName:     "Court 1",
			Date:          "2024-01-15",
			StartTime:     "18:00",
			EndTime:       "19:00",
			Price:         25.00,
			Currency:      "GBP",
			Available:     true,
			BookingURL:    "https://premium-tennis.com/book/court1",
			Provider:      "lta",
			LastScraped:   time.Now(),
			ScrapingLogID: primitive.NewObjectID(),
		},
		{
			ID:            "premium_court2_2024-01-15_19:00",
			VenueID:       venue1ID,
			VenueName:     "Premium Tennis Club",
			CourtID:       "court_2",
			CourtName:     "Court 2",
			Date:          "2024-01-15",
			StartTime:     "19:00",
			EndTime:       "20:00",
			Price:         25.00,
			Currency:      "GBP",
			Available:     true,
			BookingURL:    "https://premium-tennis.com/book/court2",
			Provider:      "lta",
			LastScraped:   time.Now(),
			ScrapingLogID: primitive.NewObjectID(),
		},
		{
			ID:            "community_court1_2024-01-15_18:00",
			VenueID:       venue2ID,
			VenueName:     "Community Sports Center",
			CourtID:       "court_1",
			CourtName:     "Court 1",
			Date:          "2024-01-15",
			StartTime:     "18:00",
			EndTime:       "19:00",
			Price:         15.00,
			Currency:      "GBP",
			Available:     true,
			BookingURL:    "https://community-sports.com/book/court1",
			Provider:      "courtsides",
			LastScraped:   time.Now(),
			ScrapingLogID: primitive.NewObjectID(),
		},
		{
			ID:            "premium_court1_2024-01-16_18:00",
			VenueID:       venue1ID,
			VenueName:     "Premium Tennis Club",
			CourtID:       "court_1",
			CourtName:     "Court 1",
			Date:          "2024-01-16",
			StartTime:     "18:00",
			EndTime:       "19:00",
			Price:         25.00,
			Currency:      "GBP",
			Available:     true,
			BookingURL:    "https://premium-tennis.com/book/court1",
			Provider:      "lta",
			LastScraped:   time.Now(),
			ScrapingLogID: primitive.NewObjectID(),
		},
	}

	// Create mock repositories with test data
	mockVenueRepo := &MockVenueRepository{venues: testVenues}
	mockScrapingRepo := &MockScrapingLogRepository{courtSlots: testCourtSlots}

	// Create handler
	handler := NewCourtHandler(mockVenueRepo, mockScrapingRepo)

	// Generate test tokens
	testUserID := primitive.NewObjectID().Hex()
	testUsername := "test@example.com"

	validToken, err := jwtService.GenerateToken(testUserID, testUsername, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate valid token: %v", err)
	}

	// Create expired token (using negative duration)
	expiredToken, err := jwtService.GenerateToken(testUserID, testUsername, -time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}

	return &IntegrationTestSuite{
		handler:        handler,
		jwtService:     jwtService,
		testVenues:     testVenues,
		testCourtSlots: testCourtSlots,
		validToken:     validToken,
		expiredToken:   expiredToken,
	}
}

// Helper function to create authenticated request
func (suite *IntegrationTestSuite) createAuthenticatedRequest(method, url, token string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

// Helper function to validate JWT middleware (simulated)
func (suite *IntegrationTestSuite) validateJWTMiddleware(w http.ResponseWriter, r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return false
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return false
	}

	token := authHeader[7:]

	// Validate token using JWT service
	_, err := suite.jwtService.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return false
	}

	return true
}

// TestVenuesEndpointIntegration tests the /api/venues endpoint comprehensively
func TestVenuesEndpointIntegration(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedCount  int
		checkContent   bool
	}{
		{
			name:           "Successful retrieval with valid JWT",
			token:          suite.validToken,
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			checkContent:   true,
		},
		{
			name:           "Unauthorized access - no token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedCount:  0,
			checkContent:   false,
		},
		{
			name:           "Unauthorized access - invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedCount:  0,
			checkContent:   false,
		},
		{
			name:           "Unauthorized access - expired token",
			token:          suite.expiredToken,
			expectedStatus: http.StatusUnauthorized,
			expectedCount:  0,
			checkContent:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := suite.createAuthenticatedRequest(http.MethodGet, "/api/venues", tt.token)
			w := httptest.NewRecorder()

			// Simulate JWT middleware validation
			if !suite.validateJWTMiddleware(w, req) {
				// Middleware rejected the request
				if w.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
				}
				return
			}

			// Call the handler
			suite.handler.ListVenues(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkContent {
				var venues []*models.Venue
				if err := json.Unmarshal(w.Body.Bytes(), &venues); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if len(venues) != tt.expectedCount {
					t.Errorf("Expected %d venues, got %d", tt.expectedCount, len(venues))
				}

				// Verify venue structure and content
				if len(venues) > 0 {
					venue := venues[0]
					if venue.Name == "" {
						t.Error("Venue name should not be empty")
					}
					if venue.Provider == "" {
						t.Error("Venue provider should not be empty")
					}
					if venue.Location.Address == "" {
						t.Error("Venue address should not be empty")
					}
					if venue.URL == "" {
						t.Error("Venue URL should not be empty")
					}
				}

				// Check content type
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}
			}
		})
	}
}

// TestCourtsEndpointIntegration tests the /api/courts endpoint comprehensively
func TestCourtsEndpointIntegration(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	tests := []struct {
		name           string
		token          string
		queryParams    string
		expectedStatus int
		expectedCount  int
		checkContent   bool
		description    string
	}{
		{
			name:           "Successful retrieval - all courts",
			token:          suite.validToken,
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
			checkContent:   true,
			description:    "Should return all available court slots",
		},
		{
			name:           "Filter by venue ID",
			token:          suite.validToken,
			queryParams:    "venueId=" + suite.testVenues[0].ID.Hex(),
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			checkContent:   true,
			description:    "Should return only slots for Premium Tennis Club",
		},
		{
			name:           "Filter by date",
			token:          suite.validToken,
			queryParams:    "date=2024-01-15",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			checkContent:   true,
			description:    "Should return only slots for January 15th",
		},
		{
			name:           "Filter by provider",
			token:          suite.validToken,
			queryParams:    "provider=lta",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			checkContent:   true,
			description:    "Should return only LTA provider slots",
		},
		{
			name:           "Filter by time range",
			token:          suite.validToken,
			queryParams:    "startTime=18:30",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkContent:   true,
			description:    "Should return only slots starting at or after 18:30",
		},
		{
			name:           "Filter by price range",
			token:          suite.validToken,
			queryParams:    "minPrice=20&maxPrice=30",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			checkContent:   true,
			description:    "Should return only premium slots (£20-30)",
		},
		{
			name:           "Combined filters",
			token:          suite.validToken,
			queryParams:    fmt.Sprintf("venueId=%s&date=2024-01-15", suite.testVenues[0].ID.Hex()),
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			checkContent:   true,
			description:    "Should return Premium Tennis Club slots for January 15th",
		},
		{
			name:           "Limit results",
			token:          suite.validToken,
			queryParams:    "limit=2",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			checkContent:   true,
			description:    "Should limit results to 2 slots",
		},
		{
			name:           "No matching results",
			token:          suite.validToken,
			queryParams:    "date=2024-01-17",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			checkContent:   true,
			description:    "Should return empty array for date with no slots",
		},
		{
			name:           "Invalid venue ID format",
			token:          suite.validToken,
			queryParams:    "venueId=invalid-id",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			checkContent:   false,
			description:    "Should return 400 for invalid venue ID format",
		},
		{
			name:           "Invalid date format",
			token:          suite.validToken,
			queryParams:    "date=invalid-date",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			checkContent:   false,
			description:    "Should return 400 for invalid date format",
		},
		{
			name:           "Unauthorized access - no token",
			token:          "",
			queryParams:    "",
			expectedStatus: http.StatusUnauthorized,
			expectedCount:  0,
			checkContent:   false,
			description:    "Should return 401 when no authorization token provided",
		},
		{
			name:           "Unauthorized access - invalid token",
			token:          "invalid-token",
			queryParams:    "",
			expectedStatus: http.StatusUnauthorized,
			expectedCount:  0,
			checkContent:   false,
			description:    "Should return 401 for invalid token",
		},
		{
			name:           "Unauthorized access - expired token",
			token:          suite.expiredToken,
			queryParams:    "",
			expectedStatus: http.StatusUnauthorized,
			expectedCount:  0,
			checkContent:   false,
			description:    "Should return 401 for expired token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/courts"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}

			req := suite.createAuthenticatedRequest(http.MethodGet, url, tt.token)
			w := httptest.NewRecorder()

			// Simulate JWT middleware validation
			if !suite.validateJWTMiddleware(w, req) {
				// Middleware rejected the request
				if w.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
				}
				return
			}

			// Call the handler
			suite.handler.ListCourts(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkContent {
				var courtSlots []*models.CourtSlot
				if err := json.Unmarshal(w.Body.Bytes(), &courtSlots); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if len(courtSlots) != tt.expectedCount {
					t.Errorf("Expected %d court slots, got %d. %s", tt.expectedCount, len(courtSlots), tt.description)
				}

				// Verify court slot structure and content
				if len(courtSlots) > 0 {
					slot := courtSlots[0]
					if slot.VenueName == "" {
						t.Error("Court slot venue name should not be empty")
					}
					if slot.CourtName == "" {
						t.Error("Court slot court name should not be empty")
					}
					if slot.Date == "" {
						t.Error("Court slot date should not be empty")
					}
					if slot.StartTime == "" {
						t.Error("Court slot start time should not be empty")
					}
					if slot.Price <= 0 {
						t.Error("Court slot price should be positive")
					}
					if slot.BookingURL == "" {
						t.Error("Court slot booking URL should not be empty")
					}
				}

				// Check content type
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}
			}
		})
	}
}

// TestEndpointPerformance tests response times and handles edge cases
func TestEndpointPerformance(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	t.Run("Venues endpoint performance", func(t *testing.T) {
		req := suite.createAuthenticatedRequest(http.MethodGet, "/api/venues", suite.validToken)
		w := httptest.NewRecorder()

		start := time.Now()

		// Simulate JWT middleware
		if !suite.validateJWTMiddleware(w, req) {
			t.Fatal("JWT validation failed")
		}

		suite.handler.ListVenues(w, req)
		duration := time.Since(start)

		if duration > 100*time.Millisecond {
			t.Errorf("Venues endpoint took too long: %v", duration)
		}

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Courts endpoint performance with filters", func(t *testing.T) {
		url := fmt.Sprintf("/api/courts?venueId=%s&date=2024-01-15&startTime=18:00", suite.testVenues[0].ID.Hex())
		req := suite.createAuthenticatedRequest(http.MethodGet, url, suite.validToken)
		w := httptest.NewRecorder()

		start := time.Now()

		// Simulate JWT middleware
		if !suite.validateJWTMiddleware(w, req) {
			t.Fatal("JWT validation failed")
		}

		suite.handler.ListCourts(w, req)
		duration := time.Since(start)

		if duration > 100*time.Millisecond {
			t.Errorf("Courts endpoint with filters took too long: %v", duration)
		}

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

// TestConcurrentRequests tests handling of concurrent requests
func TestConcurrentRequests(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	const numRequests = 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := suite.createAuthenticatedRequest(http.MethodGet, "/api/venues", suite.validToken)
			w := httptest.NewRecorder()

			// Simulate JWT middleware
			if !suite.validateJWTMiddleware(w, req) {
				results <- w.Code
				return
			}

			suite.handler.ListVenues(w, req)
			results <- w.Code
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		status := <-results
		if status != http.StatusOK {
			t.Errorf("Concurrent request failed with status %d", status)
		}
	}
}
