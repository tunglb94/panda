// Package redis provides a Redis client for FAIRRIDE services.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis connection parameters.
type Config struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

// Client is an alias for redis.Client so callers can use all Redis methods
// without importing go-redis directly.
type Client = redis.Client

// Connect creates and validates a Redis client connection.
// The caller is responsible for closing the client with client.Close().
func Connect(ctx context.Context, cfg Config) (*Client, error) {
	poolSize := cfg.PoolSize
	if poolSize <= 0 {
		poolSize = 10
	}

	minIdle := poolSize / 5
	if minIdle < 1 {
		minIdle = 1
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     poolSize,
		MinIdleConns: minIdle,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping %s: %w", cfg.Addr, err)
	}

	return client, nil
}

// MustConnect is like Connect but panics on error. Use only during service startup.
func MustConnect(ctx context.Context, cfg Config) *Client {
	client, err := Connect(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("fairride/redis: %v", err))
	}
	return client
}
