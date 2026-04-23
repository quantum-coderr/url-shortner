package memory

import (
	"context"
	"sync"

	"url-shortner/internal/domain"
	"url-shortner/internal/repository"
)

type URLRepository struct {
	mu      sync.RWMutex
	records map[string]domain.ShortURL
}

var _ repository.URLRepository = (*URLRepository)(nil)

func NewURLRepository() *URLRepository {
	return &URLRepository{
		records: make(map[string]domain.ShortURL),
	}
}

func (r *URLRepository) Save(_ context.Context, shortURL domain.ShortURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.records[shortURL.Key]; exists {
		return domain.ErrDuplicateKey
	}

	r.records[shortURL.Key] = shortURL
	return nil
}

func (r *URLRepository) GetByKey(_ context.Context, key string) (domain.ShortURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortURL, exists := r.records[key]
	if !exists {
		return domain.ShortURL{}, domain.ErrNotFound
	}

	return shortURL, nil
}

func (r *URLRepository) IncrementClicks(_ context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	shortURL, exists := r.records[key]
	if !exists {
		return domain.ErrNotFound
	}

	shortURL.Clicks++
	r.records[key] = shortURL
	return nil
}
