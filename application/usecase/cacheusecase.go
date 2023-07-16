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
	Ctx context.Context
}

func (cu CacheUsecase) GetCache(requesturi string) (bool, repository.CachedContentRepository, error) {
	cacheAssessor := cacheservice.NewCacheAssessor(cu.Ctx, cacheservice.GetCachedDB())
	if cachedvalue, cachedfound := cacheAssessor.Get(requesturi); cachedfound {
		cachedcontent := actor.NewCachedContentOperator()
		var err error
		switch xi := cachedvalue.(type) {
		case []byte:
			err = cachedcontent.GobDecode(xi)
		case string:
			err = cachedcontent.GobDecode([]byte(xi))
		default:
			err = errors.New("get cache error")
		}
		if err != nil {
			log.Error(cu.Ctx, err.Error())
			return true, cachedcontent, err
		}
		return true, cachedcontent, err
	}
	return false, nil, nil
}

func (cu CacheUsecase) SetCache(mimetype string, data []byte, requesturi string, latsModified string) {
	cachedresponse := actor.NewCachedContentOperator()
	cachedresponse.Set(mimetype, latsModified, data)
	encodedcached, err := cachedresponse.GobEncode()
	if err != nil {
		log.Error(cu.Ctx, err)
	} else {
		cacheAssessor := cacheservice.NewCacheAssessor(cu.Ctx, cacheservice.GetCachedDB())
		if setErr := cacheAssessor.Set(requesturi, encodedcached, cacheservice.GetChacheExpired()); setErr != nil {
			log.Error(cu.Ctx, setErr)
		}
	}
}
