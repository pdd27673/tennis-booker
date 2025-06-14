package ratelimit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"tennis-booker/internal/auth"
)

// TestExtractClientIP tests the IP extraction logic with various header combinations
func TestExtractClientIP(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		headers        map[string]string
		trustedProxies []string
		expectedIP     string
	}{
		{
			name:       "Direct connection - RemoteAddr only",
			remoteAddr: "192.168.1.100:12345",
			headers:    map[string]string{},
			expectedIP: "192.168.1.100",
		},
		{
			name:       "X-Forwarded-For single IP",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1",
			},
			expectedIP: "203.0.113.1",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1, 198.51.100.1, 10.0.0.1",
			},
			trustedProxies: []string{"10.0.0.1", "198.51.100.1"},
			expectedIP:     "203.0.113.1",
		},
		{
			name:       "X-Real-IP header",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.2",
			},
			expectedIP: "203.0.113.2",
		},
		{
			name:       "X-Client-IP header",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Client-IP": "203.0.113.3",
			},
			expectedIP: "203.0.113.3",
		},
		{
			name:       "CF-Connecting-IP header (Cloudflare)",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.4",
			},
			expectedIP: "203.0.113.4",
		},
		{
			name:       "X-Forwarded-For takes precedence",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1",
				"X-Real-IP":       "203.0.113.2",
				"X-Client-IP":     "203.0.113.3",
			},
			expectedIP: "203.0.113.1",
		},
		{
			name:       "IPv6 address",
			remoteAddr: "[2001:db8::1]:12345",
			headers:    map[string]string{},
			expectedIP: "2001:db8::1",
		},
		{
			name:       "RemoteAddr without port",
			remoteAddr: "192.168.1.100",
			headers:    map[string]string{},
			expectedIP: "192.168.1.100",
		},
		{
			name:       "Invalid IP in header falls back to RemoteAddr",
			remoteAddr: "192.168.1.100:12345",
			headers: map[string]string{
				"X-Forwarded-For": "invalid-ip",
			},
			expectedIP: "192.168.1.100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			if tt.trustedProxies == nil {
				tt.trustedProxies = []string{"127.0.0.1", "::1"}
			}

			result := extractClientIP(req, tt.trustedProxies)
			assert.Equal(t, tt.expectedIP, result)
		})
	}
}

// TestIPRateLimitMiddleware tests the IP-based rate limiting middleware
func TestIPRateLimitMiddleware(t *testing.T) {
	config := DefaultConfig()
	config.DefaultIPLimit = RateLimit{
		Requests: 3,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := IPRateLimitMiddleware(limiter)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := middleware(testHandler)

	// Use a unique test IP with timestamp to avoid conflicts
	testIP := fmt.Sprintf("192.168.100.%d", time.Now().Unix()%255)

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = testIP + ":12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
		assert.Equal(t, "OK", w.Body.String())

		// Check rate limit headers
		assert.Equal(t, "3", w.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, strconv.Itoa(3-i-1), w.Header().Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset-Time"))
	}

	// 4th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = testIP + ":12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many requests")
	assert.NotEmpty(t, w.Header().Get("Retry-After"))
	assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
}

// TestAuthRateLimitMiddleware tests the authentication endpoint rate limiting
func TestAuthRateLimitMiddleware(t *testing.T) {
	config := DefaultConfig()
	config.AuthEndpointLimit = RateLimit{
		Requests: 2,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := AuthRateLimitMiddleware(limiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Auth OK"))
	})

	handler := middleware(testHandler)

	testIP := fmt.Sprintf("192.168.101.%d", time.Now().Unix()%255)

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/auth/login", nil)
		req.RemoteAddr = testIP + ":12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Auth request %d should succeed", i+1)
		assert.Equal(t, "Auth OK", w.Body.String())
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest("POST", "/auth/login", nil)
	req.RemoteAddr = testIP + ":12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many authentication requests")
}

// TestDataRateLimitMiddleware tests the data endpoint rate limiting
func TestDataRateLimitMiddleware(t *testing.T) {
	config := DefaultConfig()
	config.DataEndpointLimit = RateLimit{
		Requests: 5,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := DataRateLimitMiddleware(limiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Data OK"))
	})

	handler := middleware(testHandler)

	testIP := fmt.Sprintf("192.168.102.%d", time.Now().Unix()%255)

	// First 5 requests should succeed
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/data", nil)
		req.RemoteAddr = testIP + ":12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Data request %d should succeed", i+1)
		assert.Equal(t, "Data OK", w.Body.String())
	}

	// 6th request should be rate limited
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.RemoteAddr = testIP + ":12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many data requests")
}

