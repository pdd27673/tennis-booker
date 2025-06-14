package ratelimit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tennis-booker/internal/auth"
)

// Test handlers that simulate real endpoints
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password required", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"access_token": fmt.Sprintf("mock-token-%s", req.Username),
		"expires_in":   900,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
	// For testing, we'll simulate a protected endpoint that requires user context
	// In a real scenario, this would be set by JWT middleware
	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"message":  "Protected resource accessed",
		"user_id":  claims.UserID,
		"username": claims.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TestRateLimitingIntegration tests the complete rate limiting integration
func TestRateLimitingIntegration(t *testing.T) {
	config := DefaultConfig()
	config.AuthEndpointLimit = RateLimit{
		Requests: 3,
		Window:   time.Minute,
	}
	config.DefaultIPLimit = RateLimit{
		Requests: 5,
		Window:   time.Minute,
	}
	config.DefaultUserLimit = RateLimit{
		Requests: 10,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping integration test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	// Create rate limiting middleware
	authRateLimit := AuthRateLimitMiddleware(limiter)
	ipRateLimit := IPRateLimitMiddleware(limiter)

	// Set up test server with rate-limited endpoints
	mux := http.NewServeMux()

	// Auth endpoint with auth rate limiting
	mux.Handle("/login", authRateLimit(http.HandlerFunc(loginHandler)))

	// Public endpoint with IP rate limiting
	mux.Handle("/health", ipRateLimit(http.HandlerFunc(healthHandler)))

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Run("Auth endpoint rate limiting", func(t *testing.T) {
		testAuthEndpointRateLimit(t, server.URL)
	})

	t.Run("Public endpoint rate limiting", func(t *testing.T) {
		testPublicEndpointRateLimit(t, server.URL)
	})

	t.Run("Rate limit headers", func(t *testing.T) {
		testRateLimitHeaders(t, server.URL)
	})
}

func testAuthEndpointRateLimit(t *testing.T, serverURL string) {
	client := &http.Client{}

	// Use unique timestamp for IP to avoid conflicts
	timestamp := time.Now().UnixNano()
	testIP := fmt.Sprintf("192.168.200.%d", timestamp%255)

	// First 3 requests should succeed (auth limit is 3/minute)
	for i := 0; i < 3; i++ {
		loginData := map[string]string{
			"username": fmt.Sprintf("testuser%d", i),
			"password": "testpass",
		}
		jsonData, _ := json.Marshal(loginData)

		req, err := http.NewRequest("POST", serverURL+"/login", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", testIP)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Auth request %d should succeed", i+1)

		// Check rate limit headers
		assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Limit"))
		assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Remaining"))
	}

	// 4th request should be rate limited
	loginData := map[string]string{
		"username": "testuser4",
		"password": "testpass",
	}
	jsonData, _ := json.Marshal(loginData)

	req, err := http.NewRequest("POST", serverURL+"/login", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", testIP)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("Retry-After"))
}

func testPublicEndpointRateLimit(t *testing.T, serverURL string) {
	client := &http.Client{}

	// Use unique timestamp for IP to avoid conflicts
	timestamp := time.Now().UnixNano()
	testIP := fmt.Sprintf("192.168.201.%d", timestamp%255)

	// First 5 requests should succeed (IP limit is 5/minute)
	for i := 0; i < 5; i++ {
		req, err := http.NewRequest("GET", serverURL+"/health", nil)
		require.NoError(t, err)
		req.Header.Set("X-Forwarded-For", testIP)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Health request %d should succeed", i+1)
	}

	// 6th request should be rate limited
	req, err := http.NewRequest("GET", serverURL+"/health", nil)
	require.NoError(t, err)
	req.Header.Set("X-Forwarded-For", testIP)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}

func testRateLimitHeaders(t *testing.T, serverURL string) {
	client := &http.Client{}

	// Use unique timestamp for IP to avoid conflicts
	timestamp := time.Now().UnixNano()
	testIP := fmt.Sprintf("192.168.202.%d", timestamp%255)

	req, err := http.NewRequest("GET", serverURL+"/health", nil)
	require.NoError(t, err)
	req.Header.Set("X-Forwarded-For", testIP)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check that rate limit headers are present
	assert.Equal(t, "5", resp.Header.Get("X-RateLimit-Limit"))
	assert.Equal(t, "4", resp.Header.Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Reset"))
	assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Reset-Time"))
}

// TestUserRateLimitingWithMockContext tests user-based rate limiting with simulated user context
func TestUserRateLimitingWithMockContext(t *testing.T) {
	config := DefaultConfig()
	config.DefaultUserLimit = RateLimit{
		Requests: 3,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping user rate limiting test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	userRateLimit := UserRateLimitMiddleware(limiter)

	// Create a test handler that simulates having user context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User endpoint accessed"))
	})

	// Use a consistent user ID for all requests in this test
	timestamp := time.Now().UnixNano()
	testUserID := fmt.Sprintf("testuser%d", timestamp)

	// Create middleware that adds mock user context with consistent user ID
	mockUserMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate user context being set by JWT middleware
			claims := &auth.AppClaims{
				UserID:   testUserID,
				Username: "testuser",
			}
			ctx := auth.SetUserClaimsInContext(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// Chain the middleware: mock user context -> user rate limit -> handler
	handler := mockUserMiddleware(userRateLimit(testHandler))

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/user-endpoint", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "User request %d should succeed", i+1)
		assert.Equal(t, "User endpoint accessed", w.Body.String())
	}

	// 4th request should be rate limited
	req := httptest.NewRequest("GET", "/user-endpoint", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many requests for user")
}

// Helper function to get response body as string
func getResponseBody(t *testing.T, resp *http.Response) string {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(resp.Body)
	require.NoError(t, err)
	return buf.String()
}
