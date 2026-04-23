package cache

import (
	"context"

	"url-shortner/internal/domain"
)

type NoopURLCache struct{}

var _ URLCache = (*NoopURLCache)(nil)

func NewNoopURLCache() *NoopURLCache {
	return &NoopURLCache{}
}

func (c *NoopURLCache) GetByKey(_ context.Context, _ string) (domain.ShortURL, error) {
	return domain.ShortURL{}, ErrCacheMiss
}

func (c *NoopURLCache) Set(_ context.Context, _ domain.ShortURL) error {
	return nil
}

func (c *NoopURLCache) Delete(_ context.Context, _ string) error {
	return nil
}
