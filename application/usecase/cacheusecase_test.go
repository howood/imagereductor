package usecase_test

import (
	"reflect"
	"testing"

	"github.com/howood/imagereductor/application/usecase"
)

func Test_CacheUsecase_SetGet(t *testing.T) {
	t.Setenv("CACHE_TYPE", "gocache")

	ctx := t.Context()
	uc, err := usecase.NewCacheUsecaseWithConfig(ctx)
	if err != nil {
		t.Fatalf("NewCacheUsecaseWithConfig: %v", err)
	}

	uri := "/path/to/resource?x=1"
	uc.SetCache(ctx, "image/png", []byte("hello"), uri, "Mon, 01 Jan 2024 00:00:00 GMT")

	found, content, err := uc.GetCache(ctx, uri)
	if err != nil {
		t.Fatalf("GetCache error: %v", err)
	}
	if !found {
		t.Fatal("expected cache hit")
	}
	if content.GetContentType() != "image/png" {
		t.Fatalf("ContentType = %q, want image/png", content.GetContentType())
	}
	if !reflect.DeepEqual(content.GetContent(), []byte("hello")) {
		t.Fatalf("Content = %v, want hello", content.GetContent())
	}
}

func Test_CacheUsecase_GetCache_Miss(t *testing.T) {
	t.Setenv("CACHE_TYPE", "gocache")

	ctx := t.Context()
	uc, err := usecase.NewCacheUsecaseWithConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found, content, err := uc.GetCache(ctx, "/nonexistent_path_xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Fatal("expected cache miss")
	}
	if content != nil {
		t.Fatal("expected nil content on miss")
	}
}

func Test_NewCacheUsecaseWithConfig_EmptyType(t *testing.T) {
	t.Setenv("CACHE_TYPE", "")
	_, err := usecase.NewCacheUsecaseWithConfig(t.Context())
	if err == nil {
		t.Fatal("expected error for empty CACHE_TYPE")
	}
}
