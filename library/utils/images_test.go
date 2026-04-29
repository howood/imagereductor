package utils_test

import (
	"bytes"
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
