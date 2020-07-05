package uuid

import (
	"testing"
)

func Test_GetUUID(t *testing.T) {
	result := GetUUID(SatoriUUID)
	t.Log(result)
	t.Log("success GetUUID")
}
