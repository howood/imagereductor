package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/howood/imagereductor/application/validator"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/howood/imagereductor/infrastructure/requestid"
	"github.com/howood/imagereductor/interfaces/service/config"
	"github.com/howood/imagereductor/interfaces/service/usecase"
	"github.com/howood/imagereductor/library/utils"
	"github.com/labstack/echo/v4"
)

// ImageReductionHandler struct
type ImageReductionHandler struct {
	BaseHandler
}

// Request is get from storage
func (irh ImageReductionHandler) Request(c echo.Context) error {
	requesturi := c.Request().URL.RequestURI()
	xRequestID := requestid.GetRequestID(c.Request())
	irh.ctx = context.WithValue(context.Background(), echo.HeaderXRequestID, xRequestID)
	log.Info(irh.ctx, "========= START REQUEST : "+requesturi)
	log.Info(irh.ctx, c.Request().Method)
	log.Info(irh.ctx, c.Request().Header)
	if c.FormValue(config.FormKeyStorageKey) == "" {
		return irh.errorResponse(c, http.StatusBadRequest, errors.New(config.FormKeyStorageKey+" is required"))
	}
	if c.FormValue(config.FormKeyNonUseCache) != "true" && irh.getCache(c, requesturi) {
		log.Info(irh.ctx, "cache hit!")
		return nil
	}
	contenttype, imagebyte, err := usecase.ImageUsecase{Ctx: irh.ctx}.GetImage(c, c.FormValue(config.FormKeyStorageKey))
	if err != nil {
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	irh.setCache(contenttype, imagebyte, requesturi)
	irh.setResponseHeader(
		c,
		irh.setNewLatsModified(),
		fmt.Sprintf("%d", len(string(imagebyte))),
		irh.setExpires(time.Now()),
		irh.ctx.Value(echo.HeaderXRequestID).(string),
	)
	return c.Blob(http.StatusOK, contenttype, imagebyte)
}

// RequestFile is get non image file from storage
func (irh ImageReductionHandler) RequestFile(c echo.Context) error {
	requesturi := c.Request().URL.RequestURI()
	xRequestID := requestid.GetRequestID(c.Request())
	irh.ctx = context.WithValue(context.Background(), echo.HeaderXRequestID, xRequestID)
	log.Info(irh.ctx, "========= START REQUEST : "+requesturi)
	log.Info(irh.ctx, c.Request().Method)
	log.Info(irh.ctx, c.Request().Header)
	if c.FormValue(config.FormKeyStorageKey) == "" {
		return irh.errorResponse(c, http.StatusBadRequest, errors.New(config.FormKeyStorageKey+" is required"))
	}
	if c.FormValue(config.FormKeyNonUseCache) != "true" && irh.getCache(c, requesturi) {
		log.Info(irh.ctx, "cache hit!")
		return nil
	}
	// get from storage

	contenttype, filebyte, err := usecase.ImageUsecase{Ctx: irh.ctx}.GetFile(c.FormValue(config.FormKeyStorageKey))
	if err != nil {
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	irh.setCache(contenttype, filebyte, requesturi)
	irh.setResponseHeader(
		c,
		irh.setNewLatsModified(),
		fmt.Sprintf("%d", len(string(filebyte))),
		"",
		irh.ctx.Value(echo.HeaderXRequestID).(string),
	)
	return c.Blob(http.StatusOK, contenttype, filebyte)
}

// RequestStreaming is get non image file from storage by streaming
func (irh ImageReductionHandler) RequestStreaming(c echo.Context) error {
	requesturi := c.Request().URL.RequestURI()
	xRequestID := requestid.GetRequestID(c.Request())
	irh.ctx = context.WithValue(context.Background(), echo.HeaderXRequestID, xRequestID)
	log.Info(irh.ctx, "========= START REQUEST : "+requesturi)
	log.Info(irh.ctx, c.Request().Method)
	log.Info(irh.ctx, c.Request().Header)
	if c.FormValue(config.FormKeyStorageKey) == "" {
		return irh.errorResponse(c, http.StatusBadRequest, errors.New(config.FormKeyStorageKey+" is required"))
	}
	// get from storage
	contenttype, contentLength, response, err := usecase.ImageUsecase{Ctx: irh.ctx}.GetFileStream(c.FormValue(config.FormKeyStorageKey))
	if err != nil {
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	defer response.Close()

	irh.setResponseHeader(
		c,
		irh.setNewLatsModified(),
		fmt.Sprintf("%d", contentLength),
		"",
		irh.ctx.Value(echo.HeaderXRequestID).(string),
	)
	c.Response().Header().Set(echo.HeaderContentType, contenttype)
	c.Response().WriteHeader(http.StatusOK)

	_, err = io.Copy(c.Response().Writer, response)
	if err != nil {
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	return nil
}

// RequestInfo is get info from storage
func (irh ImageReductionHandler) RequestInfo(c echo.Context) error {
	requesturi := c.Request().URL.RequestURI()
	xRequestID := requestid.GetRequestID(c.Request())
	irh.ctx = context.WithValue(context.Background(), echo.HeaderXRequestID, xRequestID)
	log.Info(irh.ctx, "========= START REQUEST : "+requesturi)
	log.Info(irh.ctx, c.Request().Method)
	log.Info(irh.ctx, c.Request().Header)
	if c.FormValue(config.FormKeyStorageKey) == "" {
		return irh.errorResponse(c, http.StatusBadRequest, errors.New(config.FormKeyStorageKey+" is required"))
	}
	if c.FormValue(config.FormKeyNonUseCache) != "true" && irh.getCache(c, requesturi) {
		log.Info(irh.ctx, "cache hit!")
		return nil
	}
	// get from storage
	objectInfo, err := usecase.ImageUsecase{Ctx: irh.ctx}.GetFileInfo(c.FormValue(config.FormKeyStorageKey))
	if err != nil {
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	if infoByteData, err := irh.jsonToByte(objectInfo); err != nil {
		log.Error(irh.ctx, err)
	} else {
		irh.setCache(echo.MIMEApplicationJSON, infoByteData, requesturi)
	}
	return c.JSONPretty(http.StatusOK, objectInfo, marshalIndent)
}

// Upload is to upload to storage
func (irh ImageReductionHandler) Upload(c echo.Context) error {
	xRequestID := requestid.GetRequestID(c.Request())
	irh.ctx = context.WithValue(context.Background(), echo.HeaderXRequestID, xRequestID)
	log.Info(irh.ctx, "========= START REQUEST : "+c.Request().URL.RequestURI())
	log.Info(irh.ctx, c.Request().Method)
	log.Info(irh.ctx, c.Request().Header)
	var err error
	// get imageoption
	imageoption, err := usecase.ImageUsecase{Ctx: irh.ctx}.GetImageOptionByFormValue(c)
	if err != nil {
		log.Warn(irh.ctx, err)
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	// read uploaded image
	var file *multipart.FileHeader
	var reader multipart.File
	if err == nil {
		file, err = c.FormFile(config.FormKeyUploadFile)
	}
	if err == nil {
		reader, err = file.Open()
	}
	if err != nil {
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	defer reader.Close()
	//validate
	if err == nil {
		err = irh.validateUploadedImage(reader)
	}
	if err != nil {
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	// resizing image
	convertedimagebyte, err := usecase.ImageUsecase{Ctx: irh.ctx}.ConvertImage(imageoption, reader)
	if err != nil {
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	return usecase.ImageUsecase{Ctx: irh.ctx}.UploadToStorage(c.FormValue(config.FormKeyPath), reader, convertedimagebyte)
}

// UploadFile is to upload non image file to storage
func (irh ImageReductionHandler) UploadFile(c echo.Context) error {
	xRequestID := requestid.GetRequestID(c.Request())
	irh.ctx = context.WithValue(context.Background(), echo.HeaderXRequestID, xRequestID)
	log.Info(irh.ctx, "========= START REQUEST : "+c.Request().URL.RequestURI())
	log.Info(irh.ctx, c.Request().Method)
	log.Info(irh.ctx, c.Request().Header)
	file, err := c.FormFile(config.FormKeyUploadFile)
	if err != nil {
		log.Error(irh.ctx, err)
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	reader, err := file.Open()
	if err != nil {
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	defer reader.Close()
	return usecase.ImageUsecase{Ctx: irh.ctx}.UploadToStorage(c.FormValue(config.FormKeyPath), reader, nil)
}

func (irh ImageReductionHandler) validateUploadedImage(reader multipart.File) error {
	imagetypearray := strings.Split(os.Getenv("VALIDATE_IMAGE_TYPE"), ",")
	maxwidth := utils.GetOsEnvInt("VALIDATE_IMAGE_MAXWIDTH", 5000)
	maxheight := utils.GetOsEnvInt("VALIDATE_IMAGE_MAXHEIGHT", 5000)
	maxfilesize := utils.GetOsEnvInt("VALIDATE_IMAGE_MAXFILESIZE", 104857600)
	imagevalidate := validator.NewImageValidator(irh.ctx, imagetypearray, maxwidth, maxheight, maxfilesize)
	return imagevalidate.Validate(reader)
}

func (irh ImageReductionHandler) getCache(c echo.Context, requesturi string) bool {
	exist, cachedcontent, err := usecase.CacheUsecase{Ctx: irh.ctx}.GetCache(requesturi)
	if !exist {
		return false
	}
	if err != nil {
		log.Error(irh.ctx, err.Error())
		return false
	}
	lastmodified, _ := time.Parse(http.TimeFormat, cachedcontent.GetLastModified())
	irh.setResponseHeader(
		c,
		cachedcontent.GetLastModified(),
		fmt.Sprintf("%d", len(string(cachedcontent.GetContent()))),
		irh.setExpires(lastmodified),
		irh.ctx.Value(echo.HeaderXRequestID).(string),
	)
	c.Response().Header().Set(echo.HeaderContentType, cachedcontent.GetContentType())
	c.Response().WriteHeader(http.StatusOK)
	if _, err = c.Response().Write(cachedcontent.GetContent()); err != nil {
		log.Error(irh.ctx, err.Error())
		return false
	}
	return true
}

func (irh ImageReductionHandler) setCache(mimetype string, data []byte, requesturi string) {
	usecase.CacheUsecase{Ctx: irh.ctx}.SetCache(mimetype, data, requesturi, irh.setNewLatsModified())
}
