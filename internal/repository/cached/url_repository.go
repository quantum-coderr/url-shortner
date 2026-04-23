package cached

import (
	"context"
	"errors"
	"log"

	"url-shortner/internal/cache"
	"url-shortner/internal/domain"
	"url-shortner/internal/repository"
)

// URLRepository decorates a storage repository with read-through caching.
type URLRepository struct {
	storage repository.URLRepository
	cache   cache.URLCache
}

var _ repository.URLRepository = (*URLRepository)(nil)

func NewURLRepository(storage repository.URLRepository, cache cache.URLCache) *URLRepository {
	return &URLRepository{
		storage: storage,
		cache:   cache,
	}
}

func (r *URLRepository) Save(ctx context.Context, shortURL domain.ShortURL) error {
	if err := r.storage.Save(ctx, shortURL); err != nil {
		return err
	}

	if err := r.cache.Set(ctx, shortURL); err != nil {
		log.Printf("cache set after save failed for key=%s: %v", shortURL.Key, err)
	} else {
		log.Printf("cache warm key=%s", shortURL.Key)
	}

	return nil
}

func (r *URLRepository) GetByKey(ctx context.Context, key string) (domain.ShortURL, error) {
	shortURL, err := r.cache.GetByKey(ctx, key)
	if err == nil {
		log.Printf("cache hit key=%s", key)
		return shortURL, nil
	}

	if errors.Is(err, cache.ErrCacheMiss) {
		log.Printf("cache miss key=%s", key)
	} else {
		log.Printf("cache get failed for key=%s, reading from storage: %v", key, err)
	}

	shortURL, err = r.storage.GetByKey(ctx, key)
	if err != nil {
		return domain.ShortURL{}, err
	}
	log.Printf("storage read key=%s", key)

	if err := r.cache.Set(ctx, shortURL); err != nil {
		log.Printf("cache set after storage read failed for key=%s: %v", key, err)
	} else {
		log.Printf("cache repopulated key=%s", key)
	}

	return shortURL, nil
}

func (r *URLRepository) IncrementClicks(ctx context.Context, key string) error {
	if err := r.storage.IncrementClicks(ctx, key); err != nil {
		return err
	}

	if err := r.cache.Delete(ctx, key); err != nil {
		log.Printf("cache delete after increment failed for key=%s: %v", key, err)
	} else {
		log.Printf("cache invalidated key=%s", key)
	}

	return nil
}
