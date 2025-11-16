package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient configures a Redis client and verifies the connection with a ping.
func NewRedisClient(url string) (*redis.Client, error) {
	options, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(options)

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
