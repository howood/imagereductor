package requestid

import (
	"net/http"

	"github.com/howood/imagereductor/infrastructure/uuid"
)

type RequestContextKey string

// KeyRequestID is XRequestId key.
const KeyRequestID = "X-Request-Id"

func generateRequestID() string {
	return uuid.GetUUID(uuid.SatoriUUID)
}

// GetRequestID returns XRequestId.
func GetRequestID(r *http.Request) string {
	if r.Header.Get(KeyRequestID) != "" {
		return r.Header.Get(KeyRequestID)
	}
	return generateRequestID()
}

// GetRequestIDKey returns RequestContextKey.
func GetRequestIDKey() RequestContextKey {
	return RequestContextKey(KeyRequestID)
}
