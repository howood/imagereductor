package caches

import "time"

// CacheInstance interface
type CacheInstance interface {
	Set(key string, value interface{}, expired time.Duration) error
	Get(key string) (interface{}, bool)
	Del(key string) error
	DelBulk(key string) error
	CloseConnect() error
}
