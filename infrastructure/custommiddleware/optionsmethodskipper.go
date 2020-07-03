package custommiddleware

import (
	"github.com/labstack/echo/v4"
)

func OptionsMethodSkipper(c echo.Context) bool {
	if c.Request().Method == "OPTIONS" {
		return true
	}
	return false
}
