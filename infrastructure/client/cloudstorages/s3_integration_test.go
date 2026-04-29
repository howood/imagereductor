package cloudstorages_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/howood/imagereductor/infrastructure/client/cloudstorages"
	"github.com/testcontainers/testcontainers-go/modules/minio"
)

func setupMinIO(t *testing.T) *cloudstorages.S3Instance {
	t.Helper()

	ctx := t.Context()

	container, err := minio.Run(ctx, "minio/minio:latest",
		minio.WithUsername("minioadmin"),
		minio.WithPassword("minioadmin"),
	)
	if err != nil {
		t.Fatalf("start minio container: %v", err)
	}
	t.Cleanup(func() {
		if termErr := container.Terminate(context.Background()); termErr != nil {
			t.Logf("terminate minio: %v", termErr)
		}
	})

	endpoint, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("minio connection string: %v", err)
	}

	inst, err := cloudstorages.NewS3WithConfig(ctx, cloudstorages.S3Config{
		Region:    "us-east-1",
		Endpoint:  "http://" + endpoint,
		UseLocal:  true,
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		Bucket:    "test-bucket",
		Timeout:   30 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewS3WithConfig: %v", err)
	}
	return inst
}

func TestS3Integration_PutAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupMinIO(t)
	ctx := t.Context()
	bucket := inst.GetBucket()

	content := []byte("hello minio world")
	reader := bytes.NewReader(content)
	if err := inst.Put(ctx, bucket, "test/hello.txt", reader); err != nil {
		t.Fatalf("Put: %v", err)
	}

	contentType, data, err := inst.Get(ctx, bucket, "test/hello.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Fatalf("Get data = %q, want %q", data, content)
	}
	if contentType == "" {
		t.Fatal("Get contentType is empty")
	}
}

func TestS3Integration_GetByStreaming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupMinIO(t)
	ctx := t.Context()
	bucket := inst.GetBucket()

	content := []byte("streaming data test")
	if err := inst.Put(ctx, bucket, "stream/data.bin", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	contentType, rc, err := inst.GetByStreaming(ctx, bucket, "stream/data.bin")
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
	if contentType == "" {
		t.Fatal("contentType empty")
	}
}

func TestS3Integration_GetObjectInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupMinIO(t)
	ctx := t.Context()
	bucket := inst.GetBucket()

	content := []byte("info test data 12345")
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
	if info.ContentType == "" {
		t.Fatal("ContentType empty")
	}
}

func TestS3Integration_List(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupMinIO(t)
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
	if len(names) != 3 {
		t.Fatalf("List len = %d, want 3; got %v", len(names), names)
	}
	for _, k := range keys {
		found := false
		for _, n := range names {
			if n == k {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("List missing key %s", k)
		}
	}
}

func TestS3Integration_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupMinIO(t)
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

func TestS3Integration_PutDetectsContentType(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	inst := setupMinIO(t)
	ctx := t.Context()
	bucket := inst.GetBucket()

	// PNG magic bytes.
	pngData := []byte("\x89PNG\r\n\x1a\n" + strings.Repeat("\x00", 100))
	if err := inst.Put(ctx, bucket, "typed/image.png", bytes.NewReader(pngData)); err != nil {
		t.Fatalf("Put png: %v", err)
	}

	info, err := inst.GetObjectInfo(ctx, bucket, "typed/image.png")
	if err != nil {
		t.Fatalf("GetObjectInfo: %v", err)
	}
	if !strings.HasPrefix(info.ContentType, "image/png") {
		t.Fatalf("ContentType = %q, want image/png prefix", info.ContentType)
	}
}
