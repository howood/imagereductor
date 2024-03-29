package caches

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func Test_GoCacheClient(t *testing.T) {
	setkey := "testkey"
	setdata := "setdata"
	client := NewGoCacheClient(context.Background())
	if err := client.Set(setkey, setdata, 60*time.Second); err != nil {
		t.Fatal(err)
	}
	getdata, ok := client.Get(setkey)
	if !ok {
		t.Fatalf("failed to get cache")
	}
	if reflect.DeepEqual(getdata.(string), setdata) == false {
		t.Fatalf("failed compare cache data ")
	}
	t.Log(getdata.(string))
	t.Log("success GoCacheClient")
}
