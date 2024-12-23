package storageservice

import (
	"context"
	"io"
	"os"

	"github.com/howood/imagereductor/domain/entity"
	"github.com/howood/imagereductor/infrastructure/client/cloudstorages"
	log "github.com/howood/imagereductor/infrastructure/logger"
)

// RecordNotFoundMsg define status 404 message.
const RecordNotFoundMsg = "status code: 404"

// CloudStorageAssessor struct.
type CloudStorageAssessor struct {
	instance cloudstorages.StorageInstance
	bucket   string
}

// NewCloudStorageAssessor creates a new CloudStorageAssessor.
func NewCloudStorageAssessor(ctx context.Context) *CloudStorageAssessor {
	var I *CloudStorageAssessor
	log.Debug(ctx, "use:"+os.Getenv("STORAGE_TYPE"))
	switch os.Getenv("STORAGE_TYPE") {
	case "s3":
		I = &CloudStorageAssessor{
			instance: cloudstorages.NewS3(ctx),
			bucket:   cloudstorages.S3BucketUploadfiles,
		}
	case "gcs":
		I = &CloudStorageAssessor{
			instance: cloudstorages.NewGCS(ctx),
			bucket:   cloudstorages.GcsBucketUploadfiles,
		}
	default:
		panic("Invalid STORAGE_TYPE")
	}
	return I
}

// Get returns storage contents.
func (csa *CloudStorageAssessor) Get(ctx context.Context, key string) (string, []byte, error) {
	return csa.instance.Get(ctx, csa.bucket, key)
}

// GetByStreaming returns storage contents by streaming.
func (csa *CloudStorageAssessor) GetByStreaming(ctx context.Context, key string) (string, io.ReadCloser, error) {
	return csa.instance.GetByStreaming(ctx, csa.bucket, key)
}

// GetObjectInfo returns storage contents info.
func (csa *CloudStorageAssessor) GetObjectInfo(ctx context.Context, key string) (entity.StorageObjectInfo, error) {
	return csa.instance.GetObjectInfo(ctx, csa.bucket, key)
}

// Put puts storage contents.
func (csa *CloudStorageAssessor) Put(ctx context.Context, path string, file io.ReadSeeker) error {
	return csa.instance.Put(ctx, csa.bucket, path, file)
}

// Delete remove storage contents.
func (csa *CloudStorageAssessor) Delete(ctx context.Context, key string) error {
	return csa.instance.Delete(ctx, csa.bucket, key)
}
