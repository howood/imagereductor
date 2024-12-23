package cloudstorages

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	extramimetype "github.com/gabriel-vasile/mimetype"
	"github.com/howood/imagereductor/domain/entity"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"google.golang.org/api/iterator"
)

// GcsBucketUploadfiles is bucket to upload.
//
//nolint:gochecknoglobals
var GcsBucketUploadfiles = os.Getenv("GCS_BUKET")

// GcsProjectID is GCS Project ID.
//
//nolint:gochecknoglobals
var gcsProjectID = os.Getenv("GCS_PROJECTID")

// GCSInstance struct.
type GCSInstance struct {
	client *storage.Client
}

// NewGCS creates a new GCSInstance.
func NewGCS(ctx context.Context) *GCSInstance {
	log.Debug(ctx, "----GCS DNS----")
	var I *GCSInstance
	client, err := storage.NewClient(ctx)
	if err != nil {
		panic(err)
	}
	I = &GCSInstance{
		client: client,
	}
	I.init(ctx)
	return I
}

func (gcsinstance *GCSInstance) init(ctx context.Context) {
	if _, exitstserr := gcsinstance.client.Bucket(GcsBucketUploadfiles).Attrs(ctx); exitstserr != nil {
		if err := gcsinstance.client.Bucket(GcsBucketUploadfiles).Create(ctx, gcsProjectID, nil); err != nil {
			log.Debug(ctx, "***CreateError****")
			log.Debug(ctx, err)
		}
	}
}

// Put puts to storage.
func (gcsinstance *GCSInstance) Put(ctx context.Context, bucket string, path string, file io.ReadSeeker) error {
	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	mimetype := http.DetectContentType(bytes)
	if mimetype == "" || mimetype == mimeOctetStream {
		mtype := extramimetype.Detect(bytes)
		log.Debug(ctx, mtype)
		mimetype = mtype.String()
	}
	log.Debug(ctx, mimetype)
	object := gcsinstance.client.Bucket(bucket).Object(path)
	writer := object.NewWriter(ctx)

	writer.ContentType = mimetype
	writer.CacheControl = "no-cache"
	defer writer.Close()

	if _, err = writer.Write(bytes); err != nil {
		return err
	}

	return nil
}

// Get gets from storage.
func (gcsinstance *GCSInstance) Get(ctx context.Context, bucket string, key string) (string, []byte, error) {
	log.Debug(ctx, bucket)
	log.Debug(ctx, key)

	reader, err := gcsinstance.client.Bucket(bucket).Object(key).NewReader(ctx)
	if err != nil {
		return "", nil, err
	}
	defer reader.Close()

	contenttype := reader.Attrs.ContentType
	response, err := io.ReadAll(reader)
	if err != nil {
		return "", nil, err
	}

	return contenttype, response, nil
}

// GetByStreaming gets from storage by streaming.
func (gcsinstance *GCSInstance) GetByStreaming(ctx context.Context, bucket string, key string) (string, io.ReadCloser, error) {
	log.Debug(ctx, bucket)
	log.Debug(ctx, key)

	reader, err := gcsinstance.client.Bucket(bucket).Object(key).NewReader(ctx)
	if err != nil {
		return "", nil, err
	}
	defer reader.Close()

	contenttype := reader.Attrs.ContentType
	return contenttype, reader, nil
}

// GetObjectInfo gets from storage.
func (gcsinstance *GCSInstance) GetObjectInfo(ctx context.Context, bucket string, key string) (entity.StorageObjectInfo, error) {
	log.Debug(ctx, bucket)
	log.Debug(ctx, key)
	so := entity.StorageObjectInfo{}
	reader, err := gcsinstance.client.Bucket(bucket).Object(key).NewReader(ctx)
	if err != nil {
		return so, err
	}
	defer reader.Close()

	so.ContentType = reader.Attrs.ContentType
	so.ContentLength = int(reader.Attrs.Size)
	return so, nil
}

// List get list from storage.
func (gcsinstance *GCSInstance) List(ctx context.Context, bucket string, key string) ([]string, error) {
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
			return names, err
		}
		names = append(names, attrs.Name)
	}
	return names, nil
}

// Delete deletes from storage.
func (gcsinstance *GCSInstance) Delete(ctx context.Context, bucket string, key string) error {
	err := gcsinstance.client.Bucket(bucket).Object(key).Delete(ctx)
	return err
}
