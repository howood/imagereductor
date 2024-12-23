package handler

import (
	"context"
	"net/http"

	"github.com/howood/imagereductor/application/usecase"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/howood/imagereductor/infrastructure/requestid"
	"github.com/howood/imagereductor/infrastructure/uuid"
	"github.com/labstack/echo/v4"
)

// TokenHandler struct.
type TokenHandler struct {
	BaseHandler
}

// Request is get from storage.
func (th TokenHandler) Request(c echo.Context) error {
	xRequestID := requestid.GetRequestID(c.Request())
	ctx := context.WithValue(context.Background(), requestid.GetRequestIDKey(), xRequestID)
	log.Info(ctx, "========= START REQUEST : "+c.Request().URL.RequestURI())
	log.Info(ctx, c.Request().Method)
	log.Info(ctx, c.Request().Header)
	claimname := uuid.GetUUID(uuid.SatoriUUID)
	tokenstr := usecase.TokenUsecase{}.CreateToken(ctx, claimname)
	return c.JSONPretty(http.StatusOK, map[string]interface{}{"token": tokenstr}, "    ")
}
