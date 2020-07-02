package cloudstorages

import (
	"io"
)

type StorageInstance interface {
	Put(bucket string, path string, file io.ReadSeeker) error
	Get(bucket string, key string) (string, []byte, error)
	Delete(bucket string, key string) error
}
