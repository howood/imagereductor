package cloudstorages_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/howood/imagereductor/infrastructure/client/cloudstorages"
)

func setupFakeGCS(t *testing.T) *cloudstorages.GCSInstance {
	t.Helper()

	const (
		bucket    = "test-bucket"
		projectID = "test-project"
	)

	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		InitialObjects:  []fakestorage.Object{},
		BucketsLocation: "US",
	})
	if err != nil {
		t.Fatalf("start fake-gcs-server: %v", err)
	}
	t.Cleanup(server.Stop)

	// Use the server's built-in client (already configured for auth).
	client := server.Client()
	ctx := t.Context()
	if err := client.Bucket(bucket).Create(ctx, projectID, nil); err != nil {
		t.Fatalf("create bucket: %v", err)
	}

	inst := cloudstorages.NewGCSInstanceForTest(client, cloudstorages.GCSConfig{
		ProjectID: projectID,
		Bucket:    bucket,
		Timeout:   30 * time.Second,
	})
	return inst
}

func TestGCSIntegration_PutAndGet(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupFakeGCS(t)
	ctx := t.Context()
	bucket := inst.GetBucket()

	content := []byte("hello gcs world")
	if err := inst.Put(ctx, bucket, "test/hello.txt", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	_, data, err := inst.Get(ctx, bucket, "test/hello.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Fatalf("Get data = %q, want %q", data, content)
	}
}

func TestGCSIntegration_GetByStreaming(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupFakeGCS(t)
	ctx := t.Context()
	bucket := inst.GetBucket()

	content := []byte("streaming gcs data")
	if err := inst.Put(ctx, bucket, "stream/data.bin", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	_, rc, err := inst.GetByStreaming(ctx, bucket, "stream/data.bin")
	if err != nil {
		t.Fatalf("GetByStreaming: %v", err)
	}
	defer rc.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(rc); err != nil {
		t.Fatalf("read stream: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), content) {
		t.Fatalf("stream data mismatch")
	}
}

func TestGCSIntegration_GetObjectInfo(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupFakeGCS(t)
	ctx := t.Context()
	bucket := inst.GetBucket()

	content := []byte("info test data gcs 12345")
	if err := inst.Put(ctx, bucket, "info/obj.txt", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	info, err := inst.GetObjectInfo(ctx, bucket, "info/obj.txt")
	if err != nil {
		t.Fatalf("GetObjectInfo: %v", err)
	}
	if info.ContentLength != len(content) {
		t.Fatalf("ContentLength = %d, want %d", info.ContentLength, len(content))
	}
}

func TestGCSIntegration_List(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupFakeGCS(t)
	ctx := t.Context()
	bucket := inst.GetBucket()

	keys := []string{"list/a.txt", "list/b.txt", "list/sub/c.txt"}
	for _, k := range keys {
		if err := inst.Put(ctx, bucket, k, bytes.NewReader([]byte("x"))); err != nil {
			t.Fatalf("Put %s: %v", k, err)
		}
	}

	names, err := inst.List(ctx, bucket, "list/")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != len(keys) {
		t.Fatalf("List len = %d, want %d; got %v", len(names), len(keys), names)
	}
}

func TestGCSIntegration_Delete(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupFakeGCS(t)
	ctx := t.Context()
	bucket := inst.GetBucket()

	if err := inst.Put(ctx, bucket, "del/target.txt", bytes.NewReader([]byte("delete me"))); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := inst.Delete(ctx, bucket, "del/target.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, _, err := inst.Get(ctx, bucket, "del/target.txt")
	if err == nil {
		t.Fatal("Get after Delete should return error")
	}
}

func TestGCSIntegration_MultipleObjects(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupFakeGCS(t)
	ctx := t.Context()
	bucket := inst.GetBucket()

	for i := range 10 {
		key := fmt.Sprintf("multi/obj_%d.txt", i)
		if err := inst.Put(ctx, bucket, key, bytes.NewReader(fmt.Appendf(nil, "data %d", i))); err != nil {
			t.Fatalf("Put %s: %v", key, err)
		}
	}

	names, err := inst.List(ctx, bucket, "multi/")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != 10 {
		t.Fatalf("expected 10 objects, got %d", len(names))
	}
}

// Verify no error for context.Background (no timeout).
func TestGCSIntegration_NoTimeout(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	const (
		bucket    = "no-timeout-bucket"
		projectID = "test-project"
	)

	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{})
	if err != nil {
		t.Fatalf("start fake-gcs: %v", err)
	}
	t.Cleanup(server.Stop)

	client := server.Client()
	ctx := context.Background()
	if err := client.Bucket(bucket).Create(ctx, projectID, nil); err != nil {
		t.Fatalf("create bucket: %v", err)
	}

	inst := cloudstorages.NewGCSInstanceForTest(client, cloudstorages.GCSConfig{
		ProjectID: projectID,
		Bucket:    bucket,
		Timeout:   0, // no timeout
	})

	content := []byte("no timeout test")
	if err := inst.Put(ctx, bucket, "nt.txt", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	_, data, err := inst.Get(ctx, bucket, "nt.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Fatalf("mismatch")
	}
}
