package repository

// JwtClaimsRepository interface
type JwtClaimsRepository interface {
	CreateToken() string
}
