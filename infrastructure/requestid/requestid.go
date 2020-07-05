package requestid

import (
	"net/http"

	"github.com/howood/imagereductor/infrastructure/uuid"
	"github.com/labstack/echo/v4"
)

func generateRequestID() string {
	return uuid.GetUUID(uuid.SATORI_UUID)
}

func GetRequestID(r *http.Request) string {
	if r.Header.Get(echo.HeaderXRequestID) != "" {
		return r.Header.Get(echo.HeaderXRequestID)
	} else {
		return generateRequestID()
	}
}
