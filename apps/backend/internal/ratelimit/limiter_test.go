package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewLimiter tests the creation of a new rate limiter
func TestNewLimiter(t *testing.T) {
	config := DefaultConfig()

	// Note: This test requires Redis to be running
	// In a real environment, you might want to use a test container
	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	assert.NotNil(t, limiter)
	assert.Equal(t, config, limiter.GetConfig())
}

// TestLimiterHealthCheck tests the health check functionality
func TestLimiterHealthCheck(t *testing.T) {
	config := DefaultConfig()

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	ctx := context.Background()
	err = limiter.HealthCheck(ctx)
	assert.NoError(t, err)
}

// TestIPRateLimit tests IP-based rate limiting
func TestIPRateLimit(t *testing.T) {
	config := DefaultConfig()
	// Set a low limit for testing
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

	ctx := context.Background()
	testIP := "192.168.1.100"

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		result, err := limiter.CheckIPLimit(ctx, testIP)
		require.NoError(t, err)
		assert.True(t, result.Allowed, "Request %d should be allowed", i+1)
		assert.Equal(t, int64(3), result.Limit)
		assert.Equal(t, int64(3-i-1), result.Remaining)
	}

	// 4th request should be denied
	result, err := limiter.CheckIPLimit(ctx, testIP)
	require.NoError(t, err)
	assert.False(t, result.Allowed, "4th request should be denied")
	assert.Equal(t, int64(0), result.Remaining)
	assert.True(t, result.RetryAfter > 0, "RetryAfter should be set")
}

// TestUserRateLimit tests user-based rate limiting
func TestUserRateLimit(t *testing.T) {
	config := DefaultConfig()
	// Set a low limit for testing
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

	ctx := context.Background()
	testUserID := "user123"

	// First 2 requests should be allowed
	for i := 0; i < 2; i++ {
		result, err := limiter.CheckUserLimit(ctx, testUserID)
		require.NoError(t, err)
		assert.True(t, result.Allowed, "Request %d should be allowed", i+1)
		assert.Equal(t, int64(2), result.Limit)
		assert.Equal(t, int64(2-i-1), result.Remaining)
	}

	// 3rd request should be denied
	result, err := limiter.CheckUserLimit(ctx, testUserID)
	require.NoError(t, err)
	assert.False(t, result.Allowed, "3rd request should be denied")
	assert.Equal(t, int64(0), result.Remaining)
}

