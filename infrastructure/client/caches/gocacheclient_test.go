package caches_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/howood/imagereductor/infrastructure/client/caches"
)

func Test_GoCacheClient(t *testing.T) {
	t.Parallel()

	setkey := "testkey"
	setdata := "setdata"
	ctx := context.Background()
	client := caches.NewGoCacheClient()
	if err := client.Set(ctx, setkey, setdata, 60*time.Second); err != nil {
		t.Fatal(err)
	}
	getdata, ok := client.Get(ctx, setkey)
	if !ok {
		t.Fatalf("failed to get cache")
	}
	//nolint:forcetypeassert
	if reflect.DeepEqual(getdata.(string), setdata) == false {
		t.Fatalf("failed compare cache data ")
	}
	//nolint:forcetypeassert
	t.Log(getdata.(string))
	t.Log("success GoCacheClient")
}
