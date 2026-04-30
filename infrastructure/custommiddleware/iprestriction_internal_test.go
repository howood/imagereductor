package custommiddleware

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func parseNetwork(cidr string) (net.IP, *net.IPNet, error) {
	return net.ParseCIDR(cidr)
}

func Test_IPRestriction_DeniedIP(t *testing.T) { //nolint:paralleltest
	origEnabled := ipRestrictionEnabled
	origNetworks := allowedNetworks
	t.Cleanup(func() {
		ipRestrictionEnabled = origEnabled
		allowedNetworks = origNetworks
	})

	ipRestrictionEnabled = true
	_, ipnet, _ := parseNetwork("10.0.0.0/8")
	allowedNetworks = []*net.IPNet{ipnet}

	e := echo.New()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.100:5000"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := IPRestriction()
	handler := mw(func(_ *echo.Context) error {
		t.Fatal("handler should not be called for denied IP")
		return nil
	})

	err := handler(c)
	if !errors.Is(err, echo.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func Test_IPRestriction_InvalidIP(t *testing.T) { //nolint:paralleltest
	origEnabled := ipRestrictionEnabled
	origNetworks := allowedNetworks
	t.Cleanup(func() {
		ipRestrictionEnabled = origEnabled
		allowedNetworks = origNetworks
	})

	ipRestrictionEnabled = true
	_, ipnet, _ := parseNetwork("10.0.0.0/8")
	allowedNetworks = []*net.IPNet{ipnet}

	e := echo.New()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.RemoteAddr = "not-a-valid-ip:1234"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := IPRestriction()
	handler := mw(func(_ *echo.Context) error {
		t.Fatal("handler should not be called for invalid IP")
		return nil
	})

	err := handler(c)
	if !errors.Is(err, echo.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func Test_IPRestriction_Disabled(t *testing.T) { //nolint:paralleltest
	origEnabled := ipRestrictionEnabled
	t.Cleanup(func() { ipRestrictionEnabled = origEnabled })

	ipRestrictionEnabled = false

	e := echo.New()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:9999"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	mw := IPRestriction()
	handler := mw(func(_ *echo.Context) error {
		called = true
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called when disabled")
	}
}

func Test_IPRestriction_AllowedIP(t *testing.T) { //nolint:paralleltest
	origEnabled := ipRestrictionEnabled
	origNetworks := allowedNetworks
	t.Cleanup(func() {
		ipRestrictionEnabled = origEnabled
		allowedNetworks = origNetworks
	})

	ipRestrictionEnabled = true
	_, ipnet, _ := parseNetwork("192.168.0.0/16")
	allowedNetworks = []*net.IPNet{ipnet}

	e := echo.New()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.50:8080"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	mw := IPRestriction()
	handler := mw(func(_ *echo.Context) error {
		called = true
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called for allowed IP")
	}
}
