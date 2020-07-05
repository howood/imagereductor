package repository

// JwtClaimsRepository interface
type JwtClaimsRepository interface {
	CreateToken(secret string) string
}
