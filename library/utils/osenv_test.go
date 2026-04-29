package utils_test

import (
	"testing"

	"github.com/howood/imagereductor/library/utils"
)

func Test_GetOsEnv(t *testing.T) {
	const key = "IMAGEREDUCTOR_TEST_OSENV_STR"
	t.Setenv(key, "hello")

	if got := utils.GetOsEnv(key, "default"); got != "hello" {
		t.Fatalf("GetOsEnv with set env = %v, want hello", got)
	}
	if got := utils.GetOsEnv("IMAGEREDUCTOR_TEST_OSENV_NOTSET", "fallback"); got != "fallback" {
		t.Fatalf("GetOsEnv with unset env = %v, want fallback", got)
	}
}

func Test_GetOsEnvInt(t *testing.T) {
	const key = "IMAGEREDUCTOR_TEST_OSENV_INT"
	t.Setenv(key, "42")
	if got := utils.GetOsEnvInt(key, 1); got != 42 {
		t.Fatalf("GetOsEnvInt with set env = %v, want 42", got)
	}

	const invalidKey = "IMAGEREDUCTOR_TEST_OSENV_INVALID"
	t.Setenv(invalidKey, "notanumber")
	if got := utils.GetOsEnvInt(invalidKey, 7); got != 7 {
		t.Fatalf("GetOsEnvInt with invalid env = %v, want 7", got)
	}

	if got := utils.GetOsEnvInt("IMAGEREDUCTOR_TEST_OSENV_INT_NOTSET", 99); got != 99 {
		t.Fatalf("GetOsEnvInt with unset env = %v, want 99", got)
	}
}
