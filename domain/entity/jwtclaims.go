package entity

import "github.com/dgrijalva/jwt-go"

// JwtClaims entity
type JwtClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.StandardClaims
}
