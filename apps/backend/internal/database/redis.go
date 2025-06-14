package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis client with additional functionality
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client connection
func NewRedisClient(host, port, password string, db int) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       db,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	// Successfully connected to Redis

	return &RedisClient{
		client: rdb,
	}, nil
}

// GetClient returns the underlying Redis client
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Ping tests the Redis connection
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Set sets a key-value pair with optional expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// PublishMessage publishes a message to a Redis channel
func (r *RedisClient) PublishMessage(ctx context.Context, channel, message string) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// SetKey sets a key-value pair in Redis with optional expiration
func (r *RedisClient) SetKey(ctx context.Context, key, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// GetKey retrieves a value from Redis by key
func (r *RedisClient) GetKey(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}
