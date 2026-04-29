package custommiddleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/howood/imagereductor/infrastructure/custommiddleware"
	"github.com/labstack/echo/v5"
)

func Test_OptionsMethodSkipper(t *testing.T) {
	t.Parallel()

	e := echo.New()
	cases := []struct {
		method string
		want   bool
	}{
		{http.MethodOptions, true},
		{http.MethodGet, false},
		{http.MethodPost, false},
	}
	for _, tc := range cases {
		req := httptest.NewRequestWithContext(t.Context(), tc.method, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if got := custommiddleware.OptionsMethodSkipper(c); got != tc.want {
			t.Fatalf("OptionsMethodSkipper(%s) = %v, want %v", tc.method, got, tc.want)
		}
	}
}

func Test_IPRestriction_AllowsLocalhost(t *testing.T) {
	t.Parallel()

	// 127.0.0.1 is included in the default container env (TOKENAPI_ALLOW_IPS) and is
	// also the pass-through target when IP restriction is disabled, so this works in
	// both configurations.
	e := echo.New()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	mw := custommiddleware.IPRestriction()
	handler := mw(func(_ *echo.Context) error {
		called = true
		return nil
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("next handler was not called")
	}
}

func Test_JSONRequestLoggerConfig(t *testing.T) {
	t.Parallel()

	cfg := custommiddleware.JSONRequestLoggerConfig()
	if cfg.LogValuesFunc == nil {
		t.Fatal("LogValuesFunc should not be nil")
	}
	if !cfg.LogLatency || !cfg.LogStatus || !cfg.LogMethod || !cfg.LogURI {
		t.Fatal("expected logger to log latency/status/method/uri")
	}
}

func Test_JSONRequestLogger_Middleware(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.Use(custommiddleware.JSONRequestLogger())
	e.GET("/ping", func(c *echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/ping", nil)
	req.Header.Set(echo.HeaderContentLength, "0")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "pong" {
		t.Fatalf("body = %q, want pong", rec.Body.String())
	}
}
