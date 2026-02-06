package custommiddleware

import (
	"net"
	"os"
	"strings"

	"github.com/labstack/echo/v5"
)

// IPRestriction permit access by IP address.
var (
	//nolint:gochecknoglobals
	allowedNetworks []*net.IPNet
	//nolint:gochecknoglobals
	ipRestrictionEnabled bool
)

//nolint:gochecknoinits
func init() {
	allowIPsEnv := os.Getenv("TOKENAPI_ALLOW_IPS")
	if allowIPsEnv == "" {
		ipRestrictionEnabled = false
		return
	}

	ipRestrictionEnabled = true
	for ip := range strings.SplitSeq(allowIPsEnv, ",") {
		_, ipnet, err := net.ParseCIDR(strings.TrimSpace(ip))
		if err != nil {
			// Log error and continue with other IPs
			continue
		}
		allowedNetworks = append(allowedNetworks, ipnet)
	}
}

func IPRestriction() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if !ipRestrictionEnabled {
				return next(c)
			}

			requestIP := net.ParseIP(c.RealIP())
			if requestIP == nil {
				return echo.ErrUnauthorized
			}

			for _, ipnet := range allowedNetworks {
				if ipnet.Contains(requestIP) {
					return next(c)
				}
			}
			return echo.ErrUnauthorized
		}
	}
}
