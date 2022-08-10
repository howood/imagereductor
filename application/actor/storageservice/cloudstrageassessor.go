package storageservice

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/howood/imagereductor/domain/entity"
	"github.com/howood/imagereductor/infrastructure/client/cloudstorages"
	log "github.com/howood/imagereductor/infrastructure/logger"
)

// RecordNotFoundMsg define status 404 message
const RecordNotFoundMsg = "status code: 404"

// CloudStorageAssessor struct
type CloudStorageAssessor struct {
	instance cloudstorages.StorageInstance
	bucket   string
	ctx      context.Context
}

// NewCloudStorageAssessor creates a new CloudStorageAssessor
func NewCloudStorageAssessor(ctx context.Context) *CloudStorageAssessor {
	var I *CloudStorageAssessor
	log.Debug(ctx, "use:"+os.Getenv("STORAGE_TYPE"))
	switch os.Getenv("STORAGE_TYPE") {
	case "s3":
		I = &CloudStorageAssessor{
			instance: cloudstorages.NewS3(ctx),
			bucket:   cloudstorages.S3BucketUploadfiles,
			ctx:      ctx,
		}
	case "gcs":
		I = &CloudStorageAssessor{
			instance: cloudstorages.NewGCS(ctx),
			bucket:   cloudstorages.GcsBucketUploadfiles,
			ctx:      ctx,
		}
	default:
		panic(errors.New("Invalid STORAGE_TYPE"))
	}
	return I
}

// Get returns storage contents
func (csa *CloudStorageAssessor) Get(key string) (string, []byte, error) {
	return csa.instance.Get(csa.bucket, key)
}

// GetByStreaming returns storage contents by streaming
func (csa *CloudStorageAssessor) GetByStreaming(key string) (string, io.ReadCloser, error) {
	return csa.instance.GetByStreaming(csa.bucket, key)
}

// GetObjectInfo returns storage contents info
func (csa *CloudStorageAssessor) GetObjectInfo(key string) (entity.StorageObjectInfo, error) {
	return csa.instance.GetObjectInfo(csa.bucket, key)
}

// Put puts storage contents
func (csa *CloudStorageAssessor) Put(path string, file io.ReadSeeker) error {
	return csa.instance.Put(csa.bucket, path, file)
}

// Delete remove storage contents
func (csa *CloudStorageAssessor) Delete(key string) error {
	return csa.instance.Delete(csa.bucket, key)
}
