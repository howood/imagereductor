package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/howood/imagereductor/application/actor/storageservice"
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

func (bh BaseHandler) setResponseHeader(c echo.Context, lastmodified, contentlength, xrequestud string) {
	c.Response().Header().Set(echo.HeaderLastModified, lastmodified)
	c.Response().Header().Set(echo.HeaderContentLength, contentlength)
	c.Response().Header().Set(echo.HeaderXRequestID, xrequestud)
}

func (bh BaseHandler) setNewLatsModified() string {
	return time.Now().UTC().Format(http.TimeFormat)
}
