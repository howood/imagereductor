package caches

import (
	"fmt"
	"time"

	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/patrickmn/go-cache"
)

const (
	NUM_INSTANCE       = 5
	PURGE_EXPIRED_TIME = 10
)

var gocacheConnectionMap map[int]*cache.Cache

func init() {
	gocacheConnectionMap = make(map[int]*cache.Cache, 0)
	for i := 0; i < NUM_INSTANCE; i++ {
		gocacheConnectionMap[i] = cache.New(60*time.Minute, PURGE_EXPIRED_TIME*time.Minute)
	}
}

type GoCacheClient struct {
}

// インスタンス作成用のメソッド
func NewGoCacheClient() *GoCacheClient {
	ret := &GoCacheClient{}
	return ret
}

func (cc *GoCacheClient) Get(key string) (interface{}, bool) {
	return cc.getInstance(key).Get(key)
}

func (cc *GoCacheClient) Set(key string, value interface{}, ttl time.Duration) error {
	cc.getInstance(key).Set(key, value, ttl)
	return nil
}

func (cc *GoCacheClient) Del(key string) error {
	cc.getInstance(key).Delete(key)
	return nil
}

func (cc *GoCacheClient) DelBulk(key string) error {
	cc.getInstance(key).Delete(key)
	return nil
}

func (cc *GoCacheClient) CloseConnect() error {
	return nil
}

func (cc *GoCacheClient) getInstance(key string) *cache.Cache {
	// djb2アルゴリズム
	i, hash := 0, uint32(5381)
	for _, c := range key {
		hash = ((hash << 5) + hash) + uint32(c)
	}
	i = int(hash) % NUM_INSTANCE
	log.Info("", fmt.Sprintf("get_instance: %d", i))
	return gocacheConnectionMap[i]
}
