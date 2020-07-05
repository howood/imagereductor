package handler

import (
	"context"
	"net/http"

	"github.com/howood/imagereductor/application/actor"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/howood/imagereductor/infrastructure/requestid"
	"github.com/howood/imagereductor/infrastructure/uuid"
	"github.com/labstack/echo/v4"
)

// TokenHandler struct
type TokenHandler struct {
	BaseHandler
}

// Request is get from storage
func (th TokenHandler) Request(c echo.Context) error {
	xRequestID := requestid.GetRequestID(c.Request())
	th.ctx = context.WithValue(context.Background(), echo.HeaderXRequestID, xRequestID)
	log.Info(th.ctx, "========= START REQUEST : "+c.Request().URL.RequestURI())
	log.Info(th.ctx, c.Request().Method)
	log.Info(th.ctx, c.Request().Header)
	claimname := uuid.GetUUID(uuid.SATORI_UUID)
	jwtinstance := actor.NewJwtOperator(th.ctx, claimname, false, actor.TokenExpired)
	tokenstr := jwtinstance.CreateToken(actor.TokenSecret)
	return c.JSONPretty(http.StatusOK, map[string]interface{}{"token": tokenstr}, "    ")
}