// TestSensitiveRateLimitMiddleware tests the sensitive endpoint rate limiting
func TestSensitiveRateLimitMiddleware(t *testing.T) {
	config := DefaultConfig()
	config.SensitiveEndpointLimit = RateLimit{
		Requests: 1,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := SensitiveRateLimitMiddleware(limiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Sensitive OK"))
	})

	handler := middleware(testHandler)

	testIP := fmt.Sprintf("192.168.103.%d", time.Now().Unix()%255)

	// First request should succeed
	req := httptest.NewRequest("POST", "/api/sensitive", nil)
	req.RemoteAddr = testIP + ":12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Sensitive OK", w.Body.String())

	// 2nd request should be rate limited
	req = httptest.NewRequest("POST", "/api/sensitive", nil)
	req.RemoteAddr = testIP + ":12345"
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many sensitive requests")
}

// TestCustomRateLimitMiddleware tests the custom rate limiting middleware
func TestCustomRateLimitMiddleware(t *testing.T) {
	config := DefaultConfig()

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	customLimit := RateLimit{
		Requests: 2,
		Window:   time.Minute,
	}

	middleware := CustomRateLimitMiddleware(limiter, customLimit, "custom")

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Custom OK"))
	})

	handler := middleware(testHandler)

	testIP := fmt.Sprintf("192.168.104.%d", time.Now().Unix()%255)

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/api/custom", nil)
		req.RemoteAddr = testIP + ":12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Custom request %d should succeed", i+1)
		assert.Equal(t, "Custom OK", w.Body.String())
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest("GET", "/api/custom", nil)
	req.RemoteAddr = testIP + ":12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Rate limit exceeded")
}

