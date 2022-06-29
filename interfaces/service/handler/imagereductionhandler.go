package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/application/actor/cacheservice"
	"github.com/howood/imagereductor/application/actor/storageservice"
	"github.com/howood/imagereductor/application/validator"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/howood/imagereductor/infrastructure/requestid"
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
	if c.FormValue(FormKeyStorageKey) == "" {
		return irh.errorResponse(c, http.StatusBadRequest, errors.New(FormKeyStorageKey+" is required"))
	}
	if c.FormValue(FormKeyNonUseCache) != "true" && irh.getCache(c, requesturi) {
		log.Info(irh.ctx, "cache hit!")
		return nil
	}
	var err error
	var contenttype string
	var imagebyte []byte
	// get imageoption
	imageoption, err := irh.getImageOptionByFormValue(c)
	// get from storage
	if err == nil {
		cloudstorageassessor := storageservice.NewCloudStorageAssessor(irh.ctx)
		contenttype, imagebyte, err = cloudstorageassessor.Get(c.FormValue(FormKeyStorageKey))
	}
	// resizing image
	if err == nil && reflect.DeepEqual(imageoption, actor.ImageOperatorOption{}) == false {
		imageOperator := actor.NewImageOperator(
			irh.ctx,
			contenttype,
			imageoption,
		)
		err = imageOperator.Decode(bytes.NewBuffer(imagebyte))
		if err == nil {
			err = imageOperator.Process()
		}
		if err == nil {
			imagebyte, err = imageOperator.ImageByte()
		}
	}
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
	if c.FormValue(FormKeyStorageKey) == "" {
		return irh.errorResponse(c, http.StatusBadRequest, errors.New(FormKeyStorageKey+" is required"))
	}
	if c.FormValue(FormKeyNonUseCache) != "true" && irh.getCache(c, requesturi) {
		log.Info(irh.ctx, "cache hit!")
		return nil
	}
	cloudstorageassessor := storageservice.NewCloudStorageAssessor(irh.ctx)
	contenttype, filebyte, err := cloudstorageassessor.Get(c.FormValue(FormKeyStorageKey))
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
	if c.FormValue(FormKeyStorageKey) == "" {
		return irh.errorResponse(c, http.StatusBadRequest, errors.New(FormKeyStorageKey+" is required"))
	}
	cloudstorageassessor := storageservice.NewCloudStorageAssessor(irh.ctx)
	contenttype, response, err := cloudstorageassessor.GetByStreaming(c.FormValue(FormKeyStorageKey))
	if err != nil {
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	defer response.Close()

	irh.setResponseHeader(
		c,
		irh.setNewLatsModified(),
		"",
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

// Upload is to upload to storage
func (irh ImageReductionHandler) Upload(c echo.Context) error {
	xRequestID := requestid.GetRequestID(c.Request())
	irh.ctx = context.WithValue(context.Background(), echo.HeaderXRequestID, xRequestID)
	log.Info(irh.ctx, "========= START REQUEST : "+c.Request().URL.RequestURI())
	log.Info(irh.ctx, c.Request().Method)
	log.Info(irh.ctx, c.Request().Header)
	var err error
	// get imageoption
	imageoption, err := irh.getImageOptionByFormValue(c)
	if err != nil {
		log.Warn(irh.ctx, err)
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	// read uploaded image
	var file *multipart.FileHeader
	var reader multipart.File
	if err == nil {
		file, err = c.FormFile(FormKeyUploadFile)
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
	// resizing image
	var convertedimagebyte []byte
	if err == nil && reflect.DeepEqual(imageoption, actor.ImageOperatorOption{}) == false {
		contenttype := utils.GetContentTypeByReadSeeker(reader.(io.ReadSeeker))
		imageOperator := actor.NewImageOperator(
			irh.ctx,
			contenttype,
			imageoption,
		)
		reader.Seek(0, os.SEEK_SET)
		err = imageOperator.Decode(reader)
		if err == nil {
			err = imageOperator.Process()
		}
		if err == nil {
			convertedimagebyte, err = imageOperator.ImageByte()
		}
	}
	if err != nil {
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	return irh.uploadToStorage(c, reader, convertedimagebyte)
}

// UploadFile is to upload non image file to storage
func (irh ImageReductionHandler) UploadFile(c echo.Context) error {
	xRequestID := requestid.GetRequestID(c.Request())
	irh.ctx = context.WithValue(context.Background(), echo.HeaderXRequestID, xRequestID)
	log.Info(irh.ctx, "========= START REQUEST : "+c.Request().URL.RequestURI())
	log.Info(irh.ctx, c.Request().Method)
	log.Info(irh.ctx, c.Request().Header)
	file, err := c.FormFile(FormKeyUploadFile)
	if err != nil {
		log.Error(irh.ctx, err)
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	reader, err := file.Open()
	if err != nil {
		return irh.errorResponse(c, http.StatusBadRequest, err)
	}
	defer reader.Close()
	return irh.uploadToStorage(c, reader, nil)
}

func (irh ImageReductionHandler) validateUploadedImage(reader multipart.File) error {
	imagetypearray := strings.Split(os.Getenv("VALIDATE_IMAGE_TYPE"), ",")
	maxwidth := utils.GetOsEnvInt("VALIDATE_IMAGE_MAXWIDTH", 5000)
	maxheight := utils.GetOsEnvInt("VALIDATE_IMAGE_MAXHEIGHT", 5000)
	maxfilesize := utils.GetOsEnvInt("VALIDATE_IMAGE_MAXFILESIZE", 104857600)
	imagevalidate := validator.NewImageValidator(irh.ctx, imagetypearray, maxwidth, maxheight, maxfilesize)
	return imagevalidate.Validate(reader)
}

func (irh ImageReductionHandler) uploadToStorage(c echo.Context, reader multipart.File, imagebyte []byte) error {
	cloudstorageassessor := storageservice.NewCloudStorageAssessor(irh.ctx)
	if imagebyte != nil {
		return cloudstorageassessor.Put(c.FormValue(FormKeyPath), bytes.NewReader(imagebyte))
	}
	reader.Seek(0, os.SEEK_SET)
	return cloudstorageassessor.Put(c.FormValue(FormKeyPath), reader.(io.ReadSeeker))
}

func (irh ImageReductionHandler) getCache(c echo.Context, requesturi string) bool {
	cacheAssessor := cacheservice.NewCacheAssessor(irh.ctx, cacheservice.GetCachedDB())
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
	return false
}

func (irh ImageReductionHandler) setCache(mimetype string, data []byte, requesturi string) {
	cachedresponse := actor.NewCachedContentOperator()
	cachedresponse.Set(mimetype, irh.setNewLatsModified(), data)
	encodedcached, err := cachedresponse.GobEncode()
	if err != nil {
		log.Error(irh.ctx, err)
	} else {
		cacheAssessor := cacheservice.NewCacheAssessor(irh.ctx, cacheservice.GetCachedDB())
		cacheAssessor.Set(requesturi, encodedcached, cacheservice.GetChacheExpired())
	}
}

func (irh ImageReductionHandler) getImageOptionByFormValue(c echo.Context) (actor.ImageOperatorOption, error) {
	var err error
	option := actor.ImageOperatorOption{}
	option.Rotate = c.FormValue(FormKeyRotate)
	option.Width, err = irh.setOptionValueInt(c.FormValue(FormKeyWidth), err)
	option.Height, err = irh.setOptionValueInt(c.FormValue(FormKeyHeight), err)
	option.Quality, err = irh.setOptionValueInt(c.FormValue(FormKeyQuality), err)
	option.Brightness, err = irh.setOptionValueInt(c.FormValue(FormKeyBrightness), err)
	option.Contrast, err = irh.setOptionValueInt(c.FormValue(FormKeyContrast), err)
	option.Gamma, err = irh.setOptionValueFloat(c.FormValue(FormKeyGamma), err)
	option.Crop, err = irh.getCropParam(c.FormValue(FormKeyCrop), err)
	return option, err
}

func (irh ImageReductionHandler) setOptionValueInt(formvalue string, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	if formvalue == "" {
		return 0, err
	}
	val, err := strconv.Atoi(formvalue)
	if err != nil {
		log.Warn(irh.ctx, err)
		err = errors.New("Invalid parameter")
	}
	return val, err
}

func (irh ImageReductionHandler) setOptionValueFloat(formvalue string, err error) (float64, error) {
	if err != nil {
		return 0, err
	}
	if formvalue == "" {
		return 0, err
	}
	val, err := strconv.ParseFloat(formvalue, 64)
	if err != nil {
		log.Warn(irh.ctx, err)
		err = errors.New("Invalid parameter")
	}
	return val, err
}

func (irh ImageReductionHandler) getCropParam(cropparam string, err error) ([4]int, error) {
	if err != nil {
		return [4]int{}, err
	}
	if cropparam == "" {
		return [4]int{}, nil
	}
	crops := strings.Split(cropparam, ",")
	if len(crops) != 4 {
		return [4]int{}, fmt.Errorf("crop parameters must need four with comma like : 111,222,333,444")
	}
	intslicecrops := make([]int, 0)
	for _, crop := range crops {
		intcrop, err := strconv.Atoi(crop)
		if err != nil {
			log.Warn(irh.ctx, err)
			err = errors.New("Invalid crop parameter")
			return [4]int{}, err
		}
		intslicecrops = append(intslicecrops, intcrop)
	}
	var intcrops [4]int
	copy(intcrops[:], intslicecrops[:4])
	return intcrops, nil
}
