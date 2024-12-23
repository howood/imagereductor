package repository

import (
	"context"
	"io"
)

// ImageObjectRepository interface.
type ImageObjectRepository interface {
	Decode(ctx context.Context, src io.ReadSeeker) error
	Process(ctx context.Context) error
	ImageByte(ctx context.Context) ([]byte, error)
}
