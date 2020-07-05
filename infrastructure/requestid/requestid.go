package requestid

import (
	"net/http"

	"github.com/howood/imagereductor/infrastructure/uuid"
)

const KeyRequestID = "X-Request-ID"

func generateRequestID() string {
	return uuid.GetUUID(uuid.SATORI_UUID)
}

func GetRequestID(r *http.Request) string {
	if r.Header.Get(KeyRequestID) != "" {
		return r.Header.Get(KeyRequestID)
	} else {
		return generateRequestID()
	}
}
