package cloudstorages

import (
	"io"

	"github.com/howood/imagereductor/domain/entity"
)

const mimeOctetStream = "application/octet-stream"

// StorageInstance interface
type StorageInstance interface {
	Put(bucket string, path string, file io.ReadSeeker) error
	Get(bucket string, key string) (string, []byte, error)
	GetByStreaming(bucket string, key string) (string, io.ReadCloser, error)
	GetObjectInfo(bucket string, key string) (entity.StorageObjectInfo, error)
	List(bucket string, key string) ([]string, error)
	Delete(bucket string, key string) error
}
