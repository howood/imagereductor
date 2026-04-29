package handler

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/textproto"
	"testing"
)

// fakeMultipartFile wraps bytes.Reader to implement multipart.File.
type fakeMultipartFile struct {
	*bytes.Reader
}

func (f *fakeMultipartFile) Close() error { return nil }

func newPNGMultipart(t *testing.T, w, h int) multipart.File {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 1, G: 2, B: 3, A: 255})
		}
	}
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}
	return &fakeMultipartFile{Reader: bytes.NewReader(buf.Bytes())}
}

func Test_validateUploadedImage_OK(t *testing.T) {
	t.Setenv("VALIDATE_IMAGE_TYPE", "png,jpeg")
	t.Setenv("VALIDATE_IMAGE_MAXWIDTH", "1000")
	t.Setenv("VALIDATE_IMAGE_MAXHEIGHT", "1000")
	t.Setenv("VALIDATE_IMAGE_MAXFILESIZE", "1000000")

	h := &ImageReductionHandler{}
	r := newPNGMultipart(t, 100, 100)
	if err := h.validateUploadedImage(context.Background(), r); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func Test_validateUploadedImage_ExceedsWidth(t *testing.T) {
	t.Setenv("VALIDATE_IMAGE_TYPE", "png")
	t.Setenv("VALIDATE_IMAGE_MAXWIDTH", "10")
	t.Setenv("VALIDATE_IMAGE_MAXHEIGHT", "10000")
	t.Setenv("VALIDATE_IMAGE_MAXFILESIZE", "10000000")

	h := &ImageReductionHandler{}
	r := newPNGMultipart(t, 100, 100)
	if err := h.validateUploadedImage(context.Background(), r); err == nil {
		t.Fatal("expected error for width over limit")
	}
}

// silence unused import
var _ = textproto.CanonicalMIMEHeaderKey
