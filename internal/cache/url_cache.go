package cache

import (
	"context"
	"errors"

	"url-shortner/internal/domain"
)

var ErrCacheMiss = errors.New("cache miss")

// URLCache stores short URLs by key for faster reads.
type URLCache interface {
	GetByKey(ctx context.Context, key string) (domain.ShortURL, error)
	Set(ctx context.Context, shortURL domain.ShortURL) error
	Delete(ctx context.Context, key string) error
}
