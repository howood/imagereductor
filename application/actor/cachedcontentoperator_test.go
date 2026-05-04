package actor_test

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"testing"

	"github.com/howood/imagereductor/application/actor"
)

func Test_CachedContentOperator_SetGet(t *testing.T) {
	t.Parallel()

	op := actor.NewCachedContentOperator()
	op.Set("image/png", "Mon, 01 Jan 2024 00:00:00 GMT", []byte("hello"))

	if got := op.GetContentType(); got != "image/png" {
		t.Fatalf("GetContentType = %q, want image/png", got)
	}
	if got := op.GetLastModified(); got != "Mon, 01 Jan 2024 00:00:00 GMT" {
		t.Fatalf("GetLastModified = %q, unexpected", got)
	}
	if got := op.GetContent(); !reflect.DeepEqual(got, []byte("hello")) {
		t.Fatalf("GetContent = %v, want hello", got)
	}
}

func Test_CachedContentOperator_GobEncodeDecode(t *testing.T) {
	t.Parallel()

	src := actor.NewCachedContentOperator()
	src.Set("image/jpeg", "lastmodified-value", []byte("payload-bytes"))
	encoded, err := src.GobEncode()
	if err != nil {
		t.Fatalf("GobEncode failed: %v", err)
	}
	if len(encoded) == 0 {
		t.Fatal("GobEncode returned empty bytes")
	}

	dst := actor.NewCachedContentOperator()
	if err := dst.GobDecode(encoded); err != nil {
		t.Fatalf("GobDecode failed: %v", err)
	}
	if dst.GetContentType() != "image/jpeg" ||
		dst.GetLastModified() != "lastmodified-value" ||
		!reflect.DeepEqual(dst.GetContent(), []byte("payload-bytes")) {
		t.Fatal("GobDecode returned mismatched data")
	}
}

func Test_CachedContentOperator_GobDecode_Invalid(t *testing.T) {
	t.Parallel()

	dst := actor.NewCachedContentOperator()
	if err := dst.GobDecode([]byte("not-a-valid-gob")); err == nil {
		t.Fatal("expected error decoding invalid gob bytes, got nil")
	}
}

func Test_CachedContentOperator_GobDecode_PartialContentType(t *testing.T) {
	t.Parallel()

	// Encode only ContentType — missing LastModified and Content
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode("image/png"); err != nil {
		t.Fatal(err)
	}

	dst := actor.NewCachedContentOperator()
	if err := dst.GobDecode(buf.Bytes()); err == nil {
		t.Fatal("expected error decoding partial gob (missing LastModified), got nil")
	}
}

func Test_CachedContentOperator_GobDecode_PartialLastModified(t *testing.T) {
	t.Parallel()

	// Encode ContentType + LastModified — missing Content
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode("image/png"); err != nil {
		t.Fatal(err)
	}
	if err := encoder.Encode("Mon, 01 Jan 2024 00:00:00 GMT"); err != nil {
		t.Fatal(err)
	}

	dst := actor.NewCachedContentOperator()
	if err := dst.GobDecode(buf.Bytes()); err == nil {
		t.Fatal("expected error decoding partial gob (missing Content), got nil")
	}
}
