package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/application/actor/cacheservice"
	"github.com/howood/imagereductor/application/actor/storageservice"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/labstack/echo/v4"
)

// ImageReductionHandler struct
type ImageReductionHandler struct {
}

// Request is get from storage
func (irh ImageReductionHandler) Request(c echo.Context) error {
	requesturi := c.Request().URL.RequestURI()
	log.Info("========= START REQUEST : " + requesturi)
	log.Info(c.Request().Method)
	log.Info(c.Request().Header)
	if irh.getCache(c, requesturi) {
		log.Info("cache hit!")
		return nil
	}
	cloudstorageassessor := storageservice.NewCloudStorageAssessor()
	contenttype, imagebyte, err := cloudstorageassessor.Get(c.FormValue("key"))
	if err != nil {
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	width, _ := strconv.Atoi(c.FormValue("w"))
	height, _ := strconv.Atoi(c.FormValue("h"))
	quality, _ := strconv.Atoi(c.FormValue("q"))
	if width > 0 || height > 0 {
		imageOperator := actor.NewImageOperator(
			contenttype,
			actor.ImageOperatorOption{
				Width:   width,
				Height:  height,
				Quality: quality,
			},
		)
		imageOperator.Decode(bytes.NewBuffer(imagebyte))
		imageOperator.Resize()
		var err error
		if imagebyte, err = imageOperator.ImageByte(); err != nil {
			return irh.errorResponse(c, http.StatusBadRequest, err)
		}
	}
	irh.setCache(contenttype, imagebyte, requesturi)
	return c.Blob(http.StatusOK, contenttype, imagebyte)
}

// Upload is to upload to storage
func (irh ImageReductionHandler) Upload(c echo.Context) error {
	log.Info("========= START REQUEST : " + c.Request().URL.RequestURI())
	log.Info(c.Request().Method)
	log.Info(c.Request().Header)
	file, err := c.FormFile("uploadfile")
	if err != nil {
		log.Error(err)
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	reader, err := file.Open()
	if err != nil {
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	defer reader.Close()
	cloudstorageassessor := storageservice.NewCloudStorageAssessor()
	return cloudstorageassessor.Put(c.FormValue("path"), reader.(io.ReadSeeker))
}

func (irh ImageReductionHandler) errorResponse(c echo.Context, statudcode int, err error) error {
	if strings.Contains(strings.ToLower(err.Error()), storageservice.RecordNotFoundMsg) {
		statudcode = http.StatusNotFound
	}
	return c.JSONPretty(statudcode, map[string]interface{}{"message": err.Error()}, "    ")
}

func (irh ImageReductionHandler) getCache(c echo.Context, requesturi string) bool {
	cacheAssessor := cacheservice.NewCacheAssessor(cacheservice.GetCachedDB())
	if cachedvalue, cachedfound := cacheAssessor.Get(requesturi); cachedfound {
		cachedcontent := &actor.CachedContent{}
		switch xi := cachedvalue.(type) {
		case []byte:
			if err := cachedcontent.GobDecode(xi); err != nil {
				log.Error("GobDecode Error byte")
				log.Error(err.Error())
				return false
			}
		case string:
			if err := cachedcontent.GobDecode([]byte(xi)); err != nil {
				log.Error("GobDecode Error string")
				log.Error(err.Error())
				return false
			}

		default:
			log.Error("get cache error")
			return false
		}

		c.Response().Header().Set(echo.HeaderContentType, cachedcontent.ContentType)
		c.Response().Header().Set(echo.HeaderLastModified, cachedcontent.LastModified)
		c.Response().Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", len(string(cachedcontent.Content))))
		c.Response().WriteHeader(http.StatusOK)
		_, err := c.Response().Write(cachedcontent.Content)
		if err != nil {
			log.Error(err.Error())
			return false
		}
		return true
	}
	log.Debug("No Cache")
	return false
}

func (irh ImageReductionHandler) setCache(mimetype string, data []byte, requesturi string) {
	cachedresponse := actor.CachedContent{}
	cachedresponse.ContentType = mimetype
	cachedresponse.LastModified = time.Now().UTC().Format(http.TimeFormat)
	cachedresponse.Content = data
	encodedcached, err := cachedresponse.GobEncode()
	if err != nil {
		log.Error(err)
	} else {
		cacheAssessor := cacheservice.NewCacheAssessor(cacheservice.GetCachedDB())
		cacheAssessor.Set(requesturi, encodedcached, cacheservice.GetChacheExpired())
	}
}
