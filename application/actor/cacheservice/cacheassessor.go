package cacheservice

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/howood/imagereductor/infrastructure/client/caches"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/howood/imagereductor/library/utils"
)

// CacheAssessor struct
type CacheAssessor struct {
	instance caches.CacheInstance
	ctx      context.Context
}

// NewCacheAssessor creates a new CacheAssessor
func NewCacheAssessor(ctx context.Context, db int) *CacheAssessor {
	var I *CacheAssessor
	log.Debug(ctx, "use:"+os.Getenv("CACHE_TYPE"))
	switch os.Getenv("CACHE_TYPE") {
	case "redis":
		I = &CacheAssessor{
			instance: caches.NewRedis(ctx, true, db),
			ctx:      ctx,
		}
	case "gocache":
		I = &CacheAssessor{
			instance: caches.NewGoCacheClient(ctx),
			ctx:      ctx,
		}
	default:
		panic(errors.New("Invalid CACHE_TYPE"))
	}
	return I
}

// Get returns cache contents
func (ca *CacheAssessor) Get(index string) (interface{}, bool) {
	defer func() {
		if r := ca.instance.CloseConnect(); r != nil {
			return
		}
	}()
	cachedvalue, cachedfound := ca.instance.Get(index)
	if cachedfound {
		return cachedvalue, true
	}
	return "", false
}

// Set puts cache contents
func (ca *CacheAssessor) Set(index string, value interface{}, expired time.Duration) error {
	defer func() {
		if r := ca.instance.CloseConnect(); r != nil {
			return
		}
	}()
	return ca.instance.Set(index, value, expired*time.Second)
}

// Delete remove cache contents
func (ca *CacheAssessor) Delete(index string) error {
	defer func() {
		if r := ca.instance.CloseConnect(); r != nil {
			return
		}
	}()
	return ca.instance.Del(index)
}

// GetChacheExpired get cache expired
func GetChacheExpired() time.Duration {
	return time.Duration(utils.GetOsEnvInt("CACHEEXPIED", 300))
}

// GetCachedDB get cache db
func GetCachedDB() int {
	return utils.GetOsEnvInt("CACHEDDB", 0)
}

// GetSessionDB get session db
func GetSessionDB() int {
	return utils.GetOsEnvInt("SESSIONDB", 1)
}
