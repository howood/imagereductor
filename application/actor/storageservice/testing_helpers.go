package storageservice

import "github.com/howood/imagereductor/infrastructure/client/cloudstorages"

// NewCloudStorageAssessorForTest creates a CloudStorageAssessor with the given instance for testing.
func NewCloudStorageAssessorForTest(inst cloudstorages.StorageInstance) *CloudStorageAssessor {
	return &CloudStorageAssessor{instance: inst}
}
