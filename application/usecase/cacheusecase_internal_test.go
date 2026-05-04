package usecase

import (
	"testing"

	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/application/actor/cacheservice"
)

func Test_GetCache_StringType(t *testing.T) {
	t.Setenv("CACHE_TYPE", "gocache")

	ctx := t.Context()
	uc, err := NewCacheUsecaseWithConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Encode valid gob data and store as string (not []byte).
	src := actor.NewCachedContentOperator()
	src.Set("image/png", "Mon, 01 Jan 2024 00:00:00 GMT", []byte("test-data"))
	encoded, err := src.GobEncode()
	if err != nil {
		t.Fatal(err)
	}

	// Set raw string directly via cacheAssessor to hit the string case in GetCache
	if err := uc.cacheAssessor.Set(ctx, "/test-string-type", string(encoded), cacheservice.GetChacheExpired()); err != nil {
		t.Fatal(err)
	}

	found, content, err := uc.GetCache(ctx, "/test-string-type")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected cache hit")
	}
	if content.GetContentType() != "image/png" {
		t.Fatalf("ContentType = %q, want image/png", content.GetContentType())
	}
}

func Test_GetCache_DefaultType(t *testing.T) {
	t.Setenv("CACHE_TYPE", "gocache")

	ctx := t.Context()
	uc, err := NewCacheUsecaseWithConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Set an int value to trigger the default case
	if err := uc.cacheAssessor.Set(ctx, "/test-default-type", 12345, cacheservice.GetChacheExpired()); err != nil {
		t.Fatal(err)
	}

	found, _, err := uc.GetCache(ctx, "/test-default-type")
	if err == nil {
		t.Fatal("expected error for default type, got nil")
	}
	if !found {
		t.Fatal("expected found=true even on error")
	}
}

func Test_GetCache_InvalidGobBytes(t *testing.T) {
	t.Setenv("CACHE_TYPE", "gocache")

	ctx := t.Context()
	uc, err := NewCacheUsecaseWithConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Set invalid gob data as []byte to trigger GobDecode error
	if err := uc.cacheAssessor.Set(ctx, "/test-invalid-gob", []byte("not-valid-gob"), cacheservice.GetChacheExpired()); err != nil {
		t.Fatal(err)
	}

	found, _, err := uc.GetCache(ctx, "/test-invalid-gob")
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
	if !found {
		t.Fatal("expected found=true even on decode error")
	}
}
