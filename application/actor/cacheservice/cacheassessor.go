package cacheservice

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/howood/imagereductor/infrastructure/client/caches"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/howood/imagereductor/library/utils"
)

// Sentinel errors for cache validation.
var (
	ErrInvalidCacheType = errors.New("invalid cache type")
	ErrCacheTypeEmpty   = errors.New("CACHE_TYPE environment variable is not set")
)

// CacheAssessor struct.
type CacheAssessor struct {
	instance caches.CacheInstance
}

// NewCacheAssessor creates a new CacheAssessor.
// Deprecated: Use NewCacheAssessorWithConfig for better error handling.
func NewCacheAssessor(db int) *CacheAssessor {
	assessor, err := NewCacheAssessorWithConfig(context.Background(), db)
	if err != nil {
		panic(err)
	}
	return assessor
}

// NewCacheAssessorWithConfig creates a new CacheAssessor with proper error handling.
func NewCacheAssessorWithConfig(ctx context.Context, db int) (*CacheAssessor, error) {
	cacheType := os.Getenv("CACHE_TYPE")
	if cacheType == "" {
		return nil, ErrCacheTypeEmpty
	}
	log.Debug(ctx, "use:"+cacheType)

	switch cacheType {
	case "redis":
		return &CacheAssessor{
			instance: caches.NewRedis(true, db),
		}, nil
	case "gocache":
		return &CacheAssessor{
			instance: caches.NewGoCacheClient(),
		}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidCacheType, cacheType)
	}
}

// Get returns cache contents.
func (ca *CacheAssessor) Get(ctx context.Context, index string) (interface{}, bool, error) {
	defer func() {
		if r := ca.instance.CloseConnect(); r != nil {
			return
		}
	}()
	cachedvalue, cachedfound, err := ca.instance.Get(ctx, index)
	if err != nil {
		log.Error(ctx, "Cache Get Error", err)
		return nil, false, err
	}
	if cachedfound {
		return cachedvalue, true, nil
	}
	return "", false, nil
}

// Set puts cache contents.
func (ca *CacheAssessor) Set(ctx context.Context, index string, value interface{}, expired int) error {
	defer func() {
		if r := ca.instance.CloseConnect(); r != nil {
			return
		}
	}()
	return ca.instance.Set(ctx, index, value, time.Duration(expired)*time.Second)
}

// Delete remove cache contents.
func (ca *CacheAssessor) Delete(ctx context.Context, index string) error {
	defer func() {
		if r := ca.instance.CloseConnect(); r != nil {
			return
		}
	}()
	return ca.instance.Del(ctx, index)
}

// GetChacheExpired get cache expired.
func GetChacheExpired() int {
	//nolint:mnd
	return utils.GetOsEnvInt("CACHEEXPIED", 300)
}

// GetCachedDB get cache db.
func GetCachedDB() int {
	return utils.GetOsEnvInt("CACHEDDB", 0)
}

// GetSessionDB get session db.
func GetSessionDB() int {
	return utils.GetOsEnvInt("SESSIONDB", 1)
}
