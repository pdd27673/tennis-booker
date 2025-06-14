package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tennis-booker/internal/auth"
)

func TestSystemHandler_Health(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		version        string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "successful health check",
			method:         http.MethodGet,
			version:        "1.0.0",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"status":  "UP",
				"version": "1.0.0",
			},
		},
		{
			name:           "successful health check without version",
			method:         http.MethodGet,
			version:        "",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"status": "UP",
			},
		},
		{
			name:           "method not allowed - POST",
			method:         http.MethodPost,
			version:        "1.0.0",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil, // Error response is plain text
		},
		{
			name:           "method not allowed - PUT",
			method:         http.MethodPut,
			version:        "1.0.0",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
		},
		{
			name:           "method not allowed - DELETE",
			method:         http.MethodDelete,
			version:        "1.0.0",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			handler := NewSystemHandler(tt.version)

			// Create request
			req := httptest.NewRequest(tt.method, "/api/health", nil)
			w := httptest.NewRecorder()

			// Call handler
			handler.Health(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check Allow header for method not allowed responses
			if tt.expectedStatus == http.StatusMethodNotAllowed {
				allowHeader := w.Header().Get("Allow")
				if allowHeader != "GET" {
					t.Errorf("expected Allow header 'GET', got '%s'", allowHeader)
				}
				return // Skip JSON parsing for error responses
			}

			// Check content type for successful responses
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
			}

			// Parse response body
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to parse response body: %v", err)
			}

			// Check required fields
			if status, ok := response["status"].(string); !ok || status != tt.expectedBody["status"] {
				t.Errorf("expected status '%s', got '%v'", tt.expectedBody["status"], response["status"])
			}

			// Check timestamp is present and recent
			if timestampStr, ok := response["timestamp"].(string); !ok {
				t.Error("expected timestamp field to be present")
			} else {
				timestamp, err := time.Parse(time.RFC3339, timestampStr)
				if err != nil {
					t.Errorf("expected valid timestamp, got '%s'", timestampStr)
				} else {
					// Check timestamp is within last 5 seconds
					if time.Since(timestamp) > 5*time.Second {
						t.Errorf("timestamp seems too old: %v", timestamp)
					}
				}
			}

			// Check version field
			if tt.version != "" {
				if version, ok := response["version"].(string); !ok || version != tt.version {
					t.Errorf("expected version '%s', got '%v'", tt.version, response["version"])
				}
			} else {
				// Version should not be present if empty
				if _, exists := response["version"]; exists {
					t.Error("expected version field to be omitted when empty")
				}
			}
		})
	}
}

func TestNewSystemHandler(t *testing.T) {
	handler := NewSystemHandler("1.0.0")
	if handler == nil {
		t.Fatal("expected handler to be created")
	}
	if handler.version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", handler.version)
	}
}

func TestSystemHandler_Status(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		withAuth       bool
		expectedStatus int
	}{
		{
			name:           "successful status request",
			method:         http.MethodGet,
			withAuth:       true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing authentication",
			method:         http.MethodGet,
			withAuth:       false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "method not allowed - POST",
			method:         http.MethodPost,
			withAuth:       true,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "method not allowed - PUT",
			method:         http.MethodPut,
			withAuth:       true,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewSystemHandler("1.0.0")
			req := httptest.NewRequest(tt.method, "/api/system/status", nil)

			// Add authentication if required
			if tt.withAuth {
				claims := &auth.AppClaims{
					UserID:   "test-user-id",
					Username: "testuser",
				}
				ctx := auth.SetUserClaimsInContext(req.Context(), claims)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			handler.Status(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusMethodNotAllowed {
				allowHeader := w.Header().Get("Allow")
				if allowHeader != "GET" {
					t.Errorf("expected Allow header 'GET', got '%s'", allowHeader)
				}
				return
			}

			if tt.expectedStatus != http.StatusOK {
				return // Skip JSON parsing for error responses
			}

			// Check content type for successful responses
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
			}

			// Parse response body
			var response SystemStatusResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to parse response body: %v", err)
			}

			// Check required fields
			if response.ScrapingStatus == "" {
				t.Error("expected scraping_status to be present")
			}

			if response.ItemsProcessed < 0 {
				t.Errorf("expected items_processed to be non-negative, got %d", response.ItemsProcessed)
			}

			if response.ErrorCount < 0 {
				t.Errorf("expected error_count to be non-negative, got %d", response.ErrorCount)
			}

			if response.SystemUptime == "" {
				t.Error("expected system_uptime to be present")
			}

			// Check timestamp is present and recent
			if response.Timestamp.IsZero() {
				t.Error("expected timestamp to be present")
			} else {
				if time.Since(response.Timestamp) > 5*time.Second {
					t.Errorf("timestamp seems too old: %v", response.Timestamp)
				}
			}
		})
	}
}

