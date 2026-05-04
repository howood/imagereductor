package actor

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/howood/imagereductor/domain/entity"
)

func newInternalTestPNG(t *testing.T, w, h int) *bytes.Reader {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}
	return bytes.NewReader(buf.Bytes())
}

func newInternalTestJPEG(t *testing.T, w, h int) *bytes.Reader {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, nil); err != nil {
		t.Fatalf("jpeg.Encode: %v", err)
	}
	return bytes.NewReader(buf.Bytes())
}

func newImageCreator(contentType string, option ImageOperatorOption, exifOrientation int) *imageCreator {
	objectOption := entity.ImageObjectOption(option)
	return &imageCreator{
		object: &entity.ImageObject{
			ContentType: contentType,
		},
		option:          &objectOption,
		exifOrientation: exifOrientation,
	}
}

func Test_Rotate_ExifOrientation3(t *testing.T) {
	t.Parallel()

	im := newImageCreator("image/png", ImageOperatorOption{Rotate: ImageRotateExifOrientation}, 3)
	if err := im.Decode(t.Context(), newInternalTestPNG(t, 80, 40)); err != nil {
		t.Fatal(err)
	}
	if err := im.rotate(t.Context()); err != nil {
		t.Fatalf("rotate exif orientation 3: %v", err)
	}
}

func Test_Rotate_ExifOrientation6(t *testing.T) {
	t.Parallel()

	im := newImageCreator("image/png", ImageOperatorOption{Rotate: ImageRotateExifOrientation}, 6)
	if err := im.Decode(t.Context(), newInternalTestPNG(t, 80, 40)); err != nil {
		t.Fatal(err)
	}
	if err := im.rotate(t.Context()); err != nil {
		t.Fatalf("rotate exif orientation 6: %v", err)
	}
}

func Test_Rotate_ExifOrientation8(t *testing.T) {
	t.Parallel()

	im := newImageCreator("image/png", ImageOperatorOption{Rotate: ImageRotateExifOrientation}, 8)
	if err := im.Decode(t.Context(), newInternalTestPNG(t, 80, 40)); err != nil {
		t.Fatal(err)
	}
	if err := im.rotate(t.Context()); err != nil {
		t.Fatalf("rotate exif orientation 8: %v", err)
	}
}

func Test_Rotate_ExifOrientation_NoRotation(t *testing.T) {
	t.Parallel()

	// orientation=1 (normal) — no rotation branches hit
	im := newImageCreator("image/png", ImageOperatorOption{Rotate: ImageRotateExifOrientation}, 1)
	if err := im.Decode(t.Context(), newInternalTestPNG(t, 80, 40)); err != nil {
		t.Fatal(err)
	}
	if err := im.rotate(t.Context()); err != nil {
		t.Fatalf("rotate exif orientation 1: %v", err)
	}
}

func Test_DecodeExifOrientation_NoExif(t *testing.T) {
	t.Parallel()

	// Standard JPEG without EXIF data — hits the exif.Decode error path
	im := newImageCreator("image/jpeg", ImageOperatorOption{}, 0)
	reader := newInternalTestJPEG(t, 40, 40)
	im.decodeExifOrientation(t.Context(), reader)
	if im.exifOrientation != 0 {
		t.Fatalf("expected 0 exifOrientation for no-exif JPEG, got %d", im.exifOrientation)
	}
}

func Test_ImageByte_GIF(t *testing.T) {
	t.Parallel()

	im := newImageCreator("image/gif", ImageOperatorOption{}, 0)
	if err := im.Decode(t.Context(), newInternalTestPNG(t, 30, 30)); err != nil {
		t.Fatal(err)
	}
	if err := im.Process(t.Context()); err != nil {
		t.Fatal(err)
	}
	out, err := im.ImageByte(t.Context())
	if err != nil {
		t.Fatalf("ImageByte GIF: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("expected non-empty GIF output")
	}
}
