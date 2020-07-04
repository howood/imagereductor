package handler

import (
	"net/http"

	"github.com/howood/imagereductor/application/actor"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/howood/imagereductor/infrastructure/uuid"
	"github.com/labstack/echo/v4"
)

// TokenHandler struct
type TokenHandler struct {
}

// Request is get from storage
func (th TokenHandler) Request(c echo.Context) error {
	log.Info("========= START REQUEST : " + c.Request().URL.RequestURI())
	log.Info(c.Request().Method)
	log.Info(c.Request().Header)
	claimname := uuid.GetUUID(uuid.SATORI_UUID)
	jwtinstance := actor.NewJwtOperator(claimname, false, actor.TokenExpired)
	tokenstr := jwtinstance.CreateToken(actor.TokenSecret)
	return c.JSONPretty(http.StatusOK, map[string]interface{}{"token": tokenstr}, "    ")
}
