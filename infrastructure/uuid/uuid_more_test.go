package uuid_test

import (
	"testing"

	"github.com/howood/imagereductor/infrastructure/uuid"
)

func Test_GetUUID_AllVariants(t *testing.T) {
	t.Parallel()

	cases := []string{
		uuid.SatoriUUID,
		uuid.SegmentioKsuid,
		uuid.RsXid,
		"unknown_default",
	}
	for _, kind := range cases {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			got := uuid.GetUUID(kind)
			if got == "" {
				t.Fatalf("GetUUID(%q) returned empty string", kind)
			}
		})
	}
}

func Test_GetUUID_Uniqueness(t *testing.T) {
	t.Parallel()

	a := uuid.GetUUID(uuid.RsXid)
	b := uuid.GetUUID(uuid.RsXid)
	if a == b {
		t.Fatalf("expected different uuids, got identical: %s", a)
	}
}
