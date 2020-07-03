package uuid

import (
	"github.com/gofrs/uuid"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
)

const (
	SEGMENTIO_KSUID = "segmentio_ksuid"
	SATORI_UUID     = "satori_gouuid"
	RS_XID          = "rs_xid"
)

func GetUUID(systemuuid string) string {
	switch systemuuid {
	case SEGMENTIO_KSUID:
		return ksuid.New().String()
	case SATORI_UUID:
		return uuid.Must(uuid.NewV4()).String()
	case RS_XID:
		return xid.New().String()
	default:
		return xid.New().String()
	}
}
