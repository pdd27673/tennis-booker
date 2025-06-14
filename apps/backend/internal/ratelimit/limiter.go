package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
)

// Limiter wraps the ulule/limiter with our configuration
type Limiter struct {
	config      *Config
	redisClient *redis.Client
	ipLimiter   *limiter.Limiter
	userLimiter *limiter.Limiter

	// Endpoint-specific limiters
	authLimiter      *limiter.Limiter
	dataLimiter      *limiter.Limiter
	sensitiveLimiter *limiter.Limiter
}

// LimitResult contains the result of a rate limit check
type LimitResult struct {
	Allowed    bool
	Limit      int64
	Remaining  int64
	ResetTime  time.Time
	RetryAfter time.Duration
}

// NewLimiter creates a new rate limiter with Redis backend
func NewLimiter(config *Config) (*Limiter, error) {
	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Create Redis store for limiter
	store, err := redisstore.NewStore(redisClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis store: %w", err)
	}

	// Create limiters for different use cases
	ipLimiter := limiter.New(store, limiter.Rate{
		Period: config.DefaultIPLimit.Window,
		Limit:  int64(config.DefaultIPLimit.Requests),
	})

	userLimiter := limiter.New(store, limiter.Rate{
		Period: config.DefaultUserLimit.Window,
		Limit:  int64(config.DefaultUserLimit.Requests),
	})

	authLimiter := limiter.New(store, limiter.Rate{
		Period: config.AuthEndpointLimit.Window,
		Limit:  int64(config.AuthEndpointLimit.Requests),
	})

	dataLimiter := limiter.New(store, limiter.Rate{
		Period: config.DataEndpointLimit.Window,
		Limit:  int64(config.DataEndpointLimit.Requests),
	})

	sensitiveLimiter := limiter.New(store, limiter.Rate{
		Period: config.SensitiveEndpointLimit.Window,
		Limit:  int64(config.SensitiveEndpointLimit.Requests),
	})

	return &Limiter{
		config:           config,
		redisClient:      redisClient,
		ipLimiter:        ipLimiter,
		userLimiter:      userLimiter,
		authLimiter:      authLimiter,
		dataLimiter:      dataLimiter,
		sensitiveLimiter: sensitiveLimiter,
	}, nil
}

// CheckIPLimit checks rate limit for an IP address
func (l *Limiter) CheckIPLimit(ctx context.Context, ip string) (*LimitResult, error) {
	return l.checkLimit(ctx, l.ipLimiter, fmt.Sprintf("ip:%s", ip))
}

// CheckUserLimit checks rate limit for a user
func (l *Limiter) CheckUserLimit(ctx context.Context, userID string) (*LimitResult, error) {
	return l.checkLimit(ctx, l.userLimiter, fmt.Sprintf("user:%s", userID))
}

// CheckAuthLimit checks rate limit for authentication endpoints
func (l *Limiter) CheckAuthLimit(ctx context.Context, identifier string) (*LimitResult, error) {
	return l.checkLimit(ctx, l.authLimiter, fmt.Sprintf("auth:%s", identifier))
}

// CheckDataLimit checks rate limit for data endpoints
func (l *Limiter) CheckDataLimit(ctx context.Context, identifier string) (*LimitResult, error) {
	return l.checkLimit(ctx, l.dataLimiter, fmt.Sprintf("data:%s", identifier))
}

// CheckSensitiveLimit checks rate limit for sensitive endpoints
func (l *Limiter) CheckSensitiveLimit(ctx context.Context, identifier string) (*LimitResult, error) {
	return l.checkLimit(ctx, l.sensitiveLimiter, fmt.Sprintf("sensitive:%s", identifier))
}

// CheckCustomLimit checks rate limit with custom configuration
func (l *Limiter) CheckCustomLimit(ctx context.Context, identifier string, rateLimit RateLimit) (*LimitResult, error) {
	// Create Redis store for custom limiter
	store, err := redisstore.NewStore(l.redisClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis store: %w", err)
	}

	customLimiter := limiter.New(store, limiter.Rate{
		Period: rateLimit.Window,
		Limit:  int64(rateLimit.Requests),
	})

	return l.checkLimit(ctx, customLimiter, identifier)
}

// checkLimit is a helper method that performs the actual rate limit check
func (l *Limiter) checkLimit(ctx context.Context, lim *limiter.Limiter, key string) (*LimitResult, error) {
	limitContext, err := lim.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}

	resetTime := time.Unix(limitContext.Reset, 0)
	result := &LimitResult{
		Allowed:   limitContext.Reached == false,
		Limit:     limitContext.Limit,
		Remaining: limitContext.Remaining,
		ResetTime: resetTime,
	}

	// Calculate retry after duration if limit exceeded
	if !result.Allowed {
		result.RetryAfter = time.Until(resetTime)
		if result.RetryAfter < 0 {
			result.RetryAfter = 0
		}
	}

	return result, nil
}

// Reset resets the rate limit for a specific key
func (l *Limiter) Reset(ctx context.Context, limiterType, identifier string) error {
	var key string
	var lim *limiter.Limiter

	switch limiterType {
	case "ip":
		key = fmt.Sprintf("ip:%s", identifier)
		lim = l.ipLimiter
	case "user":
		key = fmt.Sprintf("user:%s", identifier)
		lim = l.userLimiter
	case "auth":
		key = fmt.Sprintf("auth:%s", identifier)
		lim = l.authLimiter
	case "data":
		key = fmt.Sprintf("data:%s", identifier)
		lim = l.dataLimiter
	case "sensitive":
		key = fmt.Sprintf("sensitive:%s", identifier)
		lim = l.sensitiveLimiter
	default:
		return fmt.Errorf("unknown limiter type: %s", limiterType)
	}

	_, err := lim.Reset(ctx, key)
	return err
}

// GetConfig returns the current configuration
func (l *Limiter) GetConfig() *Config {
	return l.config
}

// Close closes the Redis connection
func (l *Limiter) Close() error {
	return l.redisClient.Close()
}

// HealthCheck verifies that the rate limiter is working correctly
func (l *Limiter) HealthCheck(ctx context.Context) error {
	// Test Redis connection
	if err := l.redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis connection failed: %w", err)
	}

	// Test rate limiter functionality with a test key
	testKey := fmt.Sprintf("healthcheck:%d", time.Now().Unix())
	_, err := l.CheckIPLimit(ctx, testKey)
	if err != nil {
		return fmt.Errorf("rate limiter check failed: %w", err)
	}

	return nil
}
