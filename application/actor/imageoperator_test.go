package actor_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"

	"github.com/howood/imagereductor/application/actor"
)

func newTestPNGReader(t *testing.T, w, h int) *bytes.Reader {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}
	return bytes.NewReader(buf.Bytes())
}

func newTestJPEGReader(t *testing.T, w, h int) *bytes.Reader {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, nil); err != nil {
		t.Fatalf("jpeg.Encode: %v", err)
	}
	return bytes.NewReader(buf.Bytes())
}

func Test_ImageOperator_DecodeAndProcess_PNG(t *testing.T) {
	t.Parallel()

	op := actor.NewImageOperator("image/png", actor.ImageOperatorOption{
		Width:  50,
		Height: 50,
	})
	if err := op.Decode(t.Context(), newTestPNGReader(t, 100, 100)); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if err := op.Process(t.Context()); err != nil {
		t.Fatalf("Process failed: %v", err)
	}
	out, err := op.ImageByte(t.Context())
	if err != nil {
		t.Fatalf("ImageByte failed: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("ImageByte returned empty bytes")
	}
	// verify the output is a valid PNG
	if _, _, err := image.DecodeConfig(bytes.NewReader(out)); err != nil {
		t.Fatalf("output is not a valid image: %v", err)
	}
}

func Test_ImageOperator_DecodeAndProcess_JPEG_QualityVariants(t *testing.T) {
	t.Parallel()

	for _, quality := range []int{1, 2, 3, 4, 99} {
		quality := quality
		t.Run("quality_"+itoa(quality), func(t *testing.T) {
			t.Parallel()
			op := actor.NewImageOperator("image/jpeg", actor.ImageOperatorOption{
				Width:   30,
				Height:  30,
				Quality: quality,
			})
			if err := op.Decode(t.Context(), newTestJPEGReader(t, 60, 60)); err != nil {
				t.Fatalf("Decode failed: %v", err)
			}
			if err := op.Process(t.Context()); err != nil {
				t.Fatalf("Process failed: %v", err)
			}
			if _, err := op.ImageByte(t.Context()); err != nil {
				t.Fatalf("ImageByte failed: %v", err)
			}
		})
	}
}

func Test_ImageOperator_NoResize(t *testing.T) {
	t.Parallel()

	op := actor.NewImageOperator("image/png", actor.ImageOperatorOption{})
	if err := op.Decode(t.Context(), newTestPNGReader(t, 80, 40)); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if err := op.Process(t.Context()); err != nil {
		t.Fatalf("Process failed: %v", err)
	}
	out, err := op.ImageByte(t.Context())
	if err != nil {
		t.Fatalf("ImageByte failed: %v", err)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if cfg.Width != 80 || cfg.Height != 40 {
		t.Fatalf("expected 80x40 unchanged, got %dx%d", cfg.Width, cfg.Height)
	}
}

func Test_ImageOperator_WidthOnly(t *testing.T) {
	t.Parallel()

	op := actor.NewImageOperator("image/png", actor.ImageOperatorOption{Width: 40})
	if err := op.Decode(t.Context(), newTestPNGReader(t, 80, 40)); err != nil {
		t.Fatal(err)
	}
	if err := op.Process(t.Context()); err != nil {
		t.Fatal(err)
	}
	out, _ := op.ImageByte(t.Context())
	cfg, _, _ := image.DecodeConfig(bytes.NewReader(out))
	if cfg.Width != 40 {
		t.Fatalf("expected width 40, got %d", cfg.Width)
	}
}

func Test_ImageOperator_HeightOnly(t *testing.T) {
	t.Parallel()

	op := actor.NewImageOperator("image/png", actor.ImageOperatorOption{Height: 20})
	if err := op.Decode(t.Context(), newTestPNGReader(t, 80, 40)); err != nil {
		t.Fatal(err)
	}
	if err := op.Process(t.Context()); err != nil {
		t.Fatal(err)
	}
	out, _ := op.ImageByte(t.Context())
	cfg, _, _ := image.DecodeConfig(bytes.NewReader(out))
	if cfg.Height != 20 {
		t.Fatalf("expected height 20, got %d", cfg.Height)
	}
}

func Test_ImageOperator_Crop(t *testing.T) {
	t.Parallel()

	op := actor.NewImageOperator("image/png", actor.ImageOperatorOption{
		Crop: [4]int{10, 10, 60, 60},
	})
	if err := op.Decode(t.Context(), newTestPNGReader(t, 100, 100)); err != nil {
		t.Fatal(err)
	}
	if err := op.Process(t.Context()); err != nil {
		t.Fatal(err)
	}
	out, err := op.ImageByte(t.Context())
	if err != nil {
		t.Fatalf("ImageByte: %v", err)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if cfg.Width != 50 || cfg.Height != 50 {
		t.Fatalf("expected 50x50 cropped, got %dx%d", cfg.Width, cfg.Height)
	}
}

func Test_ImageOperator_RotateRight(t *testing.T) {
	t.Parallel()

	op := actor.NewImageOperator("image/png", actor.ImageOperatorOption{
		Rotate: actor.ImageRotateRight,
	})
	if err := op.Decode(t.Context(), newTestPNGReader(t, 80, 40)); err != nil {
		t.Fatal(err)
	}
	if err := op.Process(t.Context()); err != nil {
		t.Fatalf("Process: %v", err)
	}
	out, err := op.ImageByte(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	cfg, _, _ := image.DecodeConfig(bytes.NewReader(out))
	if cfg.Width != 40 || cfg.Height != 80 {
		t.Fatalf("rotate right: expected 40x80, got %dx%d", cfg.Width, cfg.Height)
	}
}

func Test_ImageOperator_RotateLeft(t *testing.T) {
	t.Parallel()

	op := actor.NewImageOperator("image/png", actor.ImageOperatorOption{Rotate: actor.ImageRotateLeft})
	if err := op.Decode(t.Context(), newTestPNGReader(t, 80, 40)); err != nil {
		t.Fatal(err)
	}
	if err := op.Process(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func Test_ImageOperator_RotateUpsidedown(t *testing.T) {
	t.Parallel()

	op := actor.NewImageOperator("image/png", actor.ImageOperatorOption{Rotate: actor.ImageRotateUpsidedown})
	if err := op.Decode(t.Context(), newTestPNGReader(t, 60, 30)); err != nil {
		t.Fatal(err)
	}
	if err := op.Process(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func Test_ImageOperator_RotateAutoVertical(t *testing.T) {
	t.Parallel()

	// horizontal -> rotated to vertical
	op := actor.NewImageOperator("image/png", actor.ImageOperatorOption{Rotate: actor.ImageRotateAutoVertical})
	if err := op.Decode(t.Context(), newTestPNGReader(t, 80, 40)); err != nil {
		t.Fatal(err)
	}
	if err := op.Process(t.Context()); err != nil {
		t.Fatal(err)
	}

	// already vertical -> unchanged
	op2 := actor.NewImageOperator("image/png", actor.ImageOperatorOption{Rotate: actor.ImageRotateAutoVertical})
	if err := op2.Decode(t.Context(), newTestPNGReader(t, 40, 80)); err != nil {
		t.Fatal(err)
	}
	if err := op2.Process(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func Test_ImageOperator_RotateAutoHorizontal(t *testing.T) {
	t.Parallel()

	op := actor.NewImageOperator("image/png", actor.ImageOperatorOption{Rotate: actor.ImageRotateAutoHorizontal})
	if err := op.Decode(t.Context(), newTestPNGReader(t, 40, 80)); err != nil {
		t.Fatal(err)
	}
	if err := op.Process(t.Context()); err != nil {
		t.Fatal(err)
	}

	op2 := actor.NewImageOperator("image/png", actor.ImageOperatorOption{Rotate: actor.ImageRotateAutoHorizontal})
	if err := op2.Decode(t.Context(), newTestPNGReader(t, 80, 40)); err != nil {
		t.Fatal(err)
	}
	if err := op2.Process(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func Test_ImageOperator_Rotate_Invalid(t *testing.T) {
	t.Parallel()

	op := actor.NewImageOperator("image/png", actor.ImageOperatorOption{Rotate: "invalid_value"})
	if err := op.Decode(t.Context(), newTestPNGReader(t, 40, 40)); err != nil {
		t.Fatal(err)
	}
	if err := op.Process(t.Context()); err == nil {
		t.Fatal("expected error for invalid rotate, got nil")
	}
}

func Test_ImageOperator_BrightnessContrastGamma(t *testing.T) {
	t.Parallel()

	op := actor.NewImageOperator("image/png", actor.ImageOperatorOption{
		Brightness: 50,
		Contrast:   30,
		Gamma:      1.5,
	})
	if err := op.Decode(t.Context(), newTestPNGReader(t, 30, 30)); err != nil {
		t.Fatal(err)
	}
	if err := op.Process(t.Context()); err != nil {
		t.Fatalf("Process: %v", err)
	}
	if _, err := op.ImageByte(t.Context()); err != nil {
		t.Fatalf("ImageByte: %v", err)
	}
}

func Test_ImageOperator_Decode_Invalid(t *testing.T) {
	t.Parallel()

	op := actor.NewImageOperator("image/png", actor.ImageOperatorOption{})
	err := op.Decode(t.Context(), strings.NewReader("not-an-image"))
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func Test_ImageOperator_ImageByte_InvalidContentType(t *testing.T) {
	t.Parallel()

	op := actor.NewImageOperator("application/unknown", actor.ImageOperatorOption{})
	if err := op.Decode(t.Context(), newTestPNGReader(t, 20, 20)); err != nil {
		t.Fatal(err)
	}
	if err := op.Process(t.Context()); err != nil {
		t.Fatal(err)
	}
	if _, err := op.ImageByte(t.Context()); err == nil {
		t.Fatal("expected error for invalid content type, got nil")
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		n = -n
		neg = true
	}
	buf := make([]byte, 0, 4)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
