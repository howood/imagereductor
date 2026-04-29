package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/howood/imagereductor/application/actor"
	"github.com/labstack/echo/v5"
)

func newEchoCtx(method, target string, body string) *echo.Context {
	e := echo.New()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequestWithContext(context.Background(), method, target, nil)
	} else {
		req = httptest.NewRequestWithContext(context.Background(), method, target, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, "application/x-www-form-urlencoded")
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func Test_setOptionValueInt(t *testing.T) {
	t.Parallel()

	h := &ImageReductionHandler{}
	ctx := context.Background()

	// empty value
	v, err := h.setOptionValueInt(ctx, "", nil)
	if err != nil || v != 0 {
		t.Fatalf("empty value should return 0,nil; got %d,%v", v, err)
	}
	// valid int
	v, err = h.setOptionValueInt(ctx, "123", nil)
	if err != nil || v != 123 {
		t.Fatalf("expected 123,nil; got %d,%v", v, err)
	}
	// invalid int
	_, err = h.setOptionValueInt(ctx, "notanumber", nil)
	if err == nil {
		t.Fatal("expected error for invalid int")
	}
	// pre-existing error short-circuits
	v, err = h.setOptionValueInt(ctx, "5", errFromString("prev"))
	if err == nil || v != 0 {
		t.Fatalf("expected pre-existing error to propagate; got %d,%v", v, err)
	}
}

func Test_setOptionValueFloat(t *testing.T) {
	t.Parallel()

	h := &ImageReductionHandler{}
	ctx := context.Background()

	v, err := h.setOptionValueFloat(ctx, "", nil)
	if err != nil || v != 0 {
		t.Fatalf("empty value should return 0,nil; got %v,%v", v, err)
	}
	v, err = h.setOptionValueFloat(ctx, "1.5", nil)
	if err != nil || v != 1.5 {
		t.Fatalf("expected 1.5,nil; got %v,%v", v, err)
	}
	_, err = h.setOptionValueFloat(ctx, "notafloat", nil)
	if err == nil {
		t.Fatal("expected error for invalid float")
	}
	v, err = h.setOptionValueFloat(ctx, "1.0", errFromString("prev"))
	if err == nil || v != 0 {
		t.Fatalf("pre-existing error should propagate")
	}
}

func Test_getCropParam(t *testing.T) {
	t.Parallel()

	h := &ImageReductionHandler{}
	ctx := context.Background()

	// empty -> zero
	got, err := h.getCropParam(ctx, "", nil)
	if err != nil || got != ([4]int{}) {
		t.Fatalf("empty should return zero; got %v,%v", got, err)
	}
	// valid
	got, err = h.getCropParam(ctx, "10,20,30,40", nil)
	if err != nil || got != ([4]int{10, 20, 30, 40}) {
		t.Fatalf("expected [10,20,30,40],nil; got %v,%v", got, err)
	}
	// wrong count
	_, err = h.getCropParam(ctx, "10,20,30", nil)
	if err == nil {
		t.Fatal("expected error for 3 elements")
	}
	// invalid int element
	_, err = h.getCropParam(ctx, "10,20,bad,40", nil)
	if err == nil {
		t.Fatal("expected error for non-int element")
	}
	// pre-existing error
	_, err = h.getCropParam(ctx, "10,20,30,40", errFromString("prev"))
	if err == nil {
		t.Fatal("pre-existing error should propagate")
	}
}

func Test_getImageOptionByFormValue(t *testing.T) {
	t.Parallel()

	h := &ImageReductionHandler{}
	c := newEchoCtx(http.MethodPost, "/", "w=100&h=200&q=3&rotate=right&bri=10&cont=20&gam=2.2&crop=1,2,3,4")
	opt, err := h.getImageOptionByFormValue(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := actor.ImageOperatorOption{
		Width:      100,
		Height:     200,
		Quality:    3,
		Rotate:     "right",
		Brightness: 10,
		Contrast:   20,
		Gamma:      2.2,
		Crop:       [4]int{1, 2, 3, 4},
	}
	if !reflect.DeepEqual(opt, want) {
		t.Fatalf("opt = %#v, want %#v", opt, want)
	}
}

func Test_getImageOptionByFormValue_InvalidWidth(t *testing.T) {
	t.Parallel()

	h := &ImageReductionHandler{}
	c := newEchoCtx(http.MethodPost, "/", "w=invalid")
	_, err := h.getImageOptionByFormValue(context.Background(), c)
	if err == nil {
		t.Fatal("expected error for invalid width")
	}
}

func errFromString(s string) error {
	return &simpleError{msg: s}
}

type simpleError struct{ msg string }

func (e *simpleError) Error() string { return e.msg }
