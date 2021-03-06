package cloudstorages

import (
	"io"
)

// StorageInstance interface
type StorageInstance interface {
	Put(bucket string, path string, file io.ReadSeeker) error
	Get(bucket string, key string) (string, []byte, error)
	List(bucket string, key string) ([]string, error)
	Delete(bucket string, key string) error
}
