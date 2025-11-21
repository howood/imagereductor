package uccluster

import (
	"context"

	"github.com/howood/imagereductor/application/usecase"
)

// UsecaseCluster interface.
type UsecaseCluster struct {
	CacheUC *usecase.CacheUsecase
	ImageUC *usecase.ImageUsecase
	TokenUC *usecase.TokenUsecase
}

// NewUsecaseCluster returns UsecaseCluster interface.
func NewUsecaseCluster() (*UsecaseCluster, error) {
	ctx := context.Background()
	cacheUC, err := usecase.NewCacheUsecaseWithConfig(ctx)
	if err != nil {
		return nil, err
	}
	imageUC, err := usecase.NewImageUsecaseWithConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &UsecaseCluster{
		CacheUC: cacheUC,
		ImageUC: imageUC,
		TokenUC: usecase.NewTokenUsecase(),
	}, nil
}
