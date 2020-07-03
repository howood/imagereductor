package handler

import (
	"io"
	"net/http"

	"github.com/howood/imagereductor/application/actor/storageservice"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/labstack/echo/v4"
)

type ImageReductionHandler struct {
}

func (irh ImageReductionHandler) Request(c echo.Context) error {
	log.Info("========= START REQUEST : " + c.Request().URL.RequestURI())
	log.Info(c.Request().Method)
	log.Info(c.Request().Header)
	cloudstorageassessor := storageservice.NewCloudStorageAssessor()
	contenttype, imagebyte, err := cloudstorageassessor.Get(c.FormValue("storagekey"))
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, contenttype, imagebyte)
}

func (irh ImageReductionHandler) Upload(c echo.Context) error {
	log.Info("========= START REQUEST : " + c.Request().URL.RequestURI())
	log.Info(c.Request().Method)
	log.Info(c.Request().Header)
	file, err := c.FormFile("uploadfile")
	if err != nil {
		log.Error(err)
		return err
	}
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()
	cloudstorageassessor := storageservice.NewCloudStorageAssessor()
	return cloudstorageassessor.Put(c.FormValue("path"), reader.(io.ReadSeeker))
}
