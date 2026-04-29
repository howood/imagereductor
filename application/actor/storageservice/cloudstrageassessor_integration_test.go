package storageservice_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/howood/imagereductor/application/actor/storageservice"
	"github.com/howood/imagereductor/infrastructure/client/cloudstorages"
)

func setupStorageAssessor(t *testing.T) *storageservice.CloudStorageAssessor {
	t.Helper()

	const (
		bucket    = "storageservice-test"
		projectID = "test-project"
	)

	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{})
	if err != nil {
		t.Fatalf("start fake-gcs: %v", err)
	}
	t.Cleanup(server.Stop)

	client := server.Client()
	ctx := t.Context()
	if err := client.Bucket(bucket).Create(ctx, projectID, nil); err != nil {
		t.Fatalf("create bucket: %v", err)
	}

	gcsInst := cloudstorages.NewGCSInstanceForTest(client, cloudstorages.GCSConfig{
		ProjectID: projectID,
		Bucket:    bucket,
		Timeout:   30 * time.Second,
	})

	return storageservice.NewCloudStorageAssessorForTest(gcsInst)
}

func TestCloudStorageAssessor_PutAndGet(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	csa := setupStorageAssessor(t)
	ctx := t.Context()

	content := []byte("assessor put/get test")
	if err := csa.Put(ctx, "test/file.txt", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	_, data, err := csa.Get(ctx, "test/file.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Fatalf("data mismatch")
	}
}

func TestCloudStorageAssessor_GetByStreaming(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	csa := setupStorageAssessor(t)
	ctx := t.Context()

	content := []byte("streaming assessor test")
	if err := csa.Put(ctx, "stream/data.bin", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	_, rc, err := csa.GetByStreaming(ctx, "stream/data.bin")
	if err != nil {
		t.Fatalf("GetByStreaming: %v", err)
	}
	defer rc.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(rc); err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), content) {
		t.Fatalf("stream data mismatch")
	}
}

func TestCloudStorageAssessor_GetObjectInfo(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	csa := setupStorageAssessor(t)
	ctx := t.Context()

	content := []byte("info assessor test data")
	if err := csa.Put(ctx, "info/obj.txt", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	info, err := csa.GetObjectInfo(ctx, "info/obj.txt")
	if err != nil {
		t.Fatalf("GetObjectInfo: %v", err)
	}
	if info.ContentLength != len(content) {
		t.Fatalf("ContentLength = %d, want %d", info.ContentLength, len(content))
	}
}

func TestCloudStorageAssessor_Delete(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	csa := setupStorageAssessor(t)
	ctx := t.Context()

	if err := csa.Put(ctx, "del/target.txt", bytes.NewReader([]byte("delete"))); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := csa.Delete(ctx, "del/target.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, _, err := csa.Get(ctx, "del/target.txt")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}
