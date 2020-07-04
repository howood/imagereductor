package repository

import "io"

type ImageObjectRepository interface {
	Decode(src io.Reader) error
	Resize()
	ImageByte() ([]byte, error)
}
