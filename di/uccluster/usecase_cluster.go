package uccluster

import (
	"github.com/howood/imagereductor/application/usecase"
)

// DataStore interface.
type UsecaseCluster struct {
	CacheUC *usecase.CacheUsecase
	ImageUC *usecase.ImageUsecase
	TokenUC *usecase.TokenUsecase
}

// NewDatastore returns DataStore interface.
func NewUsecaseCluster() *UsecaseCluster {
	return &UsecaseCluster{
		CacheUC: usecase.NewCacheUsecase(),
		ImageUC: usecase.NewImageUsecase(),
		TokenUC: usecase.NewTokenUsecase(),
	}
}
