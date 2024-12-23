package usecase

import (
	"context"

	"github.com/howood/imagereductor/application/actor"
)

type TokenUsecase struct{}

func (tu TokenUsecase) CreateToken(ctx context.Context, claimname string) string {
	jwtinstance := actor.NewJwtOperator(claimname, false)
	return jwtinstance.CreateToken(ctx)
}
