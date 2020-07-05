package custommiddleware

import (
	"github.com/labstack/echo/v4"
)

// OptionsMethodSkipper skip when option method requested
func OptionsMethodSkipper(c echo.Context) bool {
	if c.Request().Method == "OPTIONS" {
		return true
	}
	return false
}
