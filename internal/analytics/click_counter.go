package analytics

import "context"

// ClickCounter tracks click counts by short key.
type ClickCounter interface {
	Increment(ctx context.Context, key string) error
	Get(ctx context.Context, key string) (uint64, error)
}
