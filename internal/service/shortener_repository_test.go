package service

import (
	"context"
	"testing"

	"url-shortner/internal/analytics"
	"url-shortner/internal/domain"
	"url-shortner/internal/repository"
)

type spyRepository struct {
	data           map[string]domain.ShortURL
	saveCalled     bool
	getByKeyCalled bool
}

var _ repository.URLRepository = (*spyRepository)(nil)

func newSpyRepository() *spyRepository {
	return &spyRepository{
		data: make(map[string]domain.ShortURL),
	}
}

func (r *spyRepository) Save(_ context.Context, shortURL domain.ShortURL) error {
	r.saveCalled = true
	r.data[shortURL.Key] = shortURL
	return nil
}

func (r *spyRepository) GetByKey(_ context.Context, key string) (domain.ShortURL, error) {
	r.getByKeyCalled = true
	shortURL, ok := r.data[key]
	if !ok {
		return domain.ShortURL{}, domain.ErrNotFound
	}
	return shortURL, nil
}

func (r *spyRepository) IncrementClicks(_ context.Context, key string) error {
	shortURL, ok := r.data[key]
	if !ok {
		return domain.ErrNotFound
	}

	shortURL.Clicks++
	r.data[key] = shortURL
	return nil
}

type spyClickCounter struct {
	counts          map[string]uint64
	incrementCalled bool
	getCalled       bool
}

var _ analytics.ClickCounter = (*spyClickCounter)(nil)

func newSpyClickCounter() *spyClickCounter {
	return &spyClickCounter{
		counts: make(map[string]uint64),
	}
}

func (c *spyClickCounter) Increment(_ context.Context, key string) error {
	c.incrementCalled = true
	c.counts[key]++
	return nil
}

func (c *spyClickCounter) Get(_ context.Context, key string) (uint64, error) {
	c.getCalled = true
	return c.counts[key], nil
}

func TestShortenerServiceUsesRepositoryAbstraction(t *testing.T) {
	repo := newSpyRepository()
	counter := newSpyClickCounter()
	svc := NewShortenerService(repo, fixedGenerator{key: "repo1"}, counter)

	created, err := svc.Shorten(context.Background(), "https://example.com/repo")
	if err != nil {
		t.Fatalf("shorten: %v", err)
	}

	if !repo.saveCalled {
		t.Fatalf("expected Save to be called")
	}

	_, err = svc.Resolve(context.Background(), created.Key)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if !repo.getByKeyCalled {
		t.Fatalf("expected GetByKey to be called")
	}

	if err := svc.RegisterClick(context.Background(), created.Key); err != nil {
		t.Fatalf("register click: %v", err)
	}

	if !counter.incrementCalled {
		t.Fatalf("expected analytics Increment to be called")
	}

	clicks, err := svc.Clicks(context.Background(), created.Key)
	if err != nil {
		t.Fatalf("clicks: %v", err)
	}
	if clicks != 1 {
		t.Fatalf("expected clicks=1, got %d", clicks)
	}
	if !counter.getCalled {
		t.Fatalf("expected analytics Get to be called")
	}
}
