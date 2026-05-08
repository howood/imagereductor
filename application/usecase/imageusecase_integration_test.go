package usecase_test

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"testing"
	"time"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/howood/imagereductor/application/actor"
	"github.com/howood/imagereductor/application/actor/storageservice"
	"github.com/howood/imagereductor/application/usecase"
	"github.com/howood/imagereductor/infrastructure/client/cloudstorages"
)

func createTestPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := range 100 {
		for x := range 100 {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

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

func TestImageUsecase_GetImage_WithResize(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	pngData := createTestPNG(t)
	if err := s.csa.Put(ctx, "img/resize.png", bytes.NewReader(pngData)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	opt := actor.ImageOperatorOption{Width: 50, Height: 50}
	contenttype, data, err := s.uc.GetImage(ctx, opt, "img/resize.png")
	if err != nil {
		t.Fatalf("GetImage: %v", err)
	}
	if contenttype != "image/png" {
		t.Fatalf("contenttype = %q, want image/png", contenttype)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty resized image")
	}
}

func TestImageUsecase_ConvertImage(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	pngData := createTestPNG(t)
	reader := bytes.NewReader(pngData)

	opt := actor.ImageOperatorOption{Width: 30, Height: 30}
	result, err := s.uc.ConvertImage(ctx, opt, newFakeMultipartFile(reader))
	if err != nil {
		t.Fatalf("ConvertImage: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty converted image")
	}
}

func TestImageUsecase_UploadToStorage_WithReader(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	content := []byte("uploaded via reader")
	reader := bytes.NewReader(content)

	if err := s.uc.UploadToStorage(ctx, "up/reader.txt", newFakeMultipartFile(reader), nil); err != nil {
		t.Fatalf("UploadToStorage: %v", err)
	}

	_, got, err := s.uc.GetFile(ctx, "up/reader.txt")
	if err != nil {
		t.Fatalf("GetFile: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("data mismatch")
	}
}

// fakeMultipartFile implements multipart.File using bytes.Reader.
type fakeMultipartFile struct {
	*bytes.Reader
}

func newFakeMultipartFile(r *bytes.Reader) *fakeMultipartFile {
	return &fakeMultipartFile{Reader: r}
}

func (f *fakeMultipartFile) Close() error {
	return nil
}

func TestImageUsecase_UploadToStorage_NotReadSeeker(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	err := s.uc.UploadToStorage(ctx, "up/fail.txt", newFakeMultipartFile(bytes.NewReader([]byte("data"))), nil)
	// fakeMultipartFile embeds *bytes.Reader which IS io.ReadSeeker, so this should succeed.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestImageUsecase_GetImage_StorageGetError(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	_, _, err := s.uc.GetImage(ctx, actor.ImageOperatorOption{}, "missing/key.png")
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
}

func TestImageUsecase_GetImage_DecodeError(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	// Put non-image data and request resize so Decode is invoked.
	if err := s.csa.Put(ctx, "img/notimage.bin", bytes.NewReader([]byte("not-an-image"))); err != nil {
		t.Fatalf("Put: %v", err)
	}

	_, _, err := s.uc.GetImage(ctx, actor.ImageOperatorOption{Width: 10, Height: 10}, "img/notimage.bin")
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestImageUsecase_GetImage_ProcessError(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	pngData := createTestPNG(t)
	if err := s.csa.Put(ctx, "img/processfail.png", bytes.NewReader(pngData)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Invalid Rotate value triggers Process error.
	opt := actor.ImageOperatorOption{Rotate: "invalid_value"}
	_, _, err := s.uc.GetImage(ctx, opt, "img/processfail.png")
	if err == nil {
		t.Fatal("expected process error, got nil")
	}
}

// nonReadSeekerFile implements multipart.File but does NOT implement io.ReadSeeker
// (Seek returns an error, but the type assertion to io.ReadSeeker still succeeds
// because multipart.File embeds io.Seeker). Use a wrapper that hides ReadSeeker.
type nonReadSeekerFile struct{}

func (nonReadSeekerFile) Read(_ []byte) (int, error)            { return 0, io.EOF }
func (nonReadSeekerFile) Close() error                          { return nil }
func (nonReadSeekerFile) Seek(_ int64, _ int) (int64, error)    { return 0, errSeekFailed }
func (nonReadSeekerFile) ReadAt(_ []byte, _ int64) (int, error) { return 0, io.EOF }

var errSeekFailed = errors.New("seek failed")

func TestImageUsecase_ConvertImage_NoOption(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	result, err := s.uc.ConvertImage(ctx, actor.ImageOperatorOption{}, newFakeMultipartFile(bytes.NewReader([]byte("data"))))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %d bytes", len(result))
	}
}

func TestImageUsecase_ConvertImage_SeekError(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	opt := actor.ImageOperatorOption{Width: 10, Height: 10}
	_, err := s.uc.ConvertImage(ctx, opt, nonReadSeekerFile{})
	if err == nil {
		t.Fatal("expected seek error, got nil")
	}
}

func TestImageUsecase_ConvertImage_DecodeError(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	opt := actor.ImageOperatorOption{Width: 10, Height: 10}
	_, err := s.uc.ConvertImage(ctx, opt, newFakeMultipartFile(bytes.NewReader([]byte("not-an-image"))))
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestImageUsecase_ConvertImage_ProcessError(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	pngData := createTestPNG(t)
	opt := actor.ImageOperatorOption{Rotate: "invalid_value"}
	_, err := s.uc.ConvertImage(ctx, opt, newFakeMultipartFile(bytes.NewReader(pngData)))
	if err == nil {
		t.Fatal("expected process error, got nil")
	}
}

func TestImageUsecase_UploadToStorage_SeekError(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	s := setupImageUsecaseRaw(t)
	ctx := t.Context()

	err := s.uc.UploadToStorage(ctx, "up/seekfail.txt", nonReadSeekerFile{}, nil)
	if err == nil {
		t.Fatal("expected seek error, got nil")
	}
}
