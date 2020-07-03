package actor

import (
	"os"
	"strconv"
	"time"

	"github.com/howood/imagereductor/infrastructure/client"
	log "github.com/howood/imagereductor/infrastructure/logger"
)

func SetToRedis(index string, data interface{}, expired time.Duration, connectionpersistent bool, redisdb int) error {
	cached := client.NewRedis(connectionpersistent, redisdb)
	defer cached.CloseConnect()
	return cached.Set(index, data, expired*time.Minute)
}

func SetToRedisExternal(index string, data interface{}, connectionpersistent bool, redisdb int) error {
	log.Debug(redisdb)
	cached := client.NewRedis(connectionpersistent, redisdb)
	defer cached.CloseConnect()
	return cached.Set(index, data, 0)
}

func GetFromRedis(index string, connectionpersistent bool, redisdb int) (interface{}, bool) {
	cached := client.NewRedis(connectionpersistent, redisdb)
	cachedvalue, cachedfound := cached.Get(index)
	defer cached.CloseConnect()
	if cachedfound {
		return cachedvalue, true
	} else {
		return "", false
	}
}

func DeleteToRedis(index string, connectionpersistent bool, redisdb int) error {
	cached := client.NewRedis(connectionpersistent, redisdb)
	defer cached.CloseConnect()
	return cached.Del(index)
}

func DeleteBulkToRedis(index string, connectionpersistent bool, redisdb int) error {
	cached := client.NewRedis(connectionpersistent, redisdb)
	defer cached.CloseConnect()
	return cached.DelBulk(index)
}

func GetChacheExpired() time.Duration {
	expired, err := strconv.Atoi(os.Getenv("REDISCACHEEXPIED"))
	if err != nil {
		panic(err)
	}
	return time.Duration(expired)
}

func GetCachedDB() int {
	db, err := strconv.Atoi(os.Getenv("REDISCACHEDDB"))
	if err != nil {
		panic(err)
	}
	return db
}
