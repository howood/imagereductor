package uuid

import (
	"github.com/gofrs/uuid"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
)

const (
	// SegmentioKsuid is type of id
	SegmentioKsuid = "segmentio_ksuid"
	// SatoriUUID is type of id
	SatoriUUID = "satori_gouuid"
	// RsXid is type of id
	RsXid = "rs_xid"
)

// GetUUID returns a new uuid
func GetUUID(systemuuid string) string {
	switch systemuuid {
	case SegmentioKsuid:
		return ksuid.New().String()
	case SatoriUUID:
		return uuid.Must(uuid.NewV4()).String()
	case RsXid:
		return xid.New().String()
	default:
		return xid.New().String()
	}
}
