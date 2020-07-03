package actor

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/howood/imagereductor/infrastructure/logger"
)

const TokenExpired = 60

var TokenSecret = os.Getenv("TOKEN_SECRET")

type JWTClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.StandardClaims
}

func NewJWTClaims(username string, admin bool, expired time.Duration) *JWTClaims {
	claims := &JWTClaims{
		username,
		admin,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * expired).Unix(),
		},
	}
	return claims
}
func (jc *JWTClaims) CreateToken(secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jc)
	tokenstring, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Error("", err.Error())
	}
	return tokenstring
}
