package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"real-time-leaderboard/internal/config"
	"real-time-leaderboard/internal/shared/logger"
)

// Client represents a Redis client
type Client struct {
	client *redis.Client
	logger *logger.Logger
}

// NewClient creates a new Redis client
func NewClient(cfg config.RedisConfig, l *logger.Logger) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.GetAddr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	l.Info("Redis connection established")

	return &Client{
		client: rdb,
		logger: l,
	}, nil
}

// GetClient returns the underlying Redis client
func (c *Client) GetClient() *redis.Client {
	return c.client
}

// Close closes the Redis connection
func (c *Client) Close() error {
	if c.client != nil {
		err := c.client.Close()
		if err == nil {
			c.logger.Info("Redis connection closed")
		}
		return err
	}
	return nil
}

// Health checks the health of the Redis connection
func (c *Client) Health(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

