package caches

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"os"
	"time"

	log "github.com/howood/imagereductor/infrastructure/logger"
	redis "github.com/redis/go-redis/v9"
)

const (
	// RedisMaxRetry is max retry count
	RedisMaxRetry = 3
	// RedisConnectionRandmax is using generate connection key
	RedisConnectionRandmax = 10000
)

var redisConnectionMap map[int]*redis.Client

// RedisInstance struct
type RedisInstance struct {
	ConnectionPersistent bool
	client               *redis.Client
	redisdb              int
	connectionkey        int
	ctx                  context.Context
}

func init() {
	redisConnectionMap = make(map[int]*redis.Client, 0)
}

// NewRedis creates a new RedisInstance
func NewRedis(ctx context.Context, connectionpersistent bool, redisdb int) *RedisInstance {

	log.Debug(ctx, "----DNS----")
	log.Debug(ctx, os.Getenv("REDISHOST")+":"+fmt.Sprint(os.Getenv("REDISPORT")))
	log.Debug(ctx, os.Getenv("REDISPASSWORD"))
	log.Debug(ctx, redisdb)
	log.Debug(ctx, redisConnectionMap)
	var connectionkey int
	if connectionpersistent == true {
		connectionkey = redisdb
	} else {
		rand.Seed(time.Now().UnixNano())
		connectionkey = rand.Intn(RedisConnectionRandmax)
	}
	if redisConnectionMap[connectionkey] == nil || checkConnect(ctx, connectionkey) == false {
		log.Info(ctx, "--- Create Redis Connection ---  ")
		if err := createNewConnect(ctx, redisdb, connectionkey); err != nil {
			panic(err)
		}
	}
	I := &RedisInstance{
		ConnectionPersistent: connectionpersistent,
		client:               redisConnectionMap[connectionkey],
		redisdb:              redisdb,
		connectionkey:        connectionkey,
		ctx:                  ctx,
	}

	//	defer I.client.Close()
	return I
}

// Set puts to cache
func (i *RedisInstance) Set(key string, value interface{}, expired time.Duration) error {
	log.Debug(i.ctx, "-----SET----")
	log.Debug(i.ctx, key)
	log.Debug(i.ctx, expired)
	return i.client.Set(i.ctx, key, value, expired).Err()
}

// Get gets from cache
func (i *RedisInstance) Get(key string) (interface{}, bool) {
	cachedvalue, err := i.client.Get(i.ctx, key).Result()
	log.Debug(i.ctx, "-----GET----")
	log.Debug(i.ctx, key)
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		return nil, false
	} else {
		return cachedvalue, true
	}
}

// Del deletes from cache
func (i *RedisInstance) Del(key string) error {
	log.Debug(i.ctx, "-----DEL----")
	log.Debug(i.ctx, key)
	return i.client.Del(i.ctx, key).Err()
}

// DelBulk bulk deletes from cache
func (i *RedisInstance) DelBulk(key string) error {
	log.Debug(i.ctx, "-----DelBulk----")
	log.Debug(i.ctx, key)
	targetkeys := i.client.Keys(i.ctx, key)
	log.Debug(i.ctx, targetkeys.Val())
	for _, key := range targetkeys.Val() {
		if err := i.client.Del(i.ctx, key).Err(); err != nil {
			return err
		}
	}
	return nil
}

// CloseConnect close connection
func (i *RedisInstance) CloseConnect() error {
	if i.ConnectionPersistent == false {
		err := i.client.Close()
		delete(redisConnectionMap, i.connectionkey)
		return err
	}
	return nil
}

func checkConnect(ctx context.Context, connectionkey int) bool {
	if err := checkPing(ctx, connectionkey); err != nil {
		log.Error(ctx, err)
		return false
	}
	return true
}

func checkPing(ctx context.Context, connectionkey int) error {
	if _, err := redisConnectionMap[connectionkey].Ping(ctx).Result(); err != nil {
		return fmt.Errorf("did not connect: %v", err)
	}
	return nil
}

func createNewConnect(ctx context.Context, redisdb int, connectionkey int) error {
	var tlsConfig *tls.Config
	if os.Getenv("REDISTLS") == "use" {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	redisConnectionMap[connectionkey] = redis.NewClient(&redis.Options{
		Addr:       os.Getenv("REDISHOST") + ":" + fmt.Sprint(os.Getenv("REDISPORT")),
		Password:   os.Getenv("REDISPASSWORD"),
		DB:         redisdb,
		MaxRetries: RedisMaxRetry,
		TLSConfig:  tlsConfig,
	})
	return checkPing(ctx, connectionkey)
}
