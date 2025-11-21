package caches

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	log "github.com/howood/imagereductor/infrastructure/logger"
	redis "github.com/redis/go-redis/v9"
)

const (
	// RedisMaxRetry is max retry count.
	RedisMaxRetry = 3
	// RedisConnectionRandmax is using generate connection key.
	RedisConnectionRandmax = 10000
)

//nolint:gochecknoglobals
var redisConnectionMap map[int]*redis.Client

// RedisInstance struct.
type RedisInstance struct {
	ConnectionPersistent bool
	client               *redis.Client
	redisdb              int
	connectionkey        int
}

//nolint:gochecknoinits
func init() {
	redisConnectionMap = make(map[int]*redis.Client, 0)
}

// NewRedis creates a new RedisInstance.
func NewRedis(connectionpersistent bool, redisdb int) *RedisInstance {
	ctx := context.Background()
	log.Debug(ctx, "----DNS----")
	log.Debug(ctx, os.Getenv("REDISHOST")+":"+os.Getenv("REDISPORT"))
	log.Debug(ctx, os.Getenv("REDISPASSWORD"))
	log.Debug(ctx, redisdb)
	log.Debug(ctx, redisConnectionMap)
	var connectionkey int
	if connectionpersistent {
		connectionkey = redisdb
	} else {
		n, err := rand.Int(rand.Reader, big.NewInt(RedisConnectionRandmax))
		if err != nil {
			panic(err)
		}
		connectionkey = int(n.Int64())
	}
	if redisConnectionMap[connectionkey] == nil || !checkConnect(ctx, connectionkey) {
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
	}

	//	defer I.client.Close()
	return I
}

// Set puts to cache.
func (i *RedisInstance) Set(ctx context.Context, key string, value any, expired time.Duration) error {
	log.Debug(ctx, "-----SET----")
	log.Debug(ctx, key)
	log.Debug(ctx, expired)
	return i.client.Set(ctx, key, value, expired).Err()
}

// Get gets from cache.
func (i *RedisInstance) Get(ctx context.Context, key string) (any, bool, error) {
	cachedvalue, err := i.client.Get(ctx, key).Result()
	log.Debug(ctx, "-----GET----")
	log.Debug(ctx, key)
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return cachedvalue, true, nil
}

// Del deletes from cache.
func (i *RedisInstance) Del(ctx context.Context, key string) error {
	log.Debug(ctx, "-----DEL----")
	log.Debug(ctx, key)
	return i.client.Del(ctx, key).Err()
}

// DelBulk bulk deletes from cache.
func (i *RedisInstance) DelBulk(ctx context.Context, key string) error {
	log.Debug(ctx, "-----DelBulk----")
	log.Debug(ctx, key)
	targetkeys := i.client.Keys(ctx, key)
	log.Debug(ctx, targetkeys.Val())
	for _, key := range targetkeys.Val() {
		if err := i.client.Del(ctx, key).Err(); err != nil {
			return err
		}
	}
	return nil
}

// CloseConnect close connection.
func (i *RedisInstance) CloseConnect() error {
	if !i.ConnectionPersistent {
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
		return fmt.Errorf("did not connect: %w", err)
	}
	return nil
}

func createNewConnect(ctx context.Context, redisdb int, connectionkey int) error {
	var tlsConfig *tls.Config
	redistls := os.Getenv("REDISTLS")
	switch redistls {
	case "enable":
		// Production mode: verify server certificate
		tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		// Optional: custom CA certificate
		if caCertPath := os.Getenv("REDISTLS_CA_CERT"); caCertPath != "" {
			caCert, err := os.ReadFile(caCertPath)
			if err != nil {
				log.Warn(ctx, "Failed to read CA certificate: "+err.Error())
			} else {
				caCertPool := x509.NewCertPool()
				if caCertPool.AppendCertsFromPEM(caCert) {
					tlsConfig.RootCAs = caCertPool
				}
			}
		}
		// Optional: SNI server name (required for ElastiCache)
		if serverName := os.Getenv("REDISTLS_SERVER_NAME"); serverName != "" {
			tlsConfig.ServerName = serverName
		}
	case "skipverify":
		// Development mode: skip certificate verification
		tlsConfig = &tls.Config{
			//nolint:gosec
			InsecureSkipVerify: true,
		}
	}
	redisConnectionMap[connectionkey] = redis.NewClient(&redis.Options{
		Addr:       os.Getenv("REDISHOST") + ":" + os.Getenv("REDISPORT"),
		Password:   os.Getenv("REDISPASSWORD"),
		DB:         redisdb,
		MaxRetries: RedisMaxRetry,
		TLSConfig:  tlsConfig,
	})
	return checkPing(ctx, connectionkey)
}
