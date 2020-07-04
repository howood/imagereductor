package caches

import (
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/go-redis/redis"
	log "github.com/howood/imagereductor/infrastructure/logger"
)

const (
	REDIS_MAX_RETRY          = 3
	REDIS_CONNECTION_RANDMAX = 10000
)

var redisConnectionMap map[int]*redis.Client

type RedisInstance struct {
	ConnectionPersistent bool
	client               *redis.Client
	redisdb              int
	connectionkey        int
}

func init() {
	redisConnectionMap = make(map[int]*redis.Client, 0)
}

// インスタンス作成用のメソッド
func NewRedis(connectionpersistent bool, redisdb int) *RedisInstance {

	log.Debug("----DNS----")
	log.Debug(os.Getenv("REDISHOST") + ":" + fmt.Sprint(os.Getenv("REDISPORT")))
	log.Debug(os.Getenv("REDISPASSWORD"))
	log.Debug(redisdb)
	log.Debug(redisConnectionMap)
	var connectionkey int
	if connectionpersistent == true {
		connectionkey = redisdb
	} else {
		rand.Seed(time.Now().UnixNano())
		connectionkey = rand.Intn(REDIS_CONNECTION_RANDMAX)
	}
	if redisConnectionMap[connectionkey] == nil || checkConnect(connectionkey) == false {
		log.Info("--- Create Redis Connection ---  ")
		if err := createNewConnect(redisdb, connectionkey); err != nil {
			panic(err)
		}
	}
	I := &RedisInstance{
		ConnectionPersistent: connectionpersistent,
		client:               redisConnectionMap[connectionkey],
		redisdb:              redisdb,
		connectionkey:        connectionkey,
	}

	//	defer I.client.Close()
	return I
}
func (i RedisInstance) Set(key string, value interface{}, expired time.Duration) error {
	log.Debug("-----SET----")
	log.Debug(key)
	log.Debug(expired)
	return i.client.Set(key, value, expired).Err()
}

func (i RedisInstance) Get(key string) (interface{}, bool) {
	cachedvalue, err := i.client.Get(key).Result()
	log.Debug("-----GET----")
	log.Debug(key)
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		return nil, false
	} else {
		return cachedvalue, true
	}
}

func (i RedisInstance) Del(key string) error {
	log.Debug("-----DEL----")
	log.Debug(key)
	return i.client.Del(key).Err()
}

func (i RedisInstance) DelBulk(key string) error {
	log.Debug("-----DelBulk----")
	log.Debug(key)
	targetkeys := i.client.Keys(key)
	log.Debug(targetkeys.Val())
	for _, key := range targetkeys.Val() {
		if err := i.client.Del(key).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (i RedisInstance) CloseConnect() error {
	if i.ConnectionPersistent == false {
		err := i.client.Close()
		delete(redisConnectionMap, i.connectionkey)
		return err
	}
	return nil
}

func checkConnect(connectionkey int) bool {
	if err := checkPing(connectionkey); err != nil {
		log.Error(err)
		return false
	}
	return true
}

func checkPing(connectionkey int) error {
	if _, err := redisConnectionMap[connectionkey].Ping().Result(); err != nil {
		return errors.New(fmt.Sprintf("did not connect: %v", err))
	}
	return nil
}

func createNewConnect(redisdb int, connectionkey int) error {
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
		MaxRetries: REDIS_MAX_RETRY,
		TLSConfig:  tlsConfig,
	})
	return checkPing(connectionkey)
}
