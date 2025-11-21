package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/application/validator"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/howood/imagereductor/infrastructure/requestid"
	"github.com/howood/imagereductor/interfaces/config"
	"github.com/howood/imagereductor/library/utils"
	"github.com/labstack/echo/v4"
)

// ImageReductionHandler struct.
type ImageReductionHandler struct {
	BaseHandler
}

// Request is get from storage.
func (irh ImageReductionHandler) Request(c echo.Context) error {
	requesturi := c.Request().URL.RequestURI()
	xRequestID := requestid.GetRequestID(c.Request())
	ctx := context.WithValue(c.Request().Context(), requestid.GetRequestIDKey(), xRequestID)
	log.Info(ctx, "========= START REQUEST : "+requesturi)
	log.Info(ctx, c.Request().Method)
	log.Info(ctx, c.Request().Header)
	if c.FormValue(config.FormKeyStorageKey) == "" {
		//nolint:err113
		return irh.errorResponse(ctx, c, http.StatusBadRequest, fmt.Errorf("%s is required", config.FormKeyStorageKey))
	}
	if c.FormValue(config.FormKeyNonUseCache) != config.FormValueTrue && irh.getCache(ctx, c, requesturi) {
		log.Info(ctx, "cache hit!")
		return nil
	}
	// get imageoption
	imageoption, err := irh.getImageOptionByFormValue(ctx, c)
	if err != nil {
		log.Warn(ctx, err)
		return irh.errorResponse(ctx, c, http.StatusBadRequest, err)
	}
	contenttype, imagebyte, err := irh.UcCluster.ImageUC.GetImage(ctx, imageoption, c.FormValue(config.FormKeyStorageKey))
	if err != nil {
		return irh.errorResponse(ctx, c, http.StatusBadRequest, err)
	}
	irh.setCache(ctx, contenttype, imagebyte, requesturi)
	irh.setResponseHeader(
		c,
		irh.setNewLatsModified(),
		strconv.Itoa(len(imagebyte)),
		irh.setExpires(time.Now()),
		fmt.Sprintf("%v", ctx.Value(requestid.GetRequestIDKey())),
	)
	return c.Blob(http.StatusOK, contenttype, imagebyte)
}

// RequestFile is get non image file from storage.
func (irh ImageReductionHandler) RequestFile(c echo.Context) error {
	requesturi := c.Request().URL.RequestURI()
	xRequestID := requestid.GetRequestID(c.Request())
	ctx := context.WithValue(c.Request().Context(), requestid.GetRequestIDKey(), xRequestID)
	log.Info(ctx, "========= START REQUEST : "+requesturi)
	log.Info(ctx, c.Request().Method)
	log.Info(ctx, c.Request().Header)
	if c.FormValue(config.FormKeyStorageKey) == "" {
		//nolint:err113
		return irh.errorResponse(ctx, c, http.StatusBadRequest, fmt.Errorf("%s is required", config.FormKeyStorageKey))
	}
	if c.FormValue(config.FormKeyNonUseCache) != config.FormValueTrue && irh.getCache(ctx, c, requesturi) {
		log.Info(ctx, "cache hit!")
		return nil
	}
	// get from storage

	contenttype, filebyte, err := irh.UcCluster.ImageUC.GetFile(ctx, c.FormValue(config.FormKeyStorageKey))
	if err != nil {
		return irh.errorResponse(ctx, c, http.StatusBadRequest, err)
	}
	irh.setCache(ctx, contenttype, filebyte, requesturi)
	irh.setResponseHeader(
		c,
		irh.setNewLatsModified(),
		strconv.Itoa(len(filebyte)),
		"",
		fmt.Sprintf("%v", ctx.Value(requestid.GetRequestIDKey())),
	)
	return c.Blob(http.StatusOK, contenttype, filebyte)
}

// RequestStreaming is get non image file from storage by streaming.
func (irh ImageReductionHandler) RequestStreaming(c echo.Context) error {
	requesturi := c.Request().URL.RequestURI()
	xRequestID := requestid.GetRequestID(c.Request())
	ctx := context.WithValue(c.Request().Context(), requestid.GetRequestIDKey(), xRequestID)
	log.Info(ctx, "========= START REQUEST : "+requesturi)
	log.Info(ctx, c.Request().Method)
	log.Info(ctx, c.Request().Header)
	if c.FormValue(config.FormKeyStorageKey) == "" {
		//nolint:err113
		return irh.errorResponse(ctx, c, http.StatusBadRequest, fmt.Errorf("%s is required", config.FormKeyStorageKey))
	}
	// get from storage
	contenttype, contentLength, response, err := irh.UcCluster.ImageUC.GetFileStream(ctx, c.FormValue(config.FormKeyStorageKey))
	if err != nil {
		return irh.errorResponse(ctx, c, http.StatusBadRequest, err)
	}
	defer response.Close()

	irh.setResponseHeader(
		c,
		irh.setNewLatsModified(),
		strconv.Itoa(contentLength),
		"",
		fmt.Sprintf("%v", ctx.Value(requestid.GetRequestIDKey())),
	)
	c.Response().Header().Set(echo.HeaderContentType, contenttype)
	c.Response().WriteHeader(http.StatusOK)

	_, err = io.Copy(c.Response().Writer, response)
	if err != nil {
		return irh.errorResponse(ctx, c, http.StatusBadRequest, err)
	}
	return nil
}

