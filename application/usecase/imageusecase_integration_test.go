package usecase_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/application/actor/storageservice"
	"github.com/howood/imagereductor/application/usecase"
	"github.com/howood/imagereductor/infrastructure/client/cloudstorages"
)

type testSetup struct {
	uc  *usecase.ImageUsecase
	csa *storageservice.CloudStorageAssessor
}

func setupImageUsecaseRaw(t *testing.T) testSetup {
	t.Helper()

	const (
		bucket    = "imageuc-raw-test"
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

	csa := storageservice.NewCloudStorageAssessorForTest(gcsInst)
	uc := usecase.NewImageUsecaseForTest(csa)
	return testSetup{uc: uc, csa: csa}
}

func TestImageUsecase_GetFile(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	content := []byte("file content for GetFile test")
	if err := s.csa.Put(ctx, "getfile/doc.txt", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	_, data, err := s.uc.GetFile(ctx, "getfile/doc.txt")
	if err != nil {
		t.Fatalf("GetFile: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Fatalf("data mismatch")
	}
}

func TestImageUsecase_GetFileStream(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	content := []byte("stream content test")
	if err := s.csa.Put(ctx, "stream/file.bin", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	_, contentLength, rc, err := s.uc.GetFileStream(ctx, "stream/file.bin")
	if err != nil {
		t.Fatalf("GetFileStream: %v", err)
	}
	defer rc.Close()

	if contentLength != len(content) {
		t.Fatalf("contentLength = %d, want %d", contentLength, len(content))
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(rc); err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), content) {
		t.Fatalf("stream data mismatch")
	}
}

func TestImageUsecase_GetFileInfo(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	content := []byte("info test data abc")
	if err := s.csa.Put(ctx, "info/meta.txt", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	info, err := s.uc.GetFileInfo(ctx, "info/meta.txt")
	if err != nil {
		t.Fatalf("GetFileInfo: %v", err)
	}
	if info.ContentLength != len(content) {
		t.Fatalf("ContentLength = %d, want %d", info.ContentLength, len(content))
	}
}

func TestImageUsecase_GetImage_NoResize(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	content := []byte("raw image bytes")
	if err := s.csa.Put(ctx, "img/raw.dat", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// No resize option (zero-value ImageOperatorOption).
	_, data, err := s.uc.GetImage(ctx, actor.ImageOperatorOption{}, "img/raw.dat")
	if err != nil {
		t.Fatalf("GetImage: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Fatalf("data mismatch")
	}
}

func TestImageUsecase_UploadToStorage_WithBytes(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	data := []byte("uploaded via bytes")
	if err := s.uc.UploadToStorage(ctx, "up/bytes.txt", nil, data); err != nil {
		t.Fatalf("UploadToStorage: %v", err)
	}

	_, got, err := s.uc.GetFile(ctx, "up/bytes.txt")
	if err != nil {
		t.Fatalf("GetFile: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("data mismatch")
	}
}
