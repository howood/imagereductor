package utils

import (
	"testing"
)

func Test_StringArrayContains(t *testing.T) {
	testarray := []string{"aaa", "bbb", "ccc"}
	testarray2 := []string{}
	if StringArrayContains(testarray, "bbb") == false {
		t.Fatal("failed StringArrayContains ")
	}
	if StringArrayContains(testarray, "ddd") == true {
		t.Fatal("failed StringArrayContains ")
	}
	if StringArrayContains(testarray2, "aaa") == true {
		t.Fatal("failed StringArrayContains ")
	}
	if StringArrayContains(nil, "aaa") == true {
		t.Fatal("failed StringArrayContains ")
	}
	t.Log("success StringArrayContains")
}

func Test_StringArrayContainsForwardMatch(t *testing.T) {
	testarray := []string{"abc", "bcd", "ert"}
	testarray2 := []string{}
	if StringArrayContainsForwardMatch(testarray, "bc") == false {
		t.Fatal("failed StringArrayContainsForwardMatch ")
	}
	if StringArrayContainsForwardMatch(testarray, "rt") == true {
		t.Fatal("failed StringArrayContainsForwardMatch ")
	}
	if StringArrayContainsForwardMatch(testarray2, "aaa") == true {
		t.Fatal("failed StringArrayContainsForwardMatch ")
	}
	if StringArrayContainsForwardMatch(nil, "aaa") == true {
		t.Fatal("failed StringArrayContainsForwardMatch ")
	}
	t.Log("success StringArrayContainsForwardMatch")
}

func Test_IntArrayContains(t *testing.T) {
	testarray := []int{234, 435, 7666}
	testarray2 := []int{}
	if IntArrayContains(testarray, 435) == false {
		t.Fatal("failed IntArrayContains ")
	}
	if IntArrayContains(testarray, 297) == true {
		t.Fatal("failed IntArrayContains ")
	}
	if IntArrayContains(testarray2, 545) == true {
		t.Fatal("failed IntArrayContains ")
	}
	if IntArrayContains(nil, 324) == true {
		t.Fatal("failed IntArrayContains ")
	}
	t.Log("success IntArrayContains")
}
