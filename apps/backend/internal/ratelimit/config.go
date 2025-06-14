package ratelimit

import (
	"fmt"
	"time"
)

// Config holds rate limiting configuration
type Config struct {
	// Redis connection settings
	RedisAddr     string `mapstructure:"redis_addr"`
	RedisPassword string `mapstructure:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db"`

	// Default rate limits
	DefaultIPLimit   RateLimit `mapstructure:"default_ip_limit"`
	DefaultUserLimit RateLimit `mapstructure:"default_user_limit"`

	// Endpoint-specific limits
	AuthEndpointLimit      RateLimit `mapstructure:"auth_endpoint_limit"`
	DataEndpointLimit      RateLimit `mapstructure:"data_endpoint_limit"`
	SensitiveEndpointLimit RateLimit `mapstructure:"sensitive_endpoint_limit"`

	// Rate limit headers
	IncludeHeaders bool `mapstructure:"include_headers"`

	// Trusted proxy settings for IP extraction
	TrustedProxies []string `mapstructure:"trusted_proxies"`
}

// RateLimit defines a rate limit configuration
type RateLimit struct {
	Requests int           `mapstructure:"requests"`
	Window   time.Duration `mapstructure:"window"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,

		// Default limits
		DefaultIPLimit: RateLimit{
			Requests: 100,
			Window:   time.Minute,
		},
		DefaultUserLimit: RateLimit{
			Requests: 500,
			Window:   time.Minute,
		},

		// Endpoint-specific limits
		AuthEndpointLimit: RateLimit{
			Requests: 10,
			Window:   time.Minute,
		},
		DataEndpointLimit: RateLimit{
			Requests: 200,
			Window:   time.Minute,
		},
		SensitiveEndpointLimit: RateLimit{
			Requests: 5,
			Window:   time.Minute,
		},

		// Include rate limit headers in responses
		IncludeHeaders: true,

		// Common trusted proxy headers
		TrustedProxies: []string{
			"127.0.0.1",
			"::1",
		},
	}
}

// String returns a string representation of the rate limit
func (rl RateLimit) String() string {
	return fmt.Sprintf("%d requests per %v", rl.Requests, rl.Window)
}
