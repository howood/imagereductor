package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/application/actor/storageservice"
	"github.com/howood/imagereductor/domain/entity"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/howood/imagereductor/interfaces/service/config"
	"github.com/howood/imagereductor/library/utils"
	"github.com/labstack/echo/v4"
)

type ImageUsecase struct {
	Ctx context.Context
}

func (iu ImageUsecase) GetImage(c echo.Context, storageKeyValue string) (contenttype string, imagebyte []byte, err error) {
	// get imageoption
	imageoption, err := iu.GetImageOptionByFormValue(c)
	// get from storage
	if err == nil {
		cloudstorageassessor := storageservice.NewCloudStorageAssessor(iu.Ctx)
		contenttype, imagebyte, err = cloudstorageassessor.Get(storageKeyValue)
	}
	// resizing image
	if err == nil && reflect.DeepEqual(imageoption, actor.ImageOperatorOption{}) == false {
		imageOperator := actor.NewImageOperator(
			iu.Ctx,
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
	return contenttype, imagebyte, err
}

func (iu ImageUsecase) GetFile(storageKeyValue string) (contenttype string, filebyte []byte, err error) {
	// get from storage
	cloudstorageassessor := storageservice.NewCloudStorageAssessor(iu.Ctx)
	contenttype, filebyte, err = cloudstorageassessor.Get(storageKeyValue)
	return contenttype, filebyte, err
}

func (iu ImageUsecase) GetFileStream(storageKeyValue string) (contenttype string, contentLength int, filebyte io.ReadCloser, err error) {
	// get from storage
	cloudstorageassessor := storageservice.NewCloudStorageAssessor(iu.Ctx)
	objectInfo, err := cloudstorageassessor.GetObjectInfo(storageKeyValue)
	if err != nil {
		return "", 0, nil, err
	}
	contentLength = objectInfo.ContentLength
	contenttype, response, err := cloudstorageassessor.GetByStreaming(storageKeyValue)
	return contenttype, contentLength, response, err
}

func (iu ImageUsecase) GetFileInfo(storageKeyValue string) (objectInfo entity.StorageObjectInfo, err error) {
	// get from storage
	cloudstorageassessor := storageservice.NewCloudStorageAssessor(iu.Ctx)
	objectInfo, err = cloudstorageassessor.GetObjectInfo(storageKeyValue)
	return objectInfo, err
}

func (iu ImageUsecase) ConvertImage(imageoption actor.ImageOperatorOption, reader multipart.File) (convertedimagebyte []byte, err error) {
	if reflect.DeepEqual(imageoption, actor.ImageOperatorOption{}) == false {
		contenttype := utils.GetContentTypeByReadSeeker(reader.(io.ReadSeeker))
		imageOperator := actor.NewImageOperator(
			iu.Ctx,
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
	return convertedimagebyte, err
}

func (iu ImageUsecase) UploadToStorage(FormKeyPath string, reader multipart.File, imagebyte []byte) error {
	cloudstorageassessor := storageservice.NewCloudStorageAssessor(iu.Ctx)
	if imagebyte != nil {
		return cloudstorageassessor.Put(FormKeyPath, bytes.NewReader(imagebyte))
	}
	reader.Seek(0, os.SEEK_SET)
	return cloudstorageassessor.Put(FormKeyPath, reader.(io.ReadSeeker))
}

func (iu ImageUsecase) GetImageOptionByFormValue(c echo.Context) (actor.ImageOperatorOption, error) {
	var err error
	option := actor.ImageOperatorOption{}
	option.Rotate = c.FormValue(config.FormKeyRotate)
	option.Width, err = iu.setOptionValueInt(c.FormValue(config.FormKeyWidth), err)
	option.Height, err = iu.setOptionValueInt(c.FormValue(config.FormKeyHeight), err)
	option.Quality, err = iu.setOptionValueInt(c.FormValue(config.FormKeyQuality), err)
	option.Brightness, err = iu.setOptionValueInt(c.FormValue(config.FormKeyBrightness), err)
	option.Contrast, err = iu.setOptionValueInt(c.FormValue(config.FormKeyContrast), err)
	option.Gamma, err = iu.setOptionValueFloat(c.FormValue(config.FormKeyGamma), err)
	option.Crop, err = iu.getCropParam(c.FormValue(config.FormKeyCrop), err)
	return option, err
}

func (iu ImageUsecase) setOptionValueInt(formvalue string, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	if formvalue == "" {
		return 0, err
	}
	val, err := strconv.Atoi(formvalue)
	if err != nil {
		log.Warn(iu.Ctx, err)
		err = errors.New("Invalid parameter")
	}
	return val, err
}

func (iu ImageUsecase) setOptionValueFloat(formvalue string, err error) (float64, error) {
	if err != nil {
		return 0, err
	}
	if formvalue == "" {
		return 0, err
	}
	val, err := strconv.ParseFloat(formvalue, 64)
	if err != nil {
		log.Warn(iu.Ctx, err)
		err = errors.New("Invalid parameter")
	}
	return val, err
}

func (irh ImageUsecase) getCropParam(cropparam string, err error) ([4]int, error) {
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
			log.Warn(irh.Ctx, err)
			err = errors.New("Invalid crop parameter")
			return [4]int{}, err
		}
		intslicecrops = append(intslicecrops, intcrop)
	}
	var intcrops [4]int
	copy(intcrops[:], intslicecrops[:4])
	return intcrops, nil
}
