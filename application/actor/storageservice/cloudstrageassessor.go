package storageservice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/howood/imagereductor/domain/entity"
	"github.com/howood/imagereductor/infrastructure/client/cloudstorages"
	log "github.com/howood/imagereductor/infrastructure/logger"
)

// RecordNotFoundMsg define status 404 message.
const RecordNotFoundMsg = "status code: 404"

// Sentinel errors for storage validation.
var (
	ErrInvalidStorageType = errors.New("invalid storage type")
	ErrStorageTypeEmpty   = errors.New("STORAGE_TYPE environment variable is not set")
)

// CloudStorageAssessor struct.
type CloudStorageAssessor struct {
	instance cloudstorages.StorageInstance
}

// NewCloudStorageAssessor creates a new CloudStorageAssessor.
//
// Deprecated: Use NewCloudStorageAssessorWithConfig for better error handling.
func NewCloudStorageAssessor() *CloudStorageAssessor {
	ctx := context.Background()
	assessor, err := NewCloudStorageAssessorWithConfig(ctx)
	if err != nil {
		panic(err)
	}
	return assessor
}

// NewCloudStorageAssessorWithConfig creates a new CloudStorageAssessor with proper error handling.
func NewCloudStorageAssessorWithConfig(ctx context.Context) (*CloudStorageAssessor, error) {
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		return nil, ErrStorageTypeEmpty
	}
	log.Debug(ctx, "use:"+storageType)

	switch storageType {
	case "s3":
		s3cfg := cloudstorages.LoadS3ConfigFromEnv()
		inst, err := cloudstorages.NewS3WithConfig(ctx, s3cfg)
		if err != nil {
			return nil, fmt.Errorf("create s3 instance: %w", err)
		}
		return &CloudStorageAssessor{instance: inst}, nil
	case "gcs":
		gcscfg := cloudstorages.LoadGCSConfigFromEnv()
		inst, err := cloudstorages.NewGCSWithConfig(ctx, gcscfg)
		if err != nil {
			return nil, fmt.Errorf("create gcs instance: %w", err)
		}
		return &CloudStorageAssessor{instance: inst}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidStorageType, storageType)
	}
}

// Get returns storage contents.
func (csa *CloudStorageAssessor) Get(ctx context.Context, key string) (string, []byte, error) {
	return csa.instance.Get(ctx, csa.instance.GetBucket(), key)
}

// GetByStreaming returns storage contents by streaming.
func (csa *CloudStorageAssessor) GetByStreaming(ctx context.Context, key string) (string, io.ReadCloser, error) {
	return csa.instance.GetByStreaming(ctx, csa.instance.GetBucket(), key)
}

// GetObjectInfo returns storage contents info.
func (csa *CloudStorageAssessor) GetObjectInfo(ctx context.Context, key string) (entity.StorageObjectInfo, error) {
	return csa.instance.GetObjectInfo(ctx, csa.instance.GetBucket(), key)
}

// Put puts storage contents.
func (csa *CloudStorageAssessor) Put(ctx context.Context, path string, file io.ReadSeeker) error {
	return csa.instance.Put(ctx, csa.instance.GetBucket(), path, file)
}

// Delete remove storage contents.
func (csa *CloudStorageAssessor) Delete(ctx context.Context, key string) error {
	return csa.instance.Delete(ctx, csa.instance.GetBucket(), key)
}
