package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/howood/imagereductor/application/actor/storageservice"
	"github.com/howood/imagereductor/infrastructure/requestid"
	"github.com/howood/imagereductor/library/utils"
	"github.com/labstack/echo/v4"
)

const (
	marshalPrefix = ""
	marshalIndent = "    "
)

// BaseHandler struct.
type BaseHandler struct{}

//nolint:unparam
func (bh BaseHandler) errorResponse(ctx context.Context, c echo.Context, statudcode int, err error) error {
	if strings.Contains(strings.ToLower(err.Error()), storageservice.RecordNotFoundMsg) {
		statudcode = http.StatusNotFound
	}
	c.Response().Header().Set(echo.HeaderXRequestID, fmt.Sprintf("%v", ctx.Value(requestid.GetRequestIDKey())))
	return c.JSONPretty(statudcode, map[string]interface{}{"message": err.Error()}, marshalIndent)
}

func (bh BaseHandler) setResponseHeader(c echo.Context, lastmodified, contentlength string, expires, xrequestid string) {
	c.Response().Header().Set(echo.HeaderLastModified, lastmodified)
	c.Response().Header().Set(echo.HeaderContentLength, contentlength)
	c.Response().Header().Set(echo.HeaderXRequestID, xrequestid)
	c.Response().Header().Set("Cache-Control", fmt.Sprintf("max-age:%g, public", (time.Duration(bh.getHeaderExpires())*time.Second).Seconds()))
	if expires != "" {
		c.Response().Header().Set("Expires", expires)
	}
}

func (bh BaseHandler) setNewLatsModified() string {
	return time.Now().UTC().Format(http.TimeFormat)
}

func (bh BaseHandler) setExpires(epires time.Time) string {
	return epires.Add(time.Duration(bh.getHeaderExpires()) * time.Second).UTC().Format(http.TimeFormat)
}

//nolint:mnd
func (bh BaseHandler) getHeaderExpires() int {
	return utils.GetOsEnvInt("HEADEREXPIRED", 300)
}

func (bh BaseHandler) jsonToByte(jsondata interface{}) ([]byte, error) {
	return json.MarshalIndent(jsondata, marshalPrefix, marshalIndent)
}