// TestMiddlewareWithDifferentIPs tests that different IPs have separate rate limits
func TestMiddlewareWithDifferentIPs(t *testing.T) {
	config := DefaultConfig()
	config.DefaultIPLimit = RateLimit{
		Requests: 2,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := IPRateLimitMiddleware(limiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := middleware(testHandler)

	timestamp := time.Now().Unix()
	ip1 := fmt.Sprintf("192.168.105.%d", timestamp%255)
	ip2 := fmt.Sprintf("192.168.106.%d", (timestamp+1)%255)

	// Use up limit for IP1
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip1 + ":12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// IP1 should now be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = ip1 + ":12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// IP2 should still work
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = ip2 + ":12345"
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestMiddlewareWithXForwardedFor tests IP extraction from X-Forwarded-For header
func TestMiddlewareWithXForwardedFor(t *testing.T) {
	config := DefaultConfig()
	config.DefaultIPLimit = RateLimit{
		Requests: 1,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := IPRateLimitMiddleware(limiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := middleware(testHandler)

	// Use unique IP with X-Forwarded-For header
	clientIP := fmt.Sprintf("203.0.113.%d", time.Now().Unix()%255)

	// First request with X-Forwarded-For should succeed
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", clientIP)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Second request with same X-Forwarded-For IP should be rate limited
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.2:12345"           // Different proxy IP
	req.Header.Set("X-Forwarded-For", clientIP) // Same client IP
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

// TestRateLimitHeaders tests that rate limit headers are correctly set
func TestRateLimitHeaders(t *testing.T) {
	config := DefaultConfig()
	config.DefaultIPLimit = RateLimit{
		Requests: 5,
		Window:   time.Minute,
	}
	config.IncludeHeaders = true

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := IPRateLimitMiddleware(limiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("192.168.107.%d:12345", time.Now().Unix()%255)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check that rate limit headers are present
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "4", w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset-Time"))

	// Verify reset time is a valid Unix timestamp
	resetTime := w.Header().Get("X-RateLimit-Reset")
	_, err = strconv.ParseInt(resetTime, 10, 64)
	assert.NoError(t, err, "Reset time should be a valid Unix timestamp")

	// Verify reset time is in RFC3339 format
	resetTimeFormatted := w.Header().Get("X-RateLimit-Reset-Time")
	_, err = time.Parse(time.RFC3339, resetTimeFormatted)
	assert.NoError(t, err, "Reset time should be in RFC3339 format")
}

// TestMiddlewareWithoutHeaders tests that headers are not included when disabled
func TestMiddlewareWithoutHeaders(t *testing.T) {
	config := DefaultConfig()
	config.DefaultIPLimit = RateLimit{
		Requests: 5,
		Window:   time.Minute,
	}
	config.IncludeHeaders = false

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := IPRateLimitMiddleware(limiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("192.168.108.%d:12345", time.Now().Unix()%255)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check that rate limit headers are NOT present
	assert.Empty(t, w.Header().Get("X-RateLimit-Limit"))
	assert.Empty(t, w.Header().Get("X-RateLimit-Remaining"))
	assert.Empty(t, w.Header().Get("X-RateLimit-Reset"))
	assert.Empty(t, w.Header().Get("X-RateLimit-Reset-Time"))
}

// TestUserRateLimitMiddleware tests the user-based rate limiting middleware
func TestUserRateLimitMiddleware(t *testing.T) {
	config := DefaultConfig()
	config.DefaultUserLimit = RateLimit{
		Requests: 3,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := UserRateLimitMiddleware(limiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User OK"))
	})

	handler := middleware(testHandler)

	// Create test user claims
	testUserID := fmt.Sprintf("user-%d", time.Now().Unix())
	testClaims := &auth.AppClaims{
		UserID:   testUserID,
		Username: "testuser",
	}

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.200:12345"

		// Add user claims to context (simulating JWT middleware)
		ctx := auth.SetUserClaimsInContext(req.Context(), testClaims)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "User request %d should succeed", i+1)
		assert.Equal(t, "User OK", w.Body.String())

		// Check rate limit headers
		assert.Equal(t, "3", w.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, strconv.Itoa(3-i-1), w.Header().Get("X-RateLimit-Remaining"))
	}

	// 4th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.200:12345"
	ctx := auth.SetUserClaimsInContext(req.Context(), testClaims)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many requests for user")
	assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
}

// TestUserRateLimitMiddleware_FallbackToIP tests fallback to IP limiting when no user context
func TestUserRateLimitMiddleware_FallbackToIP(t *testing.T) {
	config := DefaultConfig()
	config.DefaultIPLimit = RateLimit{
		Requests: 2,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := UserRateLimitMiddleware(limiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Fallback OK"))
	})

	handler := middleware(testHandler)

	testIP := fmt.Sprintf("192.168.109.%d", time.Now().Unix()%255)

	// First 2 requests should succeed (using IP limit since no user context)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = testIP + ":12345"
		// No user context - should fall back to IP limiting
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Fallback request %d should succeed", i+1)
		assert.Equal(t, "Fallback OK", w.Body.String())
	}

	// 3rd request should be rate limited (IP limit is 2)
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = testIP + ":12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many requests")
}

// TestCombinedRateLimitMiddleware tests the combined IP and user rate limiting
func TestCombinedRateLimitMiddleware(t *testing.T) {
	config := DefaultConfig()
	config.DefaultIPLimit = RateLimit{
		Requests: 5,
		Window:   time.Minute,
	}
	config.DefaultUserLimit = RateLimit{
		Requests: 3,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := CombinedRateLimitMiddleware(limiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Combined OK"))
	})

	handler := middleware(testHandler)

	// Use unique timestamp-based user ID to avoid conflicts
	timestamp := time.Now().UnixNano()
	testUserID := fmt.Sprintf("combined-user-%d", timestamp)
	testClaims := &auth.AppClaims{
		UserID:   testUserID,
		Username: "testuser",
	}

	testIP := fmt.Sprintf("192.168.110.%d", timestamp%255)

	// First 3 requests should succeed (limited by user limit which is lower)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = testIP + ":12345"
		ctx := auth.SetUserClaimsInContext(req.Context(), testClaims)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Combined request %d should succeed", i+1)
		assert.Equal(t, "Combined OK", w.Body.String())
	}

	// 4th request should be rate limited by user limit
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = testIP + ":12345"
	ctx := auth.SetUserClaimsInContext(req.Context(), testClaims)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many requests for user")
}

// TestCombinedRateLimitMiddleware_IPLimitFirst tests IP limit being hit before user limit
func TestCombinedRateLimitMiddleware_IPLimitFirst(t *testing.T) {
	config := DefaultConfig()
	config.DefaultIPLimit = RateLimit{
		Requests: 2,
		Window:   time.Minute,
	}
	config.DefaultUserLimit = RateLimit{
		Requests: 5,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := CombinedRateLimitMiddleware(limiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Combined OK"))
	})

	handler := middleware(testHandler)

	// Use unique timestamp-based user ID to avoid conflicts
	timestamp := time.Now().UnixNano()
	testUserID := fmt.Sprintf("combined-ip-user-%d", timestamp)
	testClaims := &auth.AppClaims{
		UserID:   testUserID,
		Username: "testuser",
	}

	testIP := fmt.Sprintf("192.168.111.%d", timestamp%255)

	// First 2 requests should succeed (limited by IP limit which is lower)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = testIP + ":12345"
		ctx := auth.SetUserClaimsInContext(req.Context(), testClaims)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Combined request %d should succeed", i+1)
		assert.Equal(t, "Combined OK", w.Body.String())
	}

	// 3rd request should be rate limited by IP limit
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = testIP + ":12345"
	ctx := auth.SetUserClaimsInContext(req.Context(), testClaims)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many requests from IP")
}

// TestUserAuthRateLimitMiddleware tests the user-based auth endpoint rate limiting
func TestUserAuthRateLimitMiddleware(t *testing.T) {
	config := DefaultConfig()
	config.AuthEndpointLimit = RateLimit{
		Requests: 2,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := UserAuthRateLimitMiddleware(limiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Auth OK"))
	})

	handler := middleware(testHandler)

	testUserID := fmt.Sprintf("user-%d", time.Now().Unix())
	testClaims := &auth.AppClaims{
		UserID:   testUserID,
		Username: "testuser",
	}

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/auth/refresh", nil)
		req.RemoteAddr = "192.168.1.300:12345"
		ctx := auth.SetUserClaimsInContext(req.Context(), testClaims)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "User auth request %d should succeed", i+1)
		assert.Equal(t, "Auth OK", w.Body.String())
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.RemoteAddr = "192.168.1.300:12345"
	ctx := auth.SetUserClaimsInContext(req.Context(), testClaims)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many authentication requests")
}

// TestUserAuthRateLimitMiddleware_NoUserContext tests auth middleware without user context
func TestUserAuthRateLimitMiddleware_NoUserContext(t *testing.T) {
	config := DefaultConfig()
	config.AuthEndpointLimit = RateLimit{
		Requests: 2,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := UserAuthRateLimitMiddleware(limiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Login OK"))
	})

	handler := middleware(testHandler)

	testIP := fmt.Sprintf("192.168.113.%d", time.Now().Unix()%255)

	// First 2 requests should succeed (using IP-based auth limiting)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/auth/login", nil)
		req.RemoteAddr = testIP + ":12345"
		// No user context (typical for login endpoint)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Login request %d should succeed", i+1)
		assert.Equal(t, "Login OK", w.Body.String())
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest("POST", "/auth/login", nil)
	req.RemoteAddr = testIP + ":12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many authentication requests")
}