// TestAuthEndpointLimit tests authentication endpoint rate limiting
func TestAuthEndpointLimit(t *testing.T) {
	config := DefaultConfig()
	// Set a very low limit for testing
	config.AuthEndpointLimit = RateLimit{
		Requests: 1,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	ctx := context.Background()
	testIdentifier := "auth_test"

	// First request should be allowed
	result, err := limiter.CheckAuthLimit(ctx, testIdentifier)
	require.NoError(t, err)
	assert.True(t, result.Allowed, "First auth request should be allowed")
	assert.Equal(t, int64(1), result.Limit)
	assert.Equal(t, int64(0), result.Remaining)

	// Second request should be denied
	result, err = limiter.CheckAuthLimit(ctx, testIdentifier)
	require.NoError(t, err)
	assert.False(t, result.Allowed, "Second auth request should be denied")
}

// TestCustomRateLimit tests custom rate limiting
func TestCustomRateLimit(t *testing.T) {
	config := DefaultConfig()

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	ctx := context.Background()
	testIdentifier := "custom_test"
	customLimit := RateLimit{
		Requests: 5,
		Window:   time.Minute,
	}

	// Test custom rate limit
	for i := 0; i < 5; i++ {
		result, err := limiter.CheckCustomLimit(ctx, testIdentifier, customLimit)
		require.NoError(t, err)
		assert.True(t, result.Allowed, "Request %d should be allowed", i+1)
		assert.Equal(t, int64(5), result.Limit)
	}

	// 6th request should be denied
	result, err := limiter.CheckCustomLimit(ctx, testIdentifier, customLimit)
	require.NoError(t, err)
	assert.False(t, result.Allowed, "6th request should be denied")
}

// TestRateLimitReset tests the reset functionality
func TestRateLimitReset(t *testing.T) {
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

	ctx := context.Background()
	testIP := "192.168.1.200"

	// Use up the limit
	result, err := limiter.CheckIPLimit(ctx, testIP)
	require.NoError(t, err)
	assert.True(t, result.Allowed)

	// Next request should be denied
	result, err = limiter.CheckIPLimit(ctx, testIP)
	require.NoError(t, err)
	assert.False(t, result.Allowed)

	// Reset the limit
	err = limiter.Reset(ctx, "ip", testIP)
	require.NoError(t, err)

	// Request should now be allowed again
	result, err = limiter.CheckIPLimit(ctx, testIP)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

// TestDifferentEndpointLimits tests that different endpoints have different limits
func TestDifferentEndpointLimits(t *testing.T) {
	config := DefaultConfig()
	config.AuthEndpointLimit = RateLimit{Requests: 1, Window: time.Minute}
	config.DataEndpointLimit = RateLimit{Requests: 2, Window: time.Minute}
	config.SensitiveEndpointLimit = RateLimit{Requests: 1, Window: time.Minute}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	ctx := context.Background()
	testIdentifier := "endpoint_test"

	// Test auth endpoint (limit: 1)
	result, err := limiter.CheckAuthLimit(ctx, testIdentifier)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(1), result.Limit)

	// Test data endpoint (limit: 2)
	result, err = limiter.CheckDataLimit(ctx, testIdentifier)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(2), result.Limit)

	// Test sensitive endpoint (limit: 1)
	result, err = limiter.CheckSensitiveLimit(ctx, testIdentifier)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(1), result.Limit)
}

// TestConcurrentRequests tests rate limiting under concurrent load
func TestConcurrentRequests(t *testing.T) {
	config := DefaultConfig()
	config.DefaultIPLimit = RateLimit{
		Requests: 10,
		Window:   time.Minute,
	}

	limiter, err := NewLimiter(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer limiter.Close()

	ctx := context.Background()
	testIP := "192.168.1.300"

	// Run 20 concurrent requests (should only allow 10)
	results := make(chan bool, 20)
	for i := 0; i < 20; i++ {
		go func() {
			result, err := limiter.CheckIPLimit(ctx, testIP)
			if err != nil {
				results <- false
				return
			}
			results <- result.Allowed
		}()
	}

	// Collect results
	allowedCount := 0
	deniedCount := 0
	for i := 0; i < 20; i++ {
		allowed := <-results
		if allowed {
			allowedCount++
		} else {
			deniedCount++
		}
	}

	// Should have exactly 10 allowed and 10 denied
	assert.Equal(t, 10, allowedCount, "Should allow exactly 10 requests")
	assert.Equal(t, 10, deniedCount, "Should deny exactly 10 requests")
}

// TestDefaultConfig tests the default configuration
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "localhost:6379", config.RedisAddr)
	assert.Equal(t, "", config.RedisPassword)
	assert.Equal(t, 0, config.RedisDB)

	assert.Equal(t, 100, config.DefaultIPLimit.Requests)
	assert.Equal(t, time.Minute, config.DefaultIPLimit.Window)

	assert.Equal(t, 500, config.DefaultUserLimit.Requests)
	assert.Equal(t, time.Minute, config.DefaultUserLimit.Window)

	assert.Equal(t, 10, config.AuthEndpointLimit.Requests)
	assert.Equal(t, 200, config.DataEndpointLimit.Requests)
	assert.Equal(t, 5, config.SensitiveEndpointLimit.Requests)

	assert.True(t, config.IncludeHeaders)
	assert.Contains(t, config.TrustedProxies, "127.0.0.1")
}

// TestRateLimitString tests the String method of RateLimit
func TestRateLimitString(t *testing.T) {
	rl := RateLimit{
		Requests: 100,
		Window:   time.Minute,
	}

	expected := "100 requests per 1m0s"
	assert.Equal(t, expected, rl.String())
}
