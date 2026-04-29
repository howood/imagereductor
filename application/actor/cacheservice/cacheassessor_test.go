package cacheservice_test

import (
	"errors"
	"testing"

	"github.com/howood/imagereductor/application/actor/cacheservice"
)

func Test_GetChacheExpired_Default(t *testing.T) {
	t.Setenv("CACHEEXPIED", "")

	if got := cacheservice.GetChacheExpired(); got != 300 {
		t.Fatalf("GetChacheExpired default = %d, want 300", got)
	}
}

func Test_GetCachedDB_Default(t *testing.T) {
	t.Setenv("CACHEDDB", "")

	if got := cacheservice.GetCachedDB(); got != 0 {
		t.Fatalf("GetCachedDB default = %d, want 0", got)
	}
}

func Test_GetSessionDB_Default(t *testing.T) {
	t.Setenv("SESSIONDB", "")

	if got := cacheservice.GetSessionDB(); got != 1 {
		t.Fatalf("GetSessionDB default = %d, want 1", got)
	}
}

func Test_NewCacheAssessorWithConfig_EmptyType(t *testing.T) {
	// Note: must NOT run in parallel because it manipulates env vars
	t.Setenv("CACHE_TYPE", "")
	_, err := cacheservice.NewCacheAssessorWithConfig(t.Context(), 0)
	if !errors.Is(err, cacheservice.ErrCacheTypeEmpty) {
		t.Fatalf("expected ErrCacheTypeEmpty, got %v", err)
	}
}

func Test_NewCacheAssessorWithConfig_InvalidType(t *testing.T) {
	t.Setenv("CACHE_TYPE", "unknown")
	_, err := cacheservice.NewCacheAssessorWithConfig(t.Context(), 0)
	if !errors.Is(err, cacheservice.ErrInvalidCacheType) {
		t.Fatalf("expected ErrInvalidCacheType, got %v", err)
	}
}

func Test_NewCacheAssessorWithConfig_GoCache(t *testing.T) {
	t.Setenv("CACHE_TYPE", "gocache")
	assessor, err := cacheservice.NewCacheAssessorWithConfig(t.Context(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if assessor == nil {
		t.Fatal("assessor is nil")
	}
}

func Test_CacheAssessor_SetGetDelete_GoCache(t *testing.T) {
	t.Setenv("CACHE_TYPE", "gocache")
	ctx := t.Context()
	assessor, err := cacheservice.NewCacheAssessorWithConfig(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}
	const key = "cacheassessor_test_key"
	if err := assessor.Set(ctx, key, "value", 60); err != nil {
		t.Fatalf("Set: %v", err)
	}
	val, ok, err := assessor.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit")
	}
	if s, _ := val.(string); s != "value" {
		t.Fatalf("Get value = %v, want value", val)
	}
	if err := assessor.Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, ok, err = assessor.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected cache miss after delete")
	}
}