// RequestInfo is get info from storage.
func (irh ImageReductionHandler) RequestInfo(c echo.Context) error {
	requesturi := c.Request().URL.RequestURI()
	xRequestID := requestid.GetRequestID(c.Request())
	ctx := context.WithValue(c.Request().Context(), requestid.GetRequestIDKey(), xRequestID)
	log.Info(ctx, "========= START REQUEST : "+requesturi)
	log.Info(ctx, c.Request().Method)
	log.Info(ctx, c.Request().Header)
	if c.FormValue(config.FormKeyStorageKey) == "" {
		//nolint:err113
		return irh.errorResponse(ctx, c, http.StatusBadRequest, fmt.Errorf("%s is required", config.FormKeyStorageKey))
	}
	if c.FormValue(config.FormKeyNonUseCache) != config.FormValueTrue && irh.getCache(ctx, c, requesturi) {
		log.Info(ctx, "cache hit!")
		return nil
	}
	// get from storage
	objectInfo, err := irh.UcCluster.ImageUC.GetFileInfo(ctx, c.FormValue(config.FormKeyStorageKey))
	if err != nil {
		return irh.errorResponse(ctx, c, http.StatusBadRequest, err)
	}
	if infoByteData, err := irh.jsonToByte(objectInfo); err != nil {
		log.Error(ctx, err)
	} else {
		irh.setCache(ctx, echo.MIMEApplicationJSON, infoByteData, requesturi)
	}
	return c.JSONPretty(http.StatusOK, objectInfo, marshalIndent)
}

// Upload is to upload to storage.
func (irh ImageReductionHandler) Upload(c echo.Context) error {
	xRequestID := requestid.GetRequestID(c.Request())
	ctx := context.WithValue(c.Request().Context(), requestid.GetRequestIDKey(), xRequestID)
	log.Info(ctx, "========= START REQUEST : "+c.Request().URL.RequestURI())
	log.Info(ctx, c.Request().Method)
	log.Info(ctx, c.Request().Header)
	// get imageoption
	imageoption, err := irh.getImageOptionByFormValue(ctx, c)
	if err != nil {
		log.Warn(ctx, err)
		return irh.errorResponse(ctx, c, http.StatusBadRequest, err)
	}
	// read uploaded image
	file, err := c.FormFile(config.FormKeyUploadFile)
	if err != nil {
		return irh.errorResponse(ctx, c, http.StatusBadRequest, err)
	}
	reader, err := file.Open()
	if err != nil {
		return irh.errorResponse(ctx, c, http.StatusBadRequest, err)
	}
	defer reader.Close()
	// validate
	err = irh.validateUploadedImage(ctx, reader)
	if err != nil {
		return irh.errorResponse(ctx, c, http.StatusBadRequest, err)
	}
	// resizing image
	convertedimagebyte, err := irh.UcCluster.ImageUC.ConvertImage(ctx, imageoption, reader)
	if err != nil {
		return irh.errorResponse(ctx, c, http.StatusBadRequest, err)
	}
	return irh.UcCluster.ImageUC.UploadToStorage(ctx, c.FormValue(config.FormKeyPath), reader, convertedimagebyte)
}

// UploadFile is to upload non image file to storage.
func (irh ImageReductionHandler) UploadFile(c echo.Context) error {
	xRequestID := requestid.GetRequestID(c.Request())
	ctx := context.WithValue(c.Request().Context(), requestid.GetRequestIDKey(), xRequestID)
	log.Info(ctx, "========= START REQUEST : "+c.Request().URL.RequestURI())
	log.Info(ctx, c.Request().Method)
	log.Info(ctx, c.Request().Header)
	file, err := c.FormFile(config.FormKeyUploadFile)
	if err != nil {
		log.Error(ctx, err)
		return irh.errorResponse(ctx, c, http.StatusBadRequest, err)
	}
	reader, err := file.Open()
	if err != nil {
		return irh.errorResponse(ctx, c, http.StatusBadRequest, err)
	}
	defer reader.Close()
	return irh.UcCluster.ImageUC.UploadToStorage(ctx, c.FormValue(config.FormKeyPath), reader, nil)
}

