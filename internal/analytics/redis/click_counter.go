package redis

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"url-shortner/internal/analytics"

	goredis "github.com/redis/go-redis/v9"
)

type ClickCounter struct {
	client *goredis.Client
	prefix string
}

var _ analytics.ClickCounter = (*ClickCounter)(nil)

func NewClickCounter(client *goredis.Client) *ClickCounter {
	return &ClickCounter{
		client: client,
		prefix: "short-url:clicks:",
	}
}

func (c *ClickCounter) Increment(ctx context.Context, key string) error {
	if err := c.client.Incr(ctx, c.redisKey(key)).Err(); err != nil {
		return fmt.Errorf("redis incr: %w", err)
	}
	log.Printf("analytics increment key=%s (redis)", key)
	return nil
}

func (c *ClickCounter) Get(ctx context.Context, key string) (uint64, error) {
	value, err := c.client.Get(ctx, c.redisKey(key)).Result()
	if err != nil {
		if err == goredis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("redis get: %w", err)
	}

	clicks, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse clicks: %w", err)
	}
	log.Printf("analytics read key=%s clicks=%d (redis)", key, clicks)
	return clicks, nil
}

func (c *ClickCounter) redisKey(key string) string {
	return c.prefix + key
}
