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
	cloudstorage *storageservice.CloudStorageAssessor
}

// NewImageUsecase creates a new ImageUsecase.
//
// Deprecated: Use NewImageUsecaseWithConfig for better error handling.
func NewImageUsecase() *ImageUsecase {
	uc, err := NewImageUsecaseWithConfig(context.Background())
	if err != nil {
		panic(err)
	}
	return uc
}

// NewImageUsecaseWithConfig creates a new ImageUsecase with proper error handling.
func NewImageUsecaseWithConfig(ctx context.Context) (*ImageUsecase, error) {
	cloudstorage, err := storageservice.NewCloudStorageAssessorWithConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &ImageUsecase{
		cloudstorage: cloudstorage,
	}, nil
}

func (iu *ImageUsecase) GetImage(ctx context.Context, imageoption actor.ImageOperatorOption, storageKeyValue string) (string, []byte, error) {
	// get from storage
	contenttype, imagebyte, err := iu.cloudstorage.Get(ctx, storageKeyValue)
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

func (iu *ImageUsecase) GetFile(ctx context.Context, storageKeyValue string) (string, []byte, error) {
	// get from storage
	contenttype, filebyte, err := iu.cloudstorage.Get(ctx, storageKeyValue)
	return contenttype, filebyte, err
}

func (iu *ImageUsecase) GetFileStream(ctx context.Context, storageKeyValue string) (string, int, io.ReadCloser, error) {
	// get from storage
	objectInfo, err := iu.cloudstorage.GetObjectInfo(ctx, storageKeyValue)
	if err != nil {
		return "", 0, nil, err
	}
	contentLength := objectInfo.ContentLength
	contenttype, response, err := iu.cloudstorage.GetByStreaming(ctx, storageKeyValue)
	return contenttype, contentLength, response, err
}

func (iu *ImageUsecase) GetFileInfo(ctx context.Context, storageKeyValue string) (entity.StorageObjectInfo, error) {
	// get from storage
	objectInfo, err := iu.cloudstorage.GetObjectInfo(ctx, storageKeyValue)
	return objectInfo, err
}

func (iu *ImageUsecase) ConvertImage(ctx context.Context, imageoption actor.ImageOperatorOption, reader multipart.File) ([]byte, error) {
	var convertedimagebyte []byte
	var err error
	if !reflect.DeepEqual(imageoption, actor.ImageOperatorOption{}) {
		re, ok := reader.(io.ReadSeeker)
		if !ok {
			return nil, err
		}
		contenttype, _ := utils.GetContentTypeByReadSeeker(re)
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

func (iu *ImageUsecase) UploadToStorage(ctx context.Context, formKeyPath string, reader multipart.File, imagebyte []byte) error {
	if imagebyte != nil {
		return iu.cloudstorage.Put(ctx, formKeyPath, bytes.NewReader(imagebyte))
	}
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return err
	}
	//nolint:forcetypeassert
	return iu.cloudstorage.Put(ctx, formKeyPath, reader.(io.ReadSeeker))
}
