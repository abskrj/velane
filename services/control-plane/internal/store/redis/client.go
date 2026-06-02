// Package redisstore provides Redis-backed job queue and warm pool primitives.
package redisstore

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
)

// Client wraps a Redis connection and exposes queue and warm-pool operations.
type Client struct {
	rdb *redis.Client
}

func optionsFor(addr string) (*redis.Options, error) {
	if strings.Contains(addr, "://") {
		return redis.ParseURL(addr)
	}
	return &redis.Options{Addr: addr}, nil
}

// New creates a Redis client, pings the server, and returns a ready Client.
func New(addr string) (*Client, error) {
	opts, err := optionsFor(addr)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &Client{rdb: rdb}, nil
}

// Close releases the underlying Redis connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}
