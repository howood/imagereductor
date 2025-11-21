package caches

import (
	"context"
	"fmt"
	"time"

	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/patrickmn/go-cache"
)

const (
	// NumInstance is number of instance.
	NumInstance = 5
	// DefaultExpiration is default expiration.
	DefaultExpiration = 60
	// PurgeExpiredTime is time to purge cache.
	PurgeExpiredTime = 10
)

//nolint:gochecknoglobals
var gocacheConnectionMap map[int]*cache.Cache

//nolint:gochecknoinits
func init() {
	gocacheConnectionMap = make(map[int]*cache.Cache, 0)
	for i := range NumInstance {
		gocacheConnectionMap[i] = cache.New(DefaultExpiration*time.Minute, PurgeExpiredTime*time.Minute)
	}
}

// GoCacheClient struct.
type GoCacheClient struct{}

// NewGoCacheClient creates a new GoCacheClient.
func NewGoCacheClient() *GoCacheClient {
	ret := &GoCacheClient{}
	return ret
}

// Get gets from cache.
func (cc *GoCacheClient) Get(ctx context.Context, key string) (any, bool, error) {
	val, ok := cc.getInstance(ctx, key).Get(key)
	if !ok {
		return nil, false, nil
	}
	return val, true, nil
}

// Set puts to cache.
func (cc *GoCacheClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	cc.getInstance(ctx, key).Set(key, value, ttl)
	return nil
}

// Del deletes from cache.
func (cc *GoCacheClient) Del(ctx context.Context, key string) error {
	cc.getInstance(ctx, key).Delete(key)
	return nil
}

// DelBulk bulk deletes from cache.
func (cc *GoCacheClient) DelBulk(ctx context.Context, key string) error {
	cc.getInstance(ctx, key).Delete(key)
	return nil
}

// CloseConnect close connection.
func (cc *GoCacheClient) CloseConnect() error {
	return nil
}

//nolint:mnd
func (cc *GoCacheClient) getInstance(ctx context.Context, key string) *cache.Cache {
	// djb2 algorithm
	_, hash := 0, uint32(5381)
	for _, c := range key {
		hash = ((hash << 5) + hash) + uint32(c)
	}
	i := int(hash) % NumInstance
	log.Debug(ctx, fmt.Sprintf("get_instance: %d", i))
	return gocacheConnectionMap[i]
}
