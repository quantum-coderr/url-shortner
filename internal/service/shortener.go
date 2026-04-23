package service

import (
	"context"
	"net/url"
	"time"

	"url-shortner/internal/analytics"
	"url-shortner/internal/domain"
	"url-shortner/internal/generator"
	"url-shortner/internal/repository"
)

type ShortenerService struct {
	repo         repository.URLRepository
	generator    generator.KeyGenerator
	clickCounter analytics.ClickCounter
}

func NewShortenerService(
	repo repository.URLRepository,
	generator generator.KeyGenerator,
	clickCounter analytics.ClickCounter,
) *ShortenerService {
	return &ShortenerService{
		repo:         repo,
		generator:    generator,
		clickCounter: clickCounter,
	}
}

func (s *ShortenerService) Shorten(ctx context.Context, rawURL string) (domain.ShortURL, error) {
	if !isValidURL(rawURL) {
		return domain.ShortURL{}, domain.ErrInvalidURL
	}

	shortURL := domain.ShortURL{
		Key:         s.generator.NextKey(),
		OriginalURL: rawURL,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.repo.Save(ctx, shortURL); err != nil {
		return domain.ShortURL{}, err
	}

	return shortURL, nil
}

func (s *ShortenerService) Resolve(ctx context.Context, key string) (domain.ShortURL, error) {
	return s.repo.GetByKey(ctx, key)
}

func (s *ShortenerService) RegisterClick(ctx context.Context, key string) error {
	return s.clickCounter.Increment(ctx, key)
}

func (s *ShortenerService) Clicks(ctx context.Context, key string) (uint64, error) {
	if _, err := s.repo.GetByKey(ctx, key); err != nil {
		return 0, err
	}

	return s.clickCounter.Get(ctx, key)
}

func isValidURL(rawURL string) bool {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}

	switch parsed.Scheme {
	case "http", "https":
		return parsed.Host != ""
	default:
		return false
	}
}
