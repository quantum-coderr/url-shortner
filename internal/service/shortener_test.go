package service

import (
	"context"
	"errors"
	"testing"

	analyticsMemory "url-shortner/internal/analytics/memory"
	"url-shortner/internal/domain"
	repoMemory "url-shortner/internal/repository/memory"
)

type fixedGenerator struct {
	key string
}

func (g fixedGenerator) NextKey() string {
	return g.key
}

func TestShortenAndResolve(t *testing.T) {
	repo := repoMemory.NewURLRepository()
	svc := NewShortenerService(repo, fixedGenerator{key: "abc123"}, analyticsMemory.NewClickCounter())

	created, err := svc.Shorten(context.Background(), "https://example.com/page")
	if err != nil {
		t.Fatalf("shorten: %v", err)
	}

	if created.Key != "abc123" {
		t.Fatalf("unexpected key: got %q", created.Key)
	}

	found, err := svc.Resolve(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if found.OriginalURL != "https://example.com/page" {
		t.Fatalf("unexpected original url: got %q", found.OriginalURL)
	}
}

func TestShortenRejectsInvalidURL(t *testing.T) {
	repo := repoMemory.NewURLRepository()
	svc := NewShortenerService(repo, fixedGenerator{key: "abc123"}, analyticsMemory.NewClickCounter())

	_, err := svc.Shorten(context.Background(), "not-a-valid-url")
	if !errors.Is(err, domain.ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}
}

func TestShortenDuplicateKey(t *testing.T) {
	repo := repoMemory.NewURLRepository()
	svc := NewShortenerService(repo, fixedGenerator{key: "same"}, analyticsMemory.NewClickCounter())

	_, err := svc.Shorten(context.Background(), "https://example.com/1")
	if err != nil {
		t.Fatalf("first shorten: %v", err)
	}

	_, err = svc.Shorten(context.Background(), "https://example.com/2")
	if !errors.Is(err, domain.ErrDuplicateKey) {
		t.Fatalf("expected ErrDuplicateKey, got %v", err)
	}
}

func TestRegisterClick(t *testing.T) {
	repo := repoMemory.NewURLRepository()
	svc := NewShortenerService(repo, fixedGenerator{key: "click1"}, analyticsMemory.NewClickCounter())

	_, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("shorten: %v", err)
	}

	if err := svc.RegisterClick(context.Background(), "click1"); err != nil {
		t.Fatalf("register click: %v", err)
	}

	clicks, err := svc.Clicks(context.Background(), "click1")
	if err != nil {
		t.Fatalf("clicks: %v", err)
	}

	if clicks != 1 {
		t.Fatalf("expected clicks=1, got %d", clicks)
	}
}
