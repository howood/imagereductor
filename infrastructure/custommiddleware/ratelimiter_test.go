package custommiddleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/howood/imagereductor/infrastructure/custommiddleware"
	"github.com/labstack/echo/v5"
	"golang.org/x/time/rate"
)

func Test_RateLimiter_AllowsAndBlocks(t *testing.T) {
	t.Parallel()

	rl := custommiddleware.NewRateLimiter(custommiddleware.RateLimitConfig{
		Rate:  rate.Limit(1), // 1 req/sec
		Burst: 1,
		KeyFunc: func(_ *echo.Context) string {
			return "shared-key"
		},
	})

	e := echo.New()
	mw := rl.Middleware()

	exec := func() int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err := mw(func(_ *echo.Context) error {
			return nil
		})(c)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			return httpErr.Code
		}
		return http.StatusOK
	}

	if code := exec(); code != http.StatusOK {
		t.Fatalf("first request should be allowed, got %d", code)
	}
	if code := exec(); code != http.StatusTooManyRequests {
		t.Fatalf("second request should be limited, got %d", code)
	}
}

func Test_RateLimiter_Skipper(t *testing.T) {
	t.Parallel()

	rl := custommiddleware.NewRateLimiter(custommiddleware.RateLimitConfig{
		Rate:  rate.Limit(1),
		Burst: 1,
		Skipper: func(_ *echo.Context) bool {
			return true
		},
	})
	e := echo.New()
	mw := rl.Middleware()
	for range 5 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if err := mw(func(_ *echo.Context) error { return nil })(c); err != nil {
			t.Fatalf("skipper should bypass limiter, got err %v", err)
		}
	}
}

func Test_RateLimiter_Defaults(t *testing.T) {
	t.Parallel()

	rl := custommiddleware.NewRateLimiter(custommiddleware.RateLimitConfig{
		Rate:  rate.Limit(100),
		Burst: 100,
	})
	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}
	// Use default key func + default error msg
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	mw := rl.Middleware()
	if err := mw(func(_ *echo.Context) error { return nil })(c); err != nil {
		t.Fatalf("expected first request to pass: %v", err)
	}
}

func Test_RateLimiter_MaxKeysEviction(t *testing.T) {
	t.Parallel()

	keys := []string{"a", "b", "c"}
	idx := 0
	rl := custommiddleware.NewRateLimiter(custommiddleware.RateLimitConfig{
		Rate:    rate.Limit(10),
		Burst:   10,
		MaxKeys: 2,
		KeyFunc: func(_ *echo.Context) string {
			k := keys[idx]
			idx++
			return k
		},
	})
	e := echo.New()
	mw := rl.Middleware()
	for range 3 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if err := mw(func(_ *echo.Context) error { return nil })(c); err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	}
	// just exercises eviction path - no panic = success
}

func Test_IPKeyFunc(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if got := custommiddleware.IPKeyFunc(c); got == "" {
		t.Fatal("IPKeyFunc returned empty string")
	}
}

func Test_APIKeyFunc(t *testing.T) {
	t.Parallel()

	e := echo.New()
	// with auth header
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if got := custommiddleware.APIKeyFunc(c); got != "api:Bearer abc" {
		t.Fatalf("APIKeyFunc with auth = %q, want api:Bearer abc", got)
	}
	// without auth header -> falls back to IP
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.0.5:1234"
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if got := custommiddleware.APIKeyFunc(c2); got == "" {
		t.Fatal("APIKeyFunc fallback returned empty string")
	}
}

// ensure cleanup loop time.Tick path isn't a panic risk (limited duration)
func Test_RateLimiter_Cleanup_NoPanic(t *testing.T) {
	t.Parallel()

	rl := custommiddleware.NewRateLimiter(custommiddleware.RateLimitConfig{
		Rate:         rate.Limit(10),
		Burst:        10,
		CleanupTTL:   1 * time.Millisecond,
		CleanupEvery: 5 * time.Millisecond,
	})
	if rl == nil {
		t.Fatal("nil")
	}
	time.Sleep(20 * time.Millisecond)
}
