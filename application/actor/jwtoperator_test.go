package actor_test

import (
	"strings"
	"testing"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/domain/entity"
)

func Test_JwtOperator_CreateToken(t *testing.T) {
	t.Parallel()

	op := actor.NewJwtOperator("alice", true)
	tokenstr := op.CreateToken(t.Context())
	if tokenstr == "" {
		t.Fatal("CreateToken returned empty string")
	}
	// JWT format: header.payload.signature
	if parts := strings.Split(tokenstr, "."); len(parts) != 3 {
		t.Fatalf("expected JWT to have 3 parts, got %d (%q)", len(parts), tokenstr)
	}

	parsed, err := jwt.ParseWithClaims(tokenstr, &entity.JwtClaims{}, func(_ *jwt.Token) (any, error) {
		return []byte(actor.TokenSecret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse generated JWT: %v", err)
	}
	if !parsed.Valid {
		t.Fatal("parsed token is not valid")
	}
	claims, ok := parsed.Claims.(*entity.JwtClaims)
	if !ok {
		t.Fatalf("unexpected claims type: %T", parsed.Claims)
	}
	if claims.Name != "alice" {
		t.Fatalf("claim Name = %q, want alice", claims.Name)
	}
	if !claims.Admin {
		t.Fatal("claim Admin = false, want true")
	}
}
