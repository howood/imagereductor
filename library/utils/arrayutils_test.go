package utils_test

import (
	"testing"

	"github.com/howood/imagereductor/library/utils"
)

func Test_StringArrayContains(t *testing.T) {
	t.Parallel()

	testarray := []string{"aaa", "bbb", "ccc"}
	testarray2 := []string{}
	if utils.StringArrayContains(testarray, "bbb") == false {
		t.Fatal("failed StringArrayContains ")
	}
	if utils.StringArrayContains(testarray, "ddd") == true {
		t.Fatal("failed StringArrayContains ")
	}
	if utils.StringArrayContains(testarray2, "aaa") == true {
		t.Fatal("failed StringArrayContains ")
	}
	if utils.StringArrayContains(nil, "aaa") == true {
		t.Fatal("failed StringArrayContains ")
	}
	t.Log("success StringArrayContains")
}

func Test_StringArrayContainsForwardMatch(t *testing.T) {
	t.Parallel()

	testarray := []string{"abc", "bcd", "ert"}
	testarray2 := []string{}
	if utils.StringArrayContainsForwardMatch(testarray, "bc") == false {
		t.Fatal("failed StringArrayContainsForwardMatch ")
	}
	if utils.StringArrayContainsForwardMatch(testarray, "rt") == true {
		t.Fatal("failed StringArrayContainsForwardMatch ")
	}
	if utils.StringArrayContainsForwardMatch(testarray2, "aaa") == true {
		t.Fatal("failed StringArrayContainsForwardMatch ")
	}
	if utils.StringArrayContainsForwardMatch(nil, "aaa") == true {
		t.Fatal("failed StringArrayContainsForwardMatch ")
	}
	t.Log("success StringArrayContainsForwardMatch")
}

func Test_IntArrayContains(t *testing.T) {
	t.Parallel()

	testarray := []int{234, 435, 7666}
	testarray2 := []int{}
	if utils.IntArrayContains(testarray, 435) == false {
		t.Fatal("failed IntArrayContains ")
	}
	if utils.IntArrayContains(testarray, 297) == true {
		t.Fatal("failed IntArrayContains ")
	}
	if utils.IntArrayContains(testarray2, 545) == true {
		t.Fatal("failed IntArrayContains ")
	}
	if utils.IntArrayContains(nil, 324) == true {
		t.Fatal("failed IntArrayContains ")
	}
	t.Log("success IntArrayContains")
}
