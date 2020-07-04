package actor

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/howood/imagereductor/domain/entity"
	"github.com/howood/imagereductor/domain/repository"
	log "github.com/howood/imagereductor/infrastructure/logger"
)

const TokenExpired = 60

var TokenSecret = os.Getenv("TOKEN_SECRET")

type JwtOperator struct {
	jwtClaims *entity.JwtClaims
}

func NewJwtOperator(username string, admin bool, expired time.Duration) repository.JwtClaimsRepository {
	return &JwtOperator{
		jwtClaims: &entity.JwtClaims{
			username,
			admin,
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Minute * expired).Unix(),
			},
		},
	}
}

func (jc *JwtOperator) CreateToken(secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jc.jwtClaims)
	tokenstring, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Error("", err.Error())
	}
	return tokenstring
}
