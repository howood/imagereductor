package usecase

import (
	"context"
	"errors"

	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/application/actor/cacheservice"
	"github.com/howood/imagereductor/domain/repository"
	log "github.com/howood/imagereductor/infrastructure/logger"
)

type CacheUsecase struct{}

//nolint:ireturn
func (cu CacheUsecase) GetCache(ctx context.Context, requesturi string) (bool, repository.CachedContentRepository, error) {
	cacheAssessor := cacheservice.NewCacheAssessor(ctx, cacheservice.GetCachedDB())
	if cachedvalue, cachedfound := cacheAssessor.Get(ctx, requesturi); cachedfound {
		cachedcontent := actor.NewCachedContentOperator()
		var err error
		switch xi := cachedvalue.(type) {
		case []byte:
			err = cachedcontent.GobDecode(xi)
		case string:
			err = cachedcontent.GobDecode([]byte(xi))
		default:
			//nolint:goerr113
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

func (cu CacheUsecase) SetCache(ctx context.Context, mimetype string, data []byte, requesturi string, latsModified string) {
	cachedresponse := actor.NewCachedContentOperator()
	cachedresponse.Set(mimetype, latsModified, data)
	encodedcached, err := cachedresponse.GobEncode()
	if err != nil {
		log.Error(ctx, err)
	} else {
		cacheAssessor := cacheservice.NewCacheAssessor(ctx, cacheservice.GetCachedDB())
		if setErr := cacheAssessor.Set(ctx, requesturi, encodedcached, cacheservice.GetChacheExpired()); setErr != nil {
			log.Error(ctx, setErr)
		}
	}
}
