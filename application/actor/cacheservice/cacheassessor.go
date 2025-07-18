package cacheservice

import (
	"context"
	"os"
	"time"

	"github.com/howood/imagereductor/infrastructure/client/caches"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/howood/imagereductor/library/utils"
)

// CacheAssessor struct.
type CacheAssessor struct {
	instance caches.CacheInstance
}

// NewCacheAssessor creates a new CacheAssessor.
func NewCacheAssessor(db int) *CacheAssessor {
	var I *CacheAssessor
	log.Debug(context.Background(), "use:"+os.Getenv("CACHE_TYPE"))
	switch os.Getenv("CACHE_TYPE") {
	case "redis":
		I = &CacheAssessor{
			instance: caches.NewRedis(true, db),
		}
	case "gocache":
		I = &CacheAssessor{
			instance: caches.NewGoCacheClient(),
		}
	default:
		panic("Invalid CACHE_TYPE")
	}
	return I
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
