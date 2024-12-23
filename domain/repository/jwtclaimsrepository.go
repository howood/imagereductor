package repository

import "context"

// JwtClaimsRepository interface.
type JwtClaimsRepository interface {
	CreateToken(ctx context.Context) string
}
