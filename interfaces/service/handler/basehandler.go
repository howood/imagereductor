package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/howood/imagereductor/application/actor/storageservice"
	"github.com/howood/imagereductor/library/utils"
	"github.com/labstack/echo/v4"
)

const (
	// FormKeyStorageKey is form key of storage key
	FormKeyStorageKey = "key"
	// FormKeyWidth is form key of width
	FormKeyWidth = "w"
	// FormKeyHeight is form key of height
	FormKeyHeight = "h"
	// FormKeyQuality is form key of quality
	FormKeyQuality = "q"
	// FormKeyNonUseCache is form key of nonusecache
	FormKeyNonUseCache = "nonusecache"
	// FormKeyRotate is form key of round
	FormKeyRotate = "rotate"
	// FormKeyCrop is form key of crop
	FormKeyCrop = "crop"
	// FormKeyUploadFile is form key of uploadfile
	FormKeyUploadFile = "uploadfile"
	// FormKeyPath is form key of path
	FormKeyPath = "path"
)

// BaseHandler struct
type BaseHandler struct {
	ctx context.Context
}

func (bh BaseHandler) errorResponse(c echo.Context, statudcode int, err error) error {
	if strings.Contains(strings.ToLower(err.Error()), storageservice.RecordNotFoundMsg) {
		statudcode = http.StatusNotFound
	}
	c.Response().Header().Set(echo.HeaderXRequestID, bh.ctx.Value(echo.HeaderXRequestID).(string))
	return c.JSONPretty(statudcode, map[string]interface{}{"message": err.Error()}, "    ")
}

func (bh BaseHandler) setResponseHeader(c echo.Context, lastmodified, contentlength string, expires, xrequestid string) {
	c.Response().Header().Set(echo.HeaderLastModified, lastmodified)
	c.Response().Header().Set(echo.HeaderContentLength, contentlength)
	c.Response().Header().Set(echo.HeaderXRequestID, xrequestid)
	c.Response().Header().Set("Cache-Control", fmt.Sprintf("max-age:%g, public", (bh.getHeaderExpires()*time.Second).Seconds()))
	if expires != "" {
		c.Response().Header().Set("Expires", expires)
	}
}

func (bh BaseHandler) setNewLatsModified() string {
	return time.Now().UTC().Format(http.TimeFormat)
}

func (bh BaseHandler) setExpires(epires time.Time) string {
	return epires.Add(bh.getHeaderExpires() * time.Second).UTC().Format(http.TimeFormat)
}

func (bh BaseHandler) getHeaderExpires() time.Duration {
	return time.Duration(utils.GetOsEnvInt("HEADEREXPIRED", 300))
}
