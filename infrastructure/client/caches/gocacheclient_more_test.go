package caches_test

import (
	"testing"
	"time"

	"github.com/howood/imagereductor/infrastructure/client/caches"
)

func Test_GoCacheClient_Miss(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := caches.NewGoCacheClient()
	_, ok, err := client.Get(ctx, "nonexistent_key_xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected cache miss but got hit")
	}
}

func Test_GoCacheClient_Del(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := caches.NewGoCacheClient()
	key := "del_test_key"
	if err := client.Set(ctx, key, "value", 60*time.Second); err != nil {
		t.Fatal(err)
	}
	if err := client.Del(ctx, key); err != nil {
		t.Fatal(err)
	}
	_, ok, err := client.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func Test_GoCacheClient_DelBulk(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := caches.NewGoCacheClient()
	key := "delbulk_test_key"
	if err := client.Set(ctx, key, "value", 60*time.Second); err != nil {
		t.Fatal(err)
	}
	if err := client.DelBulk(ctx, key); err != nil {
		t.Fatal(err)
	}
	_, ok, err := client.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected key to be deleted by DelBulk")
	}
}

func Test_GoCacheClient_CloseConnect(t *testing.T) {
	t.Parallel()

	client := caches.NewGoCacheClient()
	if err := client.CloseConnect(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
