package utils_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/howood/imagereductor/library/utils"
)

func Test_GetContentTypeByReadSeeker_Text(t *testing.T) {
	t.Parallel()

	reader := strings.NewReader("hello world this is plain text")
	ct, err := utils.GetContentTypeByReadSeeker(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(ct, "text/plain") {
		t.Fatalf("expected text/plain content type, got %q", ct)
	}
}

func Test_GetContentTypeByReadSeeker_PNG(t *testing.T) {
	t.Parallel()

	// PNG signature
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	reader := bytes.NewReader(pngSignature)
	ct, err := utils.GetContentTypeByReadSeeker(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ct != "image/png" {
		t.Fatalf("expected image/png, got %q", ct)
	}
}

func Test_GetContentTypeByReadSeeker_Empty(t *testing.T) {
	t.Parallel()

	// Empty reader triggers io.EOF on first Read - should still work.
	reader := bytes.NewReader([]byte{})
	ct, err := utils.GetContentTypeByReadSeeker(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty data returns text/plain from DetectContentType.
	if ct == "" {
		t.Fatal("expected non-empty content type")
	}
}

func Test_GetContentTypeByReadSeeker_SeekError(t *testing.T) {
	t.Parallel()

	reader := &failSeekReader{}
	_, err := utils.GetContentTypeByReadSeeker(reader)
	if err == nil {
		t.Fatal("expected error on seek failure")
	}
}

// failSeekReader always fails on Seek.
type failSeekReader struct{}

func (f *failSeekReader) Read(p []byte) (int, error) {
	return 0, nil
}

func (f *failSeekReader) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("seek failed")
}
