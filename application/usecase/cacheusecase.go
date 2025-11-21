package usecase

import (
	"context"
	"errors"

	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/application/actor/cacheservice"
	"github.com/howood/imagereductor/domain/repository"
	log "github.com/howood/imagereductor/infrastructure/logger"
)

type CacheUsecase struct {
	cacheAssessor *cacheservice.CacheAssessor
}

// NewCacheUsecase creates a new CacheUsecase.
//
// Deprecated: Use NewCacheUsecaseWithConfig for better error handling.
func NewCacheUsecase() *CacheUsecase {
	uc, err := NewCacheUsecaseWithConfig(context.Background())
	if err != nil {
		panic(err)
	}
	return uc
}

// NewCacheUsecaseWithConfig creates a new CacheUsecase with proper error handling.
func NewCacheUsecaseWithConfig(ctx context.Context) (*CacheUsecase, error) {
	cacheAssessor, err := cacheservice.NewCacheAssessorWithConfig(ctx, cacheservice.GetCachedDB())
	if err != nil {
		return nil, err
	}
	return &CacheUsecase{
		cacheAssessor: cacheAssessor,
	}, nil
}

//nolint:ireturn
func (cu *CacheUsecase) GetCache(ctx context.Context, requesturi string) (bool, repository.CachedContentRepository, error) {
	if cachedvalue, cachedfound, err := cu.cacheAssessor.Get(ctx, requesturi); err != nil {
		return false, nil, err
	} else if cachedfound {
		cachedcontent := actor.NewCachedContentOperator()
		var err error
		switch xi := cachedvalue.(type) {
		case []byte:
			err = cachedcontent.GobDecode(xi)
		case string:
			err = cachedcontent.GobDecode([]byte(xi))
		default:
			//nolint:err113
			err = errors.New("get cache error")
		}
		if err != nil {
			log.Error(ctx, err.Error())
			return true, cachedcontent, err
		}
		return true, cachedcontent, err
	}
	return false, nil, nil
}

func (cu *CacheUsecase) SetCache(ctx context.Context, mimetype string, data []byte, requesturi string, latsModified string) {
	cachedresponse := actor.NewCachedContentOperator()
	cachedresponse.Set(mimetype, latsModified, data)
	encodedcached, err := cachedresponse.GobEncode()
	if err != nil {
		log.Error(ctx, err)
	} else {
		if setErr := cu.cacheAssessor.Set(ctx, requesturi, encodedcached, cacheservice.GetChacheExpired()); setErr != nil {
			log.Error(ctx, setErr)
		}
	}
}
