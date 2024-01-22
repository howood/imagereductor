package repository

import "io"

// ImageObjectRepository interface
type ImageObjectRepository interface {
	Decode(src io.ReadSeeker) error
	Process() error
	ImageByte() ([]byte, error)
}
