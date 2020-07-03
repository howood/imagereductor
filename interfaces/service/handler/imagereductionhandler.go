package handler

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/application/actor/storageservice"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/labstack/echo/v4"
)

// ImageReductionHandler struct
type ImageReductionHandler struct {
}

// Request is get from storage
func (irh ImageReductionHandler) Request(c echo.Context) error {
	log.Info("========= START REQUEST : " + c.Request().URL.RequestURI())
	log.Info(c.Request().Method)
	log.Info(c.Request().Header)
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
