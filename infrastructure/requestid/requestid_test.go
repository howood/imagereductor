package requestid_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/howood/imagereductor/infrastructure/requestid"
	"github.com/labstack/echo/v4"
)

func Test_RequestIDHandler(t *testing.T) {
	t.Parallel()

	checkvalue := "azxswedcvfr"
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Request-ID", checkvalue)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	xrequestid := requestid.GetRequestID(c.Request())
	if xrequestid != checkvalue {
		t.Fatal("failed test RequestIDHandler")
	}
	t.Log("success RequestIDHandler")
}

func Test_RequestIDHandlerCreateNew(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	xrequestid := requestid.GetRequestID(c.Request())
	if xrequestid == "" {
		t.Fatal("failed test RequestIDHandler Create New")
	}
	t.Log("success RequestIDHandler Create New")
}
