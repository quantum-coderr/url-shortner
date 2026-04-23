package memory

import (
	"context"
	"sync"

	"url-shortner/internal/analytics"
)

type ClickCounter struct {
	mu     sync.RWMutex
	counts map[string]uint64
}

var _ analytics.ClickCounter = (*ClickCounter)(nil)

func NewClickCounter() *ClickCounter {
	return &ClickCounter{
		counts: make(map[string]uint64),
	}
}

func (c *ClickCounter) Increment(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.counts[key]++
	return nil
}

func (c *ClickCounter) Get(_ context.Context, key string) (uint64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.counts[key], nil
}