func TestSystemHandler_Pause(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		withAuth        bool
		initialStatus   string
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "successful pause from running",
			method:          http.MethodPost,
			withAuth:        true,
			initialStatus:   "running",
			expectedStatus:  http.StatusOK,
			expectedMessage: "System pausing initiated",
		},
		{
			name:            "pause already paused system",
			method:          http.MethodPost,
			withAuth:        true,
			initialStatus:   "paused",
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "System is already paused",
		},
		{
			name:            "missing authentication",
			method:          http.MethodPost,
			withAuth:        false,
			initialStatus:   "running",
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "",
		},
		{
			name:            "method not allowed - GET",
			method:          http.MethodGet,
			withAuth:        true,
			initialStatus:   "running",
			expectedStatus:  http.StatusMethodNotAllowed,
			expectedMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewSystemHandler("1.0.0")
			handler.scrapingStatus = tt.initialStatus

			req := httptest.NewRequest(tt.method, "/api/system/pause", nil)

			if tt.withAuth {
				claims := &auth.AppClaims{
					UserID:   "test-user-id",
					Username: "testuser",
				}
				ctx := auth.SetUserClaimsInContext(req.Context(), claims)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			handler.Pause(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusMethodNotAllowed {
				allowHeader := w.Header().Get("Allow")
				if allowHeader != "POST" {
					t.Errorf("expected Allow header 'POST', got '%s'", allowHeader)
				}
				return
			}

			if tt.expectedStatus != http.StatusOK {
				return // Skip JSON parsing for error responses
			}

			// Check content type
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
			}

			// Parse response body
			var response SystemControlResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to parse response body: %v", err)
			}

			// Check response fields
			if response.Message != tt.expectedMessage {
				t.Errorf("expected message '%s', got '%s'", tt.expectedMessage, response.Message)
			}

			if response.Status != "paused" {
				t.Errorf("expected status 'paused', got '%s'", response.Status)
			}

			// Verify handler state was updated
			if handler.scrapingStatus != "paused" {
				t.Errorf("expected handler status 'paused', got '%s'", handler.scrapingStatus)
			}
		})
	}
}

