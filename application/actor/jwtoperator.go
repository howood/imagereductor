package actor

import (
	"context"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/howood/imagereductor/domain/entity"
	"github.com/howood/imagereductor/domain/repository"
	log "github.com/howood/imagereductor/infrastructure/logger"
)

// TokenExpired is token's expired
const TokenExpired = 60

// TokenSecret define token secrets
var TokenSecret = os.Getenv("TOKEN_SECRET")

// JwtOperator struct
type JwtOperator struct {
	jwtClaims *entity.JwtClaims
	ctx       context.Context
}

// NewJwtOperator creates a new JwtClaimsRepository
func NewJwtOperator(ctx context.Context, username string, admin bool, expired time.Duration) repository.JwtClaimsRepository {
	return &JwtOperator{
		jwtClaims: &entity.JwtClaims{
			username,
			admin,
			jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * expired)),
			},
		},
		ctx: ctx,
	}
}

// CreateToken creates a new token
func (jc *JwtOperator) CreateToken(secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jc.jwtClaims)
	tokenstring, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Error(jc.ctx, err.Error())
	}
	return tokenstring
}
