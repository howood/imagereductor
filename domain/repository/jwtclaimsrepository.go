package repository

type JwtClaimsRepository interface {
	CreateToken(secret string) string
}