func TestSystemHandler_Resume(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		withAuth        bool
		initialStatus   string
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "successful resume from paused",
			method:          http.MethodPost,
			withAuth:        true,
			initialStatus:   "paused",
			expectedStatus:  http.StatusOK,
			expectedMessage: "System resuming initiated",
		},
		{
			name:            "resume already running system",
			method:          http.MethodPost,
			withAuth:        true,
			initialStatus:   "running",
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "System is already running",
		},
		{
			name:            "missing authentication",
			method:          http.MethodPost,
			withAuth:        false,
			initialStatus:   "paused",
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "",
		},
		{
			name:            "method not allowed - GET",
			method:          http.MethodGet,
			withAuth:        true,
			initialStatus:   "paused",
			expectedStatus:  http.StatusMethodNotAllowed,
			expectedMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewSystemHandler("1.0.0")
			handler.scrapingStatus = tt.initialStatus

			req := httptest.NewRequest(tt.method, "/api/system/resume", nil)

			if tt.withAuth {
				claims := &auth.AppClaims{
					UserID:   "test-user-id",
					Username: "testuser",
				}
				ctx := auth.SetUserClaimsInContext(req.Context(), claims)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			handler.Resume(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusMethodNotAllowed {
				allowHeader := w.Header().Get("Allow")
				if allowHeader != "POST" {
					t.Errorf("expected Allow header 'POST', got '%s'", allowHeader)
				}
				return
			}

			if tt.expectedStatus != http.StatusOK {
				return // Skip JSON parsing for error responses
			}

			// Check content type
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
			}

			// Parse response body
			var response SystemControlResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to parse response body: %v", err)
			}

			// Check response fields
			if response.Message != tt.expectedMessage {
				t.Errorf("expected message '%s', got '%s'", tt.expectedMessage, response.Message)
			}

			if response.Status != "running" {
				t.Errorf("expected status 'running', got '%s'", response.Status)
			}

			// Verify handler state was updated
			if handler.scrapingStatus != "running" {
				t.Errorf("expected handler status 'running', got '%s'", handler.scrapingStatus)
			}
		})
	}
}

func TestSystemHandler_PauseResumeIntegration(t *testing.T) {
	handler := NewSystemHandler("1.0.0")

	// Helper function to create authenticated request
	createAuthRequest := func(method, path string) *http.Request {
		req := httptest.NewRequest(method, path, nil)
		claims := &auth.AppClaims{
			UserID:   "test-user-id",
			Username: "testuser",
		}
		ctx := auth.SetUserClaimsInContext(req.Context(), claims)
		return req.WithContext(ctx)
	}

	// Test initial status (should be running)
	req := createAuthRequest(http.MethodGet, "/api/system/status")
	w := httptest.NewRecorder()
	handler.Status(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var statusResp SystemStatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &statusResp); err != nil {
		t.Fatalf("failed to parse status response: %v", err)
	}

	if statusResp.ScrapingStatus != "running" {
		t.Errorf("expected initial status 'running', got '%s'", statusResp.ScrapingStatus)
	}

	// Test pause
	req = createAuthRequest(http.MethodPost, "/api/system/pause")
	w = httptest.NewRecorder()
	handler.Pause(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected pause status 200, got %d", w.Code)
	}

	var pauseResp SystemControlResponse
	if err := json.Unmarshal(w.Body.Bytes(), &pauseResp); err != nil {
		t.Fatalf("failed to parse pause response: %v", err)
	}

	if pauseResp.Status != "paused" {
		t.Errorf("expected pause status 'paused', got '%s'", pauseResp.Status)
	}

	// Verify status shows paused
	req = createAuthRequest(http.MethodGet, "/api/system/status")
	w = httptest.NewRecorder()
	handler.Status(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if err := json.Unmarshal(w.Body.Bytes(), &statusResp); err != nil {
		t.Fatalf("failed to parse status response: %v", err)
	}

	if statusResp.ScrapingStatus != "paused" {
		t.Errorf("expected status after pause 'paused', got '%s'", statusResp.ScrapingStatus)
	}

	// Test resume
	req = createAuthRequest(http.MethodPost, "/api/system/resume")
	w = httptest.NewRecorder()
	handler.Resume(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected resume status 200, got %d", w.Code)
	}

	var resumeResp SystemControlResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resumeResp); err != nil {
		t.Fatalf("failed to parse resume response: %v", err)
	}

	if resumeResp.Status != "running" {
		t.Errorf("expected resume status 'running', got '%s'", resumeResp.Status)
	}

	// Verify status shows running again
	req = createAuthRequest(http.MethodGet, "/api/system/status")
	w = httptest.NewRecorder()
	handler.Status(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if err := json.Unmarshal(w.Body.Bytes(), &statusResp); err != nil {
		t.Fatalf("failed to parse status response: %v", err)
	}

	if statusResp.ScrapingStatus != "running" {
		t.Errorf("expected status after resume 'running', got '%s'", statusResp.ScrapingStatus)
	}
}
