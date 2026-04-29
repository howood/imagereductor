package usecase_test

import (
	"strings"
	"testing"

	"github.com/howood/imagereductor/application/usecase"
)

func Test_TokenUsecase_CreateToken(t *testing.T) {
	t.Parallel()

	uc := usecase.NewTokenUsecase()
	token := uc.CreateToken(t.Context(), "alice")
	if token == "" {
		t.Fatal("CreateToken returned empty string")
	}
	if parts := strings.Split(token, "."); len(parts) != 3 {
		t.Fatalf("expected JWT to have 3 parts, got %d", len(parts))
	}
}
