package usecase

import "github.com/howood/imagereductor/application/actor/storageservice"

// NewImageUsecaseForTest creates an ImageUsecase with the given CloudStorageAssessor for testing.
func NewImageUsecaseForTest(csa *storageservice.CloudStorageAssessor) *ImageUsecase {
	return &ImageUsecase{cloudstorage: csa}
}
