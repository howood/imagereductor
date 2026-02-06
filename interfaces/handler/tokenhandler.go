package handler

import (
	"context"
	"net/http"

	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/howood/imagereductor/infrastructure/requestid"
	"github.com/howood/imagereductor/infrastructure/uuid"
	"github.com/labstack/echo/v5"
)

// TokenHandler struct.
type TokenHandler struct {
	BaseHandler
}

func NewTokenHandler(baseHandler BaseHandler) *TokenHandler {
	return &TokenHandler{BaseHandler: baseHandler}
}

// Request is get from storage.
func (th *TokenHandler) Request(c *echo.Context) error {
	xRequestID := requestid.GetRequestID(c.Request())
	ctx := context.WithValue(c.Request().Context(), requestid.GetRequestIDKey(), xRequestID)
	log.Info(ctx, "========= START REQUEST : "+c.Request().URL.RequestURI())
	log.Info(ctx, c.Request().Method)
	log.Info(ctx, c.Request().Header)
	claimname := uuid.GetUUID(uuid.SatoriUUID)
	tokenstr := th.UcCluster.TokenUC.CreateToken(ctx, claimname)
	return c.JSONPretty(http.StatusOK, map[string]any{"token": tokenstr}, "    ")
}
