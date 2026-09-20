package requestid_test

import (
	"testing"

	"github.com/howood/imagereductor/infrastructure/requestid"
)

func Test_GetRequestIDKey(t *testing.T) {
	t.Parallel()

	got := requestid.GetRequestIDKey()
	if string(got) != "X-Request-ID" {
		t.Fatalf("GetRequestIDKey = %q, want X-Request-ID", string(got))
	}
}
