package cacheservice

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/howood/imagereductor/infrastructure/client/caches"
	log "github.com/howood/imagereductor/infrastructure/logger"
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
	defer ca.instance.CloseConnect()
	cachedvalue, cachedfound := ca.instance.Get(index)
	if cachedfound {
		return cachedvalue, true
	} else {
		return "", false
	}
}

// Set puts cache contents
func (ca *CacheAssessor) Set(index string, value interface{}, expired time.Duration) error {
	defer ca.instance.CloseConnect()
	return ca.instance.Set(index, value, expired*time.Second)
}

// Delete remove cache contents
func (ca *CacheAssessor) Delete(index string) error {
	defer ca.instance.CloseConnect()
	return ca.instance.Del(index)
}

func GetChacheExpired() time.Duration {
	expired, err := strconv.Atoi(os.Getenv("CACHEEXPIED"))
	if err != nil {
		panic(err)
	}
	return time.Duration(expired)
}

func GetCachedDB() int {
	db, err := strconv.Atoi(os.Getenv("CACHEDDB"))
	if err != nil {
		panic(err)
	}
	return db
}

func GetSessionDB() int {
	db, err := strconv.Atoi(os.Getenv("SESSIONDB"))
	if err != nil {
		panic(err)
	}
	return db
}
