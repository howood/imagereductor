package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

var (
	errSomeFailure  = errors.New("some failure")
	errStatusNotFnd = errors.New("status code: 404 not found")
)

func Test_BaseHandler_errorResponse(t *testing.T) {
	t.Parallel()

	bh := BaseHandler{}
	e := echo.New()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := bh.errorResponse(context.Background(), c, http.StatusBadRequest, errSomeFailure)
	if err != nil {
		t.Fatalf("errorResponse returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func Test_BaseHandler_errorResponse_NotFound(t *testing.T) {
	t.Parallel()

	bh := BaseHandler{}
	e := echo.New()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := bh.errorResponse(context.Background(), c, http.StatusBadRequest, errStatusNotFnd)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 (auto-detect), got %d", rec.Code)
	}
}

func Test_BaseHandler_setNewLatsModified(t *testing.T) {
	t.Parallel()

	bh := BaseHandler{}
	got := bh.setNewLatsModified()
	if got == "" {
		t.Fatal("expected non-empty time string")
	}
}

func Test_BaseHandler_jsonToByte(t *testing.T) {
	t.Parallel()

	bh := BaseHandler{}
	b, err := bh.jsonToByte(map[string]string{"a": "b"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty json bytes")
	}
}

func Test_BaseHandler_setResponseHeader(t *testing.T) {
	t.Parallel()

	bh := BaseHandler{}
	e := echo.New()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	bh.setResponseHeader(c, "Mon, 01 Jan 2024 00:00:00 GMT", "100", "Mon, 01 Jan 2024 00:05:00 GMT", "req-id-1")
	if got := rec.Header().Get(echo.HeaderXRequestID); got != "req-id-1" {
		t.Fatalf("X-Request-ID = %q, want req-id-1", got)
	}
	if got := rec.Header().Get(echo.HeaderContentLength); got != "100" {
		t.Fatalf("Content-Length = %q, want 100", got)
	}
	if got := rec.Header().Get("Expires"); got == "" {
		t.Fatal("expected Expires header to be set")
	}
}

func Test_BaseHandler_setResponseHeader_NoExpires(t *testing.T) {
	t.Parallel()

	bh := BaseHandler{}
	e := echo.New()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	bh.setResponseHeader(c, "lm", "0", "", "rid")
	if got := rec.Header().Get("Expires"); got != "" {
		t.Fatalf("Expires should not be set when empty, got %q", got)
	}
}

func Test_BaseHandler_setExpires(t *testing.T) { //nolint:paralleltest
	t.Setenv("HEADEREXPIRED", "600")
	bh := BaseHandler{}
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := bh.setExpires(base)
	if got == "" {
		t.Fatal("expected non-empty expires string")
	}
	// Should be base + 600s = 00:10:00.
	want := base.Add(600 * time.Second).UTC().Format(http.TimeFormat)
	if got != want {
		t.Fatalf("setExpires = %q, want %q", got, want)
	}
}

func Test_BaseHandler_getHeaderExpires_Default(t *testing.T) { //nolint:paralleltest
	t.Setenv("HEADEREXPIRED", "")
	bh := BaseHandler{}
	got := bh.getHeaderExpires()
	if got != 300 {
		t.Fatalf("getHeaderExpires = %d, want 300", got)
	}
}

func Test_BaseHandler_getHeaderExpires_Custom(t *testing.T) { //nolint:paralleltest
	t.Setenv("HEADEREXPIRED", "120")
	bh := BaseHandler{}
	got := bh.getHeaderExpires()
	if got != 120 {
		t.Fatalf("getHeaderExpires = %d, want 120", got)
	}
}
