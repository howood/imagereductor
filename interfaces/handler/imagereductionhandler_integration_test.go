package handler_test

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/howood/imagereductor/application/actor/storageservice"
	"github.com/howood/imagereductor/application/usecase"
	"github.com/howood/imagereductor/di/uccluster"
	"github.com/howood/imagereductor/infrastructure/client/cloudstorages"
	"github.com/howood/imagereductor/interfaces/handler"
	"github.com/labstack/echo/v5"
)

type handlerTestEnv struct {
	handler *handler.ImageReductionHandler
	csa     *storageservice.CloudStorageAssessor
}

func setupHandlerEnv(t *testing.T) handlerTestEnv {
	t.Helper()

	t.Setenv("CACHE_TYPE", "gocache")
	t.Setenv("CACHEEXPIED", "60")
	t.Setenv("HEADEREXPIRED", "300")

	const (
		bucket    = "handler-test"
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
	imageUC := usecase.NewImageUsecaseForTest(csa)

	cacheUC, err := usecase.NewCacheUsecaseWithConfig(ctx)
	if err != nil {
		t.Fatalf("NewCacheUsecaseWithConfig: %v", err)
	}

	cluster := &uccluster.UsecaseCluster{
		CacheUC: cacheUC,
		ImageUC: imageUC,
		TokenUC: usecase.NewTokenUsecase(),
	}

	irh := handler.NewImageReductionHandler(handler.BaseHandler{UcCluster: cluster})
	return handlerTestEnv{handler: irh, csa: csa}
}

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

func TestImageReductionHandler_Request(t *testing.T) { //nolint:paralleltest
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	env := setupHandlerEnv(t)
	ctx := t.Context()

	// Upload a test PNG.
	pngData := createTestPNG(t)
	if err := env.csa.Put(ctx, "img/test.png", bytes.NewReader(pngData)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/image?key=img/test.png", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := env.handler.Request(c); err != nil {
		t.Fatalf("Request: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("expected non-empty body")
	}
}

func TestImageReductionHandler_Request_MissingKey(t *testing.T) { //nolint:paralleltest
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	env := setupHandlerEnv(t)

	e := echo.New()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/image", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := env.handler.Request(c); err != nil {
		t.Fatalf("Request: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestImageReductionHandler_Request_WithResize(t *testing.T) { //nolint:paralleltest
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	env := setupHandlerEnv(t)
	ctx := t.Context()

	pngData := createTestPNG(t)
	if err := env.csa.Put(ctx, "img/resize.png", bytes.NewReader(pngData)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/image?key=img/resize.png&w=50&h=50", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := env.handler.Request(c); err != nil {
		t.Fatalf("Request: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestImageReductionHandler_Request_CacheHit(t *testing.T) { //nolint:paralleltest
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	env := setupHandlerEnv(t)
	ctx := t.Context()

	pngData := createTestPNG(t)
	if err := env.csa.Put(ctx, "img/cached.png", bytes.NewReader(pngData)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	e := echo.New()
	// First request to populate cache.
	req1 := httptest.NewRequestWithContext(ctx, http.MethodGet, "/image?key=img/cached.png", nil)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := env.handler.Request(c1); err != nil {
		t.Fatalf("first Request: %v", err)
	}

	// Second request should hit cache.
	req2 := httptest.NewRequestWithContext(ctx, http.MethodGet, "/image?key=img/cached.png", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := env.handler.Request(c2); err != nil {
		t.Fatalf("second Request: %v", err)
	}
	if rec2.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec2.Code)
	}
}

func TestImageReductionHandler_RequestFile(t *testing.T) { //nolint:paralleltest
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	env := setupHandlerEnv(t)
	ctx := t.Context()

	content := []byte("hello file content")
	if err := env.csa.Put(ctx, "files/doc.txt", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/file?key=files/doc.txt", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := env.handler.RequestFile(c); err != nil {
		t.Fatalf("RequestFile: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !bytes.Equal(rec.Body.Bytes(), content) {
		t.Fatalf("body mismatch: got %q", rec.Body.String())
	}
}

func TestImageReductionHandler_RequestStreaming(t *testing.T) { //nolint:paralleltest
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	env := setupHandlerEnv(t)
	ctx := t.Context()

	content := []byte("streaming data content - extra payload for differentiation")
	if err := env.csa.Put(ctx, "stream/data.bin", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/stream?key=stream/data.bin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := env.handler.RequestStreaming(c); err != nil {
		t.Fatalf("RequestStreaming: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Body.Len() != len(content) {
		t.Fatalf("body length = %d, want %d", rec.Body.Len(), len(content))
	}
}

func TestImageReductionHandler_RequestInfo(t *testing.T) { //nolint:paralleltest
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	env := setupHandlerEnv(t)
	ctx := t.Context()

	content := []byte("info content data")
	if err := env.csa.Put(ctx, "info/item.dat", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/info?key=info/item.dat", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := env.handler.RequestInfo(c); err != nil {
		t.Fatalf("RequestInfo: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json parse: %v", err)
	}
}

func TestImageReductionHandler_UploadFile(t *testing.T) { //nolint:paralleltest
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	env := setupHandlerEnv(t)

	// Create multipart form.
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("uploadfile", "test.txt")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	fileContent := []byte("uploaded file content")
	if _, err := part.Write(fileContent); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := writer.WriteField("path", "upload/test.txt"); err != nil {
		t.Fatalf("WriteField: %v", err)
	}
	writer.Close()

	e := echo.New()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/upload-file", &body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := env.handler.UploadFile(c); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	// Verify file was uploaded.
	_, data, getErr := env.csa.Get(t.Context(), "upload/test.txt")
	if getErr != nil {
		t.Fatalf("Get uploaded file: %v", getErr)
	}
	if !bytes.Equal(data, fileContent) {
		t.Fatalf("uploaded content mismatch")
	}
}

func TestImageReductionHandler_Upload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Setenv("VALIDATE_IMAGE_TYPE", "png,jpeg")
	t.Setenv("VALIDATE_IMAGE_MAXWIDTH", "5000")
	t.Setenv("VALIDATE_IMAGE_MAXHEIGHT", "5000")
	t.Setenv("VALIDATE_IMAGE_MAXFILESIZE", "10485760")

	env := setupHandlerEnv(t)

	pngData := createTestPNG(t)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("uploadfile", "test.png")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(pngData); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := writer.WriteField("path", "upload/img.png"); err != nil {
		t.Fatalf("WriteField: %v", err)
	}
	writer.Close()

	e := echo.New()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := env.handler.Upload(c); err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}
}

func TestImageReductionHandler_Upload_PathTraversal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Setenv("VALIDATE_IMAGE_TYPE", "png,jpeg")

	env := setupHandlerEnv(t)

	pngData := createTestPNG(t)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("uploadfile", "test.png")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(pngData); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := writer.WriteField("path", "../etc/malicious.png"); err != nil {
		t.Fatalf("WriteField: %v", err)
	}
	writer.Close()

	e := echo.New()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := env.handler.Upload(c); err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for path traversal", rec.Code)
	}
}

func TestImageReductionHandler_UploadFile_PathTraversal(t *testing.T) { //nolint:paralleltest
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	env := setupHandlerEnv(t)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("uploadfile", "test.txt")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write([]byte("content")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := writer.WriteField("path", "../../secret/file.txt"); err != nil {
		t.Fatalf("WriteField: %v", err)
	}
	writer.Close()

	e := echo.New()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/upload-file", &body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := env.handler.UploadFile(c); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for path traversal", rec.Code)
	}
}
