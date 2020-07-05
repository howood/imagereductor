package repository

import "io"

// ImageObjectRepository interface
type ImageObjectRepository interface {
	Decode(src io.Reader) error
	Resize()
	ImageByte() ([]byte, error)
}
