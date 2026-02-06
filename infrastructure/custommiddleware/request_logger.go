package custommiddleware

import (
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// JSONRequestLoggerConfig returns a request logger config that outputs the legacy JSON log format.
func JSONRequestLoggerConfig() middleware.RequestLoggerConfig {
	return middleware.RequestLoggerConfig{
		LogLatency:       true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogMethod:        true,
		LogURI:           true,
		LogRequestID:     true,
		LogUserAgent:     true,
		LogStatus:        true,
		LogContentLength: true,
		LogResponseSize:  true,
		HandleError:      true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			bytesIn := int64(0)
			if v.ContentLength != "" {
				if parsed, err := strconv.ParseInt(v.ContentLength, 10, 64); err == nil {
					bytesIn = parsed
				}
			}
			errMsg := ""
			if v.Error != nil {
				errMsg = v.Error.Error()
			}

			c.Logger().Info("request",
				"id", v.RequestID,
				"remote_ip", v.RemoteIP,
				"host", v.Host,
				"method", v.Method,
				"uri", v.URI,
				"user_agent", v.UserAgent,
				"status", v.Status,
				"error", errMsg,
				"latency", v.Latency.Nanoseconds(),
				"latency_human", v.Latency.String(),
				"bytes_in", bytesIn,
				"bytes_out", v.ResponseSize,
			)
			return nil
		},
	}
}

// JSONRequestLogger returns middleware that outputs the legacy JSON log format.
func JSONRequestLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(JSONRequestLoggerConfig())
}
