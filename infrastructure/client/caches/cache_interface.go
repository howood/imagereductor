package caches

import (
	"context"
	"time"
)

// CacheInstance interface.
type CacheInstance interface {
	Set(ctx context.Context, key string, value interface{}, expired time.Duration) error
	Get(ctx context.Context, key string) (interface{}, bool, error)
	Del(ctx context.Context, key string) error
	DelBulk(ctx context.Context, key string) error
	CloseConnect() error
}
