package entity

import jwt "github.com/golang-jwt/jwt/v5"

// JwtClaims entity.
type JwtClaims struct {
	jwt.RegisteredClaims

	Name  string `json:"name"`
	Admin bool   `json:"admin"`
}
