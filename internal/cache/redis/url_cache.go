package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"url-shortner/internal/cache"
	"url-shortner/internal/domain"
)

type URLCache struct {
	client *goredis.Client
	ttl    time.Duration
	prefix string
}

var _ cache.URLCache = (*URLCache)(nil)

func NewURLCache(client *goredis.Client, ttl time.Duration) *URLCache {
	return &URLCache{
		client: client,
		ttl:    ttl,
		prefix: "short-url:",
	}
}

func (c *URLCache) GetByKey(ctx context.Context, key string) (domain.ShortURL, error) {
	value, err := c.client.Get(ctx, c.redisKey(key)).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return domain.ShortURL{}, cache.ErrCacheMiss
		}
		return domain.ShortURL{}, fmt.Errorf("redis get: %w", err)
	}

	var shortURL domain.ShortURL
	if err := json.Unmarshal([]byte(value), &shortURL); err != nil {
		return domain.ShortURL{}, fmt.Errorf("unmarshal cached short url: %w", err)
	}

	return shortURL, nil
}

func (c *URLCache) Set(ctx context.Context, shortURL domain.ShortURL) error {
	payload, err := json.Marshal(shortURL)
	if err != nil {
		return fmt.Errorf("marshal short url: %w", err)
	}

	if err := c.client.Set(ctx, c.redisKey(shortURL.Key), payload, c.ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

func (c *URLCache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, c.redisKey(key)).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}

	return nil
}

func (c *URLCache) redisKey(key string) string {
	return c.prefix + key
}
