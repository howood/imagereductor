package cloudstorages

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	extramimetype "github.com/gabriel-vasile/mimetype"
	"github.com/howood/imagereductor/domain/entity"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"google.golang.org/api/iterator"
)

// Sentinel errors (static) for validation (err113 compliant).
var (
	ErrGCSBucketEmpty    = errors.New("gcs bucket name is empty")
	ErrGCSProjectIDEmpty = errors.New("gcs project id is empty")
)

// （後方互換目的で保持していたGCS用のグローバル環境変数参照は廃止し、明示的なconfig取得に統一）

// GCSInstance struct.
// GCSConfig defines configuration for GCSInstance (similar to S3Config for S3).
type GCSConfig struct {
	ProjectID string
	Bucket    string
	Timeout   time.Duration // 0 means no timeout
}

// LoadGCSConfigFromEnv builds config from environment variables.
func LoadGCSConfigFromEnv() GCSConfig {
	return GCSConfig{
		ProjectID: os.Getenv("GCS_PROJECTID"),
		Bucket:    os.Getenv("GCS_BUKET"),
		Timeout:   0,
	}
}

type GCSInstance struct {
	client *storage.Client
	cfg    GCSConfig
}

// NewGCS creates a new GCSInstance.
// NewGCSWithConfig is new constructor returning error.
func NewGCSWithConfig(ctx context.Context, cfg GCSConfig) (*GCSInstance, error) {
	if cfg.Bucket == "" {
		return nil, ErrGCSBucketEmpty
	}
	if cfg.ProjectID == "" {
		return nil, ErrGCSProjectIDEmpty
	}
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("new storage client: %w", err)
	}
	inst := &GCSInstance{client: client, cfg: cfg}
	inst.init(ctx)
	return inst, nil
}

// NewGCS keeps backward compatibility (panic on failure) like previous version.
func NewGCS() *GCSInstance { //nolint:forcetypeassert
	ctx := context.Background()
	inst, err := NewGCSWithConfig(ctx, LoadGCSConfigFromEnv())
	if err != nil {
		panic(err)
	}
	return inst
}

// Put puts to storage.
func (gcsinstance *GCSInstance) Put(ctx context.Context, bucket string, path string, file io.ReadSeeker) error {
	ctx, cancel := gcsinstance.withTimeout(ctx)
	defer cancel()
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read source data: %w", err)
	}
	mimetype := http.DetectContentType(data)
	if mimetype == "" || mimetype == mimeOctetStream {
		mtype := extramimetype.Detect(data)
		log.Debug(ctx, mtype)
		mimetype = mtype.String()
	}
	log.Debug(ctx, mimetype)
	object := gcsinstance.client.Bucket(bucket).Object(path)
	writer := object.NewWriter(ctx)
	writer.ContentType = mimetype
	writer.CacheControl = "no-cache"
	defer writer.Close()
	if _, err = writer.Write(data); err != nil {
		return fmt.Errorf("write object bucket=%s key=%s: %w", bucket, path, err)
	}
	return nil
}

// Get gets from storage.
func (gcsinstance *GCSInstance) Get(ctx context.Context, bucket string, key string) (string, []byte, error) {
	ctx, cancel := gcsinstance.withTimeout(ctx)
	defer cancel()
	log.Debug(ctx, bucket)
	log.Debug(ctx, key)
	reader, err := gcsinstance.client.Bucket(bucket).Object(key).NewReader(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("get object bucket=%s key=%s: %w", bucket, key, err)
	}
	defer reader.Close()
	contenttype := reader.Attrs.ContentType
	response, err := io.ReadAll(reader)
	if err != nil {
		return "", nil, fmt.Errorf("read object bucket=%s key=%s: %w", bucket, key, err)
	}
	return contenttype, response, nil
}

// GetByStreaming gets from storage by streaming.
func (gcsinstance *GCSInstance) GetByStreaming(ctx context.Context, bucket string, key string) (string, io.ReadCloser, error) {
	log.Debug(ctx, bucket)
	log.Debug(ctx, key)
	reader, err := gcsinstance.client.Bucket(bucket).Object(key).NewReader(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("get(stream) bucket=%s key=%s: %w", bucket, key, err)
	}
	// streaming なので Close は呼び出し側責務 (元のバグ: deferで即closeされていた)
	contenttype := reader.Attrs.ContentType
	return contenttype, reader, nil
}

// GetObjectInfo gets from storage.
func (gcsinstance *GCSInstance) GetObjectInfo(ctx context.Context, bucket string, key string) (entity.StorageObjectInfo, error) {
	ctx, cancel := gcsinstance.withTimeout(ctx)
	defer cancel()
	log.Debug(ctx, bucket)
	log.Debug(ctx, key)
	so := entity.StorageObjectInfo{}
	reader, err := gcsinstance.client.Bucket(bucket).Object(key).NewReader(ctx)
	if err != nil {
		return so, fmt.Errorf("head object(bucket=%s key=%s): %w", bucket, key, err)
	}
	defer reader.Close()
	so.ContentType = reader.Attrs.ContentType
	so.ContentLength = int(reader.Attrs.Size)
	return so, nil
}

// List get list from storage.
func (gcsinstance *GCSInstance) List(ctx context.Context, bucket string, key string) ([]string, error) {
	ctx, cancel := gcsinstance.withTimeout(ctx)
	defer cancel()
	log.Debug(ctx, fmt.Sprintf("ListDirectory %s : %s", bucket, key))
	query := &storage.Query{Prefix: key}
	var names []string
	it := gcsinstance.client.Bucket(bucket).Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return names, fmt.Errorf("list objects bucket=%s prefix=%s: %w", bucket, key, err)
		}
		names = append(names, attrs.Name)
	}
	return names, nil
}

// Delete deletes from storage.
func (gcsinstance *GCSInstance) Delete(ctx context.Context, bucket string, key string) error {
	ctx, cancel := gcsinstance.withTimeout(ctx)
	defer cancel()
	if err := gcsinstance.client.Bucket(bucket).Object(key).Delete(ctx); err != nil {
		return fmt.Errorf("delete object bucket=%s key=%s: %w", bucket, key, err)
	}
	return nil
}

// GetBucket returns configured bucket name.
func (gcsinstance *GCSInstance) GetBucket() string {
	return gcsinstance.cfg.Bucket
}

func (gcsinstance *GCSInstance) init(ctx context.Context) {
	bucket := gcsinstance.cfg.Bucket
	if bucket == "" {
		log.Warn(ctx, "bucket name empty; skip init")
		return
	}
	if _, attrErr := gcsinstance.client.Bucket(bucket).Attrs(ctx); attrErr != nil {
		if gcsinstance.cfg.ProjectID == "" { // cannot create without project id
			log.Warn(ctx, "project id empty; cannot create bucket")
			return
		}
		if err := gcsinstance.client.Bucket(bucket).Create(ctx, gcsinstance.cfg.ProjectID, nil); err != nil {
			log.Debug(ctx, "***CreateError****")
			log.Debug(ctx, err)
		}
	}
}

// withTimeout attaches timeout if configured.
func (g *GCSInstance) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if g.cfg.Timeout > 0 {
		return context.WithTimeout(ctx, g.cfg.Timeout)
	}
	return ctx, func() {}
}
