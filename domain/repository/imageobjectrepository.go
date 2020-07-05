package repository

import "io"

// ImageObjectRepository interface
type ImageObjectRepository interface {
	Decode(src io.Reader) error
	Process() error
	ImageByte() ([]byte, error)
}
