package caches

import (
	"reflect"
	"testing"
	"time"
)

func Test_RedisClient(t *testing.T) {
	key := "testkey"
	data := "testvalue"
	cached := NewRedis(true, 2)
	defer cached.CloseConnect()
	if err := cached.Set(key, data, 20*time.Second); err != nil {
		t.Fatalf("failed test %#v", err)
	}
	cachedvalue, cachedfound := cached.Get(key)
	if cachedfound == false {
		t.Fatal("Can't Find CacheData")
	}
	if cacheddata, ok := cachedvalue.(string); !ok {
		t.Fatal("failed test Invalid Data Type")
	} else if reflect.DeepEqual(cacheddata, data) == false {
		t.Fatal("failed test Different Data")
	}
	if err := cached.Del(key); err != nil {
		t.Fatalf("failed test %#v", err)
	}
	_, cachedfound = cached.Get(key)
	if cachedfound == true {
		t.Fatal("failed test Del CacheData")
	}
	t.Log("success RedisClient")
}

func Test_RedisClientDelBulk(t *testing.T) {
	key := "testkey"
	data := "testvalue"
	cached := NewRedis(true, 2)
	defer cached.CloseConnect()
	if err := cached.Set(key, data, 20*time.Second); err != nil {
		t.Fatalf("failed test %#v", err)
	}
	cachedvalue, cachedfound := cached.Get(key)
	if cachedfound == false {
		t.Fatal("Can't Find CacheData")
	}
	if cacheddata, ok := cachedvalue.(string); !ok {
		t.Fatal("failed test Invalid Data Type")
	} else if reflect.DeepEqual(cacheddata, data) == false {
		t.Fatal("failed test Different Data")
	}
	if err := cached.DelBulk(key); err != nil {
		t.Fatalf("failed test %#v", err)
	}
	_, cachedfound = cached.Get(key)
	if cachedfound == true {
		t.Fatal("failed test DelBulk CacheData")
	}
	t.Log("success RedisClient")
}
