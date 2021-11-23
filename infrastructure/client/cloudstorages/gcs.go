package cloudstorages

import (
	"fmt"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"

	"io"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/howood/imagereductor/infrastructure/logger"
)

// GcsBucketUploadfiles is bucket to upload
var GcsBucketUploadfiles = os.Getenv("GCS_BUKET")

// GcsProjectID is GCS Project ID
var GcsProjectID = os.Getenv("GcsProjectID")

// GCSInstance struct
type GCSInstance struct {
	client *storage.Client
	ctx    context.Context
}

// NewGCS creates a new GCSInstance
func NewGCS(ctx context.Context) *GCSInstance {
	log.Debug(ctx, "----GCS DNS----")
	var I *GCSInstance
	client, err := storage.NewClient(ctx)
	if err != nil {
		panic(err)
	}
	I = &GCSInstance{
		client: client,
		ctx:    ctx,
	}
	I.init()
	return I
}

func (gcsinstance *GCSInstance) init() {
	if _, exitstserr := gcsinstance.client.Bucket(GcsBucketUploadfiles).Attrs(gcsinstance.ctx); exitstserr != nil {
		if err := gcsinstance.client.Bucket(GcsBucketUploadfiles).Create(gcsinstance.ctx, GcsProjectID, nil); err != nil {
			log.Debug(gcsinstance.ctx, "***CreateError****")
			log.Debug(gcsinstance.ctx, err)
		}
	}
}

// Put puts to storage
func (gcsinstance *GCSInstance) Put(bucket string, path string, file io.ReadSeeker) error {
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	mimetype := http.DetectContentType(bytes)

	object := gcsinstance.client.Bucket(bucket).Object(path)
	writer := object.NewWriter(gcsinstance.ctx)

	writer.ContentType = mimetype
	writer.CacheControl = "no-cache"
	defer writer.Close()

	if _, err = writer.Write(bytes); err != nil {
		return err
	}

	return nil
}

// Get gets from storage
func (gcsinstance *GCSInstance) Get(bucket string, key string) (string, []byte, error) {
	log.Debug(gcsinstance.ctx, bucket)
	log.Debug(gcsinstance.ctx, key)

	reader, err := gcsinstance.client.Bucket(bucket).Object(key).NewReader(gcsinstance.ctx)
	if err != nil {
		return "", nil, err
	}
	defer reader.Close()

	contenttype := reader.ContentType()
	// CloudStorage上のObjectの、コンテンツの読み込み
	response, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", nil, err
	}

	return contenttype, response, nil
}

// GetByStreaming gets stream from storage
func (gcsinstance *GCSInstance) GetByStreaming(bucket string, key string) (string, io.ReadCloser, error) {
	log.Debug(gcsinstance.ctx, bucket)
	log.Debug(gcsinstance.ctx, key)

	reader, err := gcsinstance.client.Bucket(bucket).Object(key).NewReader(gcsinstance.ctx)
	if err != nil {
		return "", nil, err
	}
	defer reader.Close()

	contenttype := reader.ContentType()
	return contenttype, reader, nil
}

// List get list from storage
func (gcsinstance *GCSInstance) List(bucket string, key string) ([]string, error) {
	log.Debug(gcsinstance.ctx, fmt.Sprintf("ListDirectory %s : %s", bucket, key))
	query := &storage.Query{Prefix: key}
	var names []string
	it := gcsinstance.client.Bucket(bucket).Objects(gcsinstance.ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return names, err
		}
		names = append(names, attrs.Name)
	}
	return names, nil
}

// Delete deletes from storage
func (gcsinstance *GCSInstance) Delete(bucket string, key string) error {
	err := gcsinstance.client.Bucket(bucket).Object(key).Delete(gcsinstance.ctx)
	return err
}