// TestUserRateLimitMiddleware_DifferentUsers tests that different users have separate limits
func TestUserRateLimitMiddleware_DifferentUsers(t *testing.T) {
	config := DefaultConfig()
	config.DefaultUserLimit = RateLimit{
		Requests: 2,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := UserRateLimitMiddleware(limiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User OK"))
	})

	handler := middleware(testHandler)

	timestamp := time.Now().Unix()
	user1ID := fmt.Sprintf("user1-%d", timestamp)
	user2ID := fmt.Sprintf("user2-%d", timestamp)

	user1Claims := &auth.AppClaims{
		UserID:   user1ID,
		Username: "user1",
	}

	user2Claims := &auth.AppClaims{
		UserID:   user2ID,
		Username: "user2",
	}

	// Use up limit for user1
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.400:12345"
		ctx := auth.SetUserClaimsInContext(req.Context(), user1Claims)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// User1 should now be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.400:12345"
	ctx := auth.SetUserClaimsInContext(req.Context(), user1Claims)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// User2 should still work
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.400:12345"
	ctx = auth.SetUserClaimsInContext(req.Context(), user2Claims)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestCombinedRateLimitMiddleware_UnauthenticatedUser tests combined middleware with no user context
func TestCombinedRateLimitMiddleware_UnauthenticatedUser(t *testing.T) {
	config := DefaultConfig()
	config.DefaultIPLimit = RateLimit{
		Requests: 2,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	middleware := CombinedRateLimitMiddleware(limiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Unauth OK"))
	})

	handler := middleware(testHandler)

	testIP := fmt.Sprintf("192.168.112.%d", time.Now().UnixNano()%255)

	// First 2 requests should succeed (only IP limiting applies)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = testIP + ":12345"
		// No user context
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Unauth request %d should succeed", i+1)
		assert.Equal(t, "Unauth OK", w.Body.String())
	}

	// 3rd request should be rate limited by IP limit
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = testIP + ":12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many requests from IP")
}
