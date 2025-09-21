package cloudstorages

import (
	"context"
	"io"

	"github.com/howood/imagereductor/domain/entity"
)

const mimeOctetStream = "application/octet-stream"

// StorageInstance interface.
type StorageInstance interface {
	Put(ctx context.Context, bucket string, path string, file io.ReadSeeker) error
	Get(ctx context.Context, bucket string, key string) (string, []byte, error)
	GetByStreaming(ctx context.Context, bucket string, key string) (string, io.ReadCloser, error)
	GetObjectInfo(ctx context.Context, bucket string, key string) (entity.StorageObjectInfo, error)
	List(ctx context.Context, bucket string, key string) ([]string, error)
	Delete(ctx context.Context, bucket string, key string) error
	GetBucket() string
}
