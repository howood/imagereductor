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

type ImageUsecase struct{}

//nolint:nolintlint,ireturn
func (iu ImageUsecase) GetImage(ctx context.Context, imageoption actor.ImageOperatorOption, storageKeyValue string) (string, []byte, error) {
	// get from storage
	cloudstorageassessor := storageservice.NewCloudStorageAssessor(ctx)
	contenttype, imagebyte, err := cloudstorageassessor.Get(ctx, storageKeyValue)
	if err != nil {
		return contenttype, imagebyte, err
	}
	// resizing image
	if !reflect.DeepEqual(imageoption, actor.ImageOperatorOption{}) {
		imageOperator := actor.NewImageOperator(
			contenttype,
			imageoption,
		)
		err = imageOperator.Decode(ctx, bytes.NewReader(imagebyte))
		if err == nil {
			err = imageOperator.Process(ctx)
		}
		if err == nil {
			imagebyte, err = imageOperator.ImageByte(ctx)
		}
	}
	return contenttype, imagebyte, err
}

func (iu ImageUsecase) GetFile(ctx context.Context, storageKeyValue string) (string, []byte, error) {
	// get from storage
	cloudstorageassessor := storageservice.NewCloudStorageAssessor(ctx)
	contenttype, filebyte, err := cloudstorageassessor.Get(ctx, storageKeyValue)
	return contenttype, filebyte, err
}

func (iu ImageUsecase) GetFileStream(ctx context.Context, storageKeyValue string) (string, int, io.ReadCloser, error) {
	// get from storage
	cloudstorageassessor := storageservice.NewCloudStorageAssessor(ctx)
	objectInfo, err := cloudstorageassessor.GetObjectInfo(ctx, storageKeyValue)
	if err != nil {
		return "", 0, nil, err
	}
	contentLength := objectInfo.ContentLength
	contenttype, response, err := cloudstorageassessor.GetByStreaming(ctx, storageKeyValue)
	return contenttype, contentLength, response, err
}

func (iu ImageUsecase) GetFileInfo(ctx context.Context, storageKeyValue string) (entity.StorageObjectInfo, error) {
	// get from storage
	cloudstorageassessor := storageservice.NewCloudStorageAssessor(ctx)
	objectInfo, err := cloudstorageassessor.GetObjectInfo(ctx, storageKeyValue)
	return objectInfo, err
}

func (iu ImageUsecase) ConvertImage(ctx context.Context, imageoption actor.ImageOperatorOption, reader multipart.File) ([]byte, error) {
	var convertedimagebyte []byte
	var err error
	if !reflect.DeepEqual(imageoption, actor.ImageOperatorOption{}) {
		contenttype, _ := utils.GetContentTypeByReadSeeker(reader.(io.ReadSeeker))
		imageOperator := actor.NewImageOperator(
			contenttype,
			imageoption,
		)
		_, err = reader.Seek(0, io.SeekStart)
		if err == nil {
			err = imageOperator.Decode(ctx, reader)
		}
		if err == nil {
			err = imageOperator.Process(ctx)
		}
		if err == nil {
			convertedimagebyte, err = imageOperator.ImageByte(ctx)
		}
	}
	return convertedimagebyte, err
}

func (iu ImageUsecase) UploadToStorage(ctx context.Context, formKeyPath string, reader multipart.File, imagebyte []byte) error {
	cloudstorageassessor := storageservice.NewCloudStorageAssessor(ctx)
	if imagebyte != nil {
		return cloudstorageassessor.Put(ctx, formKeyPath, bytes.NewReader(imagebyte))
	}
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return err
	}
	//nolint:forcetypeassert
	return cloudstorageassessor.Put(ctx, formKeyPath, reader.(io.ReadSeeker))
}
