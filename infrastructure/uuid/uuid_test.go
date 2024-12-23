package uuid_test

import (
	"testing"

	"github.com/howood/imagereductor/infrastructure/uuid"
)

func Test_GetUUID(t *testing.T) {
	t.Parallel()

	result := uuid.GetUUID(uuid.SatoriUUID)
	t.Log(result)
	t.Log("success GetUUID")
}