//nolint:mnd
func (irh ImageReductionHandler) validateUploadedImage(ctx context.Context, reader multipart.File) error {
	imagetypearray := strings.Split(os.Getenv("VALIDATE_IMAGE_TYPE"), ",")
	maxwidth := utils.GetOsEnvInt("VALIDATE_IMAGE_MAXWIDTH", 5000)
	maxheight := utils.GetOsEnvInt("VALIDATE_IMAGE_MAXHEIGHT", 5000)
	maxfilesize := utils.GetOsEnvInt("VALIDATE_IMAGE_MAXFILESIZE", 104857600)
	imagevalidate := validator.NewImageValidator(imagetypearray, maxwidth, maxheight, maxfilesize)
	return imagevalidate.Validate(ctx, reader)
}

func (irh ImageReductionHandler) getCache(ctx context.Context, c echo.Context, requesturi string) bool {
	exist, cachedcontent, err := irh.UcCluster.CacheUC.GetCache(ctx, requesturi)
	if !exist {
		return false
	}
	if err != nil {
		log.Error(ctx, err.Error())
		return false
	}
	lastmodified, _ := time.Parse(http.TimeFormat, cachedcontent.GetLastModified())
	irh.setResponseHeader(
		c,
		cachedcontent.GetLastModified(),
		strconv.Itoa(len(cachedcontent.GetContent())),
		irh.setExpires(lastmodified),
		fmt.Sprintf("%v", ctx.Value(requestid.GetRequestIDKey())),
	)
	c.Response().Header().Set(echo.HeaderContentType, cachedcontent.GetContentType())
	c.Response().WriteHeader(http.StatusOK)
	if _, err = c.Response().Write(cachedcontent.GetContent()); err != nil {
		log.Error(ctx, err.Error())
		return false
	}
	return true
}

func (irh ImageReductionHandler) setCache(ctx context.Context, mimetype string, data []byte, requesturi string) {
	irh.UcCluster.CacheUC.SetCache(ctx, mimetype, data, requesturi, irh.setNewLatsModified())
}

func (irh ImageReductionHandler) getImageOptionByFormValue(ctx context.Context, c echo.Context) (actor.ImageOperatorOption, error) {
	var err error
	option := actor.ImageOperatorOption{}
	option.Rotate = c.FormValue(config.FormKeyRotate)
	option.Width, err = irh.setOptionValueInt(ctx, c.FormValue(config.FormKeyWidth), err)
	option.Height, err = irh.setOptionValueInt(ctx, c.FormValue(config.FormKeyHeight), err)
	option.Quality, err = irh.setOptionValueInt(ctx, c.FormValue(config.FormKeyQuality), err)
	option.Brightness, err = irh.setOptionValueInt(ctx, c.FormValue(config.FormKeyBrightness), err)
	option.Contrast, err = irh.setOptionValueInt(ctx, c.FormValue(config.FormKeyContrast), err)
	option.Gamma, err = irh.setOptionValueFloat(ctx, c.FormValue(config.FormKeyGamma), err)
	option.Crop, err = irh.getCropParam(ctx, c.FormValue(config.FormKeyCrop), err)
	return option, err
}

func (irh ImageReductionHandler) setOptionValueInt(ctx context.Context, formvalue string, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	if formvalue == "" {
		return 0, err
	}
	val, err := strconv.Atoi(formvalue)
	if err != nil {
		log.Warn(ctx, err)
		//nolint:err113
		err = errors.New("invalid parameter")
	}
	return val, err
}

func (irh ImageReductionHandler) setOptionValueFloat(ctx context.Context, formvalue string, err error) (float64, error) {
	if err != nil {
		return 0, err
	}
	if formvalue == "" {
		return 0, err
	}
	val, err := strconv.ParseFloat(formvalue, 64)
	if err != nil {
		log.Warn(ctx, err)
		//nolint:err113
		err = errors.New("invalid parameter")
	}
	return val, err
}

//nolint:mnd
func (irh ImageReductionHandler) getCropParam(ctx context.Context, cropparam string, err error) ([4]int, error) {
	if err != nil {
		return [4]int{}, err
	}
	if cropparam == "" {
		return [4]int{}, nil
	}
	crops := strings.Split(cropparam, ",")
	if len(crops) != 4 {
		//nolint:err113
		return [4]int{}, errors.New("crop parameters must need four with comma like : 111,222,333,444")
	}
	intslicecrops := make([]int, 0)
	for _, crop := range crops {
		intcrop, err := strconv.Atoi(crop)
		if err != nil {
			log.Warn(ctx, err)
			//nolint:err113
			err = errors.New("invalid crop parameter")
			return [4]int{}, err
		}
		intslicecrops = append(intslicecrops, intcrop)
	}
	var intcrops [4]int
	copy(intcrops[:], intslicecrops[:4])
	return intcrops, nil
}
