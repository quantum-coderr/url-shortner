package cached

import (
	"context"
	"errors"
	"testing"
	"time"

	"url-shortner/internal/cache"
	"url-shortner/internal/domain"
)

type spyStorage struct {
	data       map[string]domain.ShortURL
	getByKeyN  int
	saveN      int
	incrementN int
}

func newSpyStorage() *spyStorage {
	return &spyStorage{
		data: make(map[string]domain.ShortURL),
	}
}

func (s *spyStorage) Save(_ context.Context, shortURL domain.ShortURL) error {
	s.saveN++
	if _, exists := s.data[shortURL.Key]; exists {
		return domain.ErrDuplicateKey
	}
	s.data[shortURL.Key] = shortURL
	return nil
}

func (s *spyStorage) GetByKey(_ context.Context, key string) (domain.ShortURL, error) {
	s.getByKeyN++
	shortURL, ok := s.data[key]
	if !ok {
		return domain.ShortURL{}, domain.ErrNotFound
	}
	return shortURL, nil
}

func (s *spyStorage) IncrementClicks(_ context.Context, key string) error {
	s.incrementN++
	shortURL, ok := s.data[key]
	if !ok {
		return domain.ErrNotFound
	}
	shortURL.Clicks++
	s.data[key] = shortURL
	return nil
}

type mapCache struct {
	data map[string]domain.ShortURL
}

func newMapCache() *mapCache {
	return &mapCache{
		data: make(map[string]domain.ShortURL),
	}
}

func (c *mapCache) GetByKey(_ context.Context, key string) (domain.ShortURL, error) {
	shortURL, ok := c.data[key]
	if !ok {
		return domain.ShortURL{}, cache.ErrCacheMiss
	}
	return shortURL, nil
}

func (c *mapCache) Set(_ context.Context, shortURL domain.ShortURL) error {
	c.data[shortURL.Key] = shortURL
	return nil
}

func (c *mapCache) Delete(_ context.Context, key string) error {
	delete(c.data, key)
	return nil
}

func TestGetByKeyReadThroughCache(t *testing.T) {
	storage := newSpyStorage()
	storage.data["abc"] = domain.ShortURL{
		Key:         "abc",
		OriginalURL: "https://example.com",
		CreatedAt:   time.Now().UTC(),
	}
	cacheStore := newMapCache()
	repo := NewURLRepository(storage, cacheStore)

	first, err := repo.GetByKey(context.Background(), "abc")
	if err != nil {
		t.Fatalf("first get: %v", err)
	}
	if first.Key != "abc" {
		t.Fatalf("expected key abc, got %q", first.Key)
	}
	if storage.getByKeyN != 1 {
		t.Fatalf("expected storage reads=1, got %d", storage.getByKeyN)
	}

	second, err := repo.GetByKey(context.Background(), "abc")
	if err != nil {
		t.Fatalf("second get: %v", err)
	}
	if second.Key != "abc" {
		t.Fatalf("expected key abc, got %q", second.Key)
	}
	if storage.getByKeyN != 1 {
		t.Fatalf("expected storage reads to remain 1, got %d", storage.getByKeyN)
	}
}

func TestIncrementClicksInvalidatesCache(t *testing.T) {
	storage := newSpyStorage()
	cacheStore := newMapCache()
	repo := NewURLRepository(storage, cacheStore)

	err := repo.Save(context.Background(), domain.ShortURL{
		Key:         "k1",
		OriginalURL: "https://example.com/p",
		CreatedAt:   time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	if err := repo.IncrementClicks(context.Background(), "k1"); err != nil {
		t.Fatalf("increment: %v", err)
	}

	_, err = cacheStore.GetByKey(context.Background(), "k1")
	if !errors.Is(err, cache.ErrCacheMiss) {
		t.Fatalf("expected cache miss after increment, got %v", err)
	}
}
