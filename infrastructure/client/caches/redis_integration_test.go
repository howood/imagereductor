package caches_test

import (
	"context"
	"testing"
	"time"

	"github.com/howood/imagereductor/infrastructure/client/caches"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func setupRedis(t *testing.T) *caches.RedisInstance {
	t.Helper()
	ctx := t.Context()

	container, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("start redis container: %v", err)
	}
	t.Cleanup(func() {
		if termErr := container.Terminate(context.Background()); termErr != nil {
			t.Logf("terminate redis: %v", termErr)
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("redis host: %v", err)
	}
	port, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		t.Fatalf("redis port: %v", err)
	}

	t.Setenv("REDISHOST", host)
	t.Setenv("REDISPORT", port.Port())
	t.Setenv("REDISPASSWORD", "")
	t.Setenv("REDISTLS", "")

	inst := caches.NewRedis(false, 0)
	t.Cleanup(func() { _ = inst.CloseConnect() })
	return inst
}

func TestRedisIntegration_SetAndGet(t *testing.T) { //nolint:paralleltest
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupRedis(t)
	ctx := t.Context()

	if err := inst.Set(ctx, "testkey", "testvalue", 10*time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}

	val, found, err := inst.Get(ctx, "testkey")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("expected cache hit")
	}
	if val != "testvalue" {
		t.Fatalf("Get = %v, want testvalue", val)
	}
}

func TestRedisIntegration_GetMiss(t *testing.T) { //nolint:paralleltest
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupRedis(t)
	ctx := t.Context()

	_, found, err := inst.Get(ctx, "nonexistent_key_xyz")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("expected cache miss")
	}
}

func TestRedisIntegration_Del(t *testing.T) { //nolint:paralleltest
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupRedis(t)
	ctx := t.Context()

	if err := inst.Set(ctx, "delkey", "val", 30*time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := inst.Del(ctx, "delkey"); err != nil {
		t.Fatalf("Del: %v", err)
	}
	_, found, err := inst.Get(ctx, "delkey")
	if err != nil {
		t.Fatalf("Get after Del: %v", err)
	}
	if found {
		t.Fatal("expected miss after Del")
	}
}

func TestRedisIntegration_DelBulk(t *testing.T) { //nolint:paralleltest
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupRedis(t)
	ctx := t.Context()

	for _, k := range []string{"bulk:a", "bulk:b", "bulk:c"} {
		if err := inst.Set(ctx, k, "v", 30*time.Second); err != nil {
			t.Fatalf("Set %s: %v", k, err)
		}
	}
	if err := inst.DelBulk(ctx, "bulk:*"); err != nil {
		t.Fatalf("DelBulk: %v", err)
	}
	for _, k := range []string{"bulk:a", "bulk:b", "bulk:c"} {
		_, found, _ := inst.Get(ctx, k)
		if found {
			t.Fatalf("expected %s deleted", k)
		}
	}
}

func TestRedisIntegration_CloseConnect_Persistent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := t.Context()

	container, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("start redis: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "6379/tcp")
	t.Setenv("REDISHOST", host)
	t.Setenv("REDISPORT", port.Port())
	t.Setenv("REDISPASSWORD", "")
	t.Setenv("REDISTLS", "")

	// persistent connection should NOT close on CloseConnect.
	inst := caches.NewRedis(true, 1)
	if err := inst.CloseConnect(); err != nil {
		t.Fatalf("CloseConnect persistent: %v", err)
	}
	// Should still work after CloseConnect on persistent.
	if err := inst.Set(ctx, "still", "alive", 10*time.Second); err != nil {
		t.Fatalf("Set after persistent close: %v", err)
	}
}
