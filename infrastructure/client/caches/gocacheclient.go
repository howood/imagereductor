package caches

import (
	"context"
	"fmt"
	"time"

	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/patrickmn/go-cache"
)

const (
	// NumInstance is number of instance
	NumInstance = 5
	// PurgeExpiredTime is time to purge cache
	PurgeExpiredTime = 10
)

var gocacheConnectionMap map[int]*cache.Cache

func init() {
	gocacheConnectionMap = make(map[int]*cache.Cache, 0)
	for i := 0; i < NumInstance; i++ {
		gocacheConnectionMap[i] = cache.New(60*time.Minute, PurgeExpiredTime*time.Minute)
	}
}

// GoCacheClient struct
type GoCacheClient struct {
	ctx context.Context
}

// NewGoCacheClient creates a new GoCacheClient
func NewGoCacheClient(ctx context.Context) *GoCacheClient {
	ret := &GoCacheClient{ctx: ctx}
	return ret
}

// Get gets from cache
func (cc *GoCacheClient) Get(key string) (interface{}, bool) {
	return cc.getInstance(key).Get(key)
}

// Set puts to cache
func (cc *GoCacheClient) Set(key string, value interface{}, ttl time.Duration) error {
	cc.getInstance(key).Set(key, value, ttl)
	return nil
}

// Del deletes from cache
func (cc *GoCacheClient) Del(key string) error {
	cc.getInstance(key).Delete(key)
	return nil
}

// DelBulk bulk deletes from cache
func (cc *GoCacheClient) DelBulk(key string) error {
	cc.getInstance(key).Delete(key)
	return nil
}

// CloseConnect close connection
func (cc *GoCacheClient) CloseConnect() error {
	return nil
}

func (cc *GoCacheClient) getInstance(key string) *cache.Cache {
	// djb2 algorithm
	i, hash := 0, uint32(5381)
	for _, c := range key {
		hash = ((hash << 5) + hash) + uint32(c)
	}
	i = int(hash) % NumInstance
	log.Info(cc.ctx, fmt.Sprintf("get_instance: %d", i))
	return gocacheConnectionMap[i]
}
