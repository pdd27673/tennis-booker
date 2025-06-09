package database

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClient represents the Redis client connection
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient creates a new Redis client
func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ping Redis
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Redis connected: %s\n", pong)

	return &RedisClient{
		Client: client,
	}, nil
}

// Close disconnects from Redis
func (r *RedisClient) Close() error {
	return r.Client.Close()
}

// PublishMessage publishes a message to a Redis channel
func (r *RedisClient) PublishMessage(ctx context.Context, channel, message string) error {
	return r.Client.Publish(ctx, channel, message).Err()
}

// SetKey sets a key-value pair in Redis with optional expiration
func (r *RedisClient) SetKey(ctx context.Context, key, value string, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// GetKey retrieves a value from Redis by key
func (r *RedisClient) GetKey(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
} 