package repository

import (
	"context"

	"url-shortner/internal/domain"
)

// URLRepository hides storage details from service layer.
type URLRepository interface {
	Save(ctx context.Context, shortURL domain.ShortURL) error
	GetByKey(ctx context.Context, key string) (domain.ShortURL, error)
	IncrementClicks(ctx context.Context, key string) error
}
