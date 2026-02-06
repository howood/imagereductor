package custommiddleware

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
		LogError:         true,
		LogContentLength: true,
		LogResponseSize:  true,
		HandleError:      true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
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

			entry := struct {
				Time         string `json:"time"`
				ID           string `json:"id"`
				RemoteIP     string `json:"remote_ip"`
				Host         string `json:"host"`
				Method       string `json:"method"`
				URI          string `json:"uri"`
				UserAgent    string `json:"user_agent"`
				Status       int    `json:"status"`
				Error        string `json:"error"`
				Latency      int64  `json:"latency"`
				LatencyHuman string `json:"latency_human"`
				BytesIn      int64  `json:"bytes_in"`
				BytesOut     int64  `json:"bytes_out"`
			}{
				Time:         time.Now().Format(time.RFC3339Nano),
				ID:           v.RequestID,
				RemoteIP:     v.RemoteIP,
				Host:         v.Host,
				Method:       v.Method,
				URI:          v.URI,
				UserAgent:    v.UserAgent,
				Status:       v.Status,
				Error:        errMsg,
				Latency:      v.Latency.Nanoseconds(),
				LatencyHuman: v.Latency.String(),
				BytesIn:      bytesIn,
				BytesOut:     v.ResponseSize,
			}

			payload, err := json.Marshal(entry)
			if err != nil {
				return err
			}
			_, err = c.Logger().Output().Write(append(payload, '\n'))
			return err
		},
	}
}

// JSONRequestLogger returns middleware that outputs the legacy JSON log format.
func JSONRequestLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(JSONRequestLoggerConfig())
}
