package entity

import jwt "github.com/golang-jwt/jwt/v5"

// JwtClaims entity
type JwtClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}
