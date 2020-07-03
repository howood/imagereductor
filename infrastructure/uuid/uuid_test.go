package uuid

import (
	"testing"
)

func Test_GetUUID(t *testing.T) {
	result := GetUUID(SATORI_UUID)
	t.Log(result)
	t.Log("success GetUUID")
}
