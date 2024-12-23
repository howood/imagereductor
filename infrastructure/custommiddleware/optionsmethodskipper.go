package custommiddleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// OptionsMethodSkipper skip when option method requested.
func OptionsMethodSkipper(c echo.Context) bool {
	return c.Request().Method == http.MethodOptions
}
