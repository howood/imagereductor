package usecase

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"reflect"

	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/application/actor/storageservice"
	"github.com/howood/imagereductor/domain/entity"
	"github.com/howood/imagereductor/library/utils"
)

type ImageUsecase struct {
	Ctx context.Context
}

func (iu ImageUsecase) GetImage(imageoption actor.ImageOperatorOption, storageKeyValue string) (contenttype string, imagebyte []byte, err error) {
	// get from storage
	cloudstorageassessor := storageservice.NewCloudStorageAssessor(iu.Ctx)
	contenttype, imagebyte, err = cloudstorageassessor.Get(storageKeyValue)
	if err != nil {
		return contenttype, imagebyte, err
	}
	// resizing image
	if !reflect.DeepEqual(imageoption, actor.ImageOperatorOption{}) {
		imageOperator := actor.NewImageOperator(
			iu.Ctx,
			contenttype,
			imageoption,
		)
		err = imageOperator.Decode(bytes.NewReader(imagebyte))
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
	if !reflect.DeepEqual(imageoption, actor.ImageOperatorOption{}) {
		contenttype, _ := utils.GetContentTypeByReadSeeker(reader.(io.ReadSeeker))
		imageOperator := actor.NewImageOperator(
			iu.Ctx,
			contenttype,
			imageoption,
		)
		_, err = reader.Seek(0, io.SeekStart)
		if err == nil {
			err = imageOperator.Decode(reader)
		}
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
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return cloudstorageassessor.Put(FormKeyPath, reader.(io.ReadSeeker))
}
