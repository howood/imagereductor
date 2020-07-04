package caches

import (
	"reflect"
	"testing"
)

func Test_GoCacheClient(t *testing.T) {
	setkey := "testkey"
	setdata := "setdata"
	client := NewGoCacheClient(5)
	client.Set(setkey, setdata, 5)
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
