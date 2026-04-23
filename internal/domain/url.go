package domain

import "time"

// ShortURL is the core entity used across layers.
type ShortURL struct {
	Key         string
	OriginalURL string
	CreatedAt   time.Time
	Clicks      uint64
}
