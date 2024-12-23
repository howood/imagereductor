package custommiddleware

import (
	"net"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

// IPRestriction permit access by IP address.
func IPRestriction() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if os.Getenv("TOKENAPI_ALLOW_IPS") == "" {
				return next(c)
			}
			allowIPs := strings.Split(os.Getenv("TOKENAPI_ALLOW_IPS"), ",")
			requestIP := net.ParseIP(c.RealIP())
			for _, ip := range allowIPs {
				_, ipnet, err := net.ParseCIDR(ip)
				if err != nil {
					return err
				}
				if ipnet.Contains(requestIP) {
					return next(c)
				}
			}
			return echo.ErrUnauthorized
		}
	}
}
