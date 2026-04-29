package validator_test

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"

	"github.com/howood/imagereductor/application/validator"
)

func newPNGBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}
	return buf.Bytes()
}

func newJPEGBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, nil); err != nil {
		t.Fatalf("jpeg.Encode: %v", err)
	}
	return buf.Bytes()
}

func Test_ImageValidator_Valid(t *testing.T) {
	t.Parallel()

	v := validator.NewImageValidator([]string{validator.ImageTypePng, validator.ImageTypeJpeg}, 100, 100, 1024*1024)
	data := newPNGBytes(t, 50, 50)
	if err := v.Validate(t.Context(), bytes.NewReader(data)); err != nil {
		t.Fatalf("expected valid image, got error: %v", err)
	}
}

func Test_ImageValidator_NoLimits(t *testing.T) {
	t.Parallel()

	v := validator.NewImageValidator([]string{validator.ImageTypePng}, 0, 0, 0)
	data := newPNGBytes(t, 200, 200)
	if err := v.Validate(t.Context(), bytes.NewReader(data)); err != nil {
		t.Fatalf("expected valid image with no size limits, got error: %v", err)
	}
}

func Test_ImageValidator_InvalidType(t *testing.T) {
	t.Parallel()

	v := validator.NewImageValidator([]string{validator.ImageTypePng}, 0, 0, 0)
	data := newJPEGBytes(t, 50, 50)
	err := v.Validate(t.Context(), bytes.NewReader(data))
	if err == nil {
		t.Fatal("expected invalid type error, got nil")
	}
	if !errors.Is(err, validator.ErrInvalidImageType) {
		t.Fatalf("expected ErrInvalidImageType, got %v", err)
	}
}

func Test_ImageValidator_DecodeFailure(t *testing.T) {
	t.Parallel()

	v := validator.NewImageValidator([]string{validator.ImageTypePng}, 0, 0, 0)
	err := v.Validate(t.Context(), strings.NewReader("not an image"))
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
	if !errors.Is(err, validator.ErrImageDecodeConfig) {
		t.Fatalf("expected ErrImageDecodeConfig, got %v", err)
	}
}

func Test_ImageValidator_ExceedsWidth(t *testing.T) {
	t.Parallel()

	v := validator.NewImageValidator([]string{validator.ImageTypePng}, 10, 0, 0)
	data := newPNGBytes(t, 50, 50)
	err := v.Validate(t.Context(), bytes.NewReader(data))
	if err == nil || !errors.Is(err, validator.ErrImageSizeExceeded) {
		t.Fatalf("expected ErrImageSizeExceeded for width, got %v", err)
	}
}

func Test_ImageValidator_ExceedsHeight(t *testing.T) {
	t.Parallel()

	v := validator.NewImageValidator([]string{validator.ImageTypePng}, 0, 10, 0)
	data := newPNGBytes(t, 50, 50)
	err := v.Validate(t.Context(), bytes.NewReader(data))
	if err == nil || !errors.Is(err, validator.ErrImageSizeExceeded) {
		t.Fatalf("expected ErrImageSizeExceeded for height, got %v", err)
	}
}

func Test_ImageValidator_ExceedsFileSize(t *testing.T) {
	t.Parallel()

	v := validator.NewImageValidator([]string{validator.ImageTypePng}, 0, 0, 10)
	data := newPNGBytes(t, 50, 50)
	err := v.Validate(t.Context(), bytes.NewReader(data))
	if err == nil || !errors.Is(err, validator.ErrImageSizeExceeded) {
		t.Fatalf("expected ErrImageSizeExceeded for filesize, got %v", err)
	}
}

func Test_ImageValidator_FiltersUnknownTypes(t *testing.T) {
	t.Parallel()

	// "unknown" should be filtered out, leaving only PNG; PNG image should still validate.
	v := validator.NewImageValidator([]string{"unknown", validator.ImageTypePng}, 0, 0, 0)
	data := newPNGBytes(t, 10, 10)
	if err := v.Validate(t.Context(), bytes.NewReader(data)); err != nil {
		t.Fatalf("expected valid image, got error: %v", err)
	}
}
