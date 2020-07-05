package cloudstorages

import (
	"cloud.google.com/go/storage"
	"golang.org/x/net/context"

	"io"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/howood/imagereductor/infrastructure/logger"
)

var GCS_BUCKET_UPLOADFILES = os.Getenv("GCS_BUKET")
var GCS_PROJECTID = os.Getenv("GCS_PROJECTID")

type GCSInstance struct {
	client *storage.Client
	ctx    context.Context
}

// インスタンス作成用のメソッド
func NewGCS(ctx context.Context) *GCSInstance {
	log.Debug(ctx, "----GCS DNS----")
	var I *GCSInstance
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil
	}
	I = &GCSInstance{
		client: client,
		ctx:    ctx,
	}
	I.init()
	return I
}

func (gcsinstance GCSInstance) init() {
	if _, exitstserr := gcsinstance.client.Bucket(GCS_BUCKET_UPLOADFILES).Attrs(gcsinstance.ctx); exitstserr != nil {
		if err := gcsinstance.client.Bucket(GCS_BUCKET_UPLOADFILES).Create(gcsinstance.ctx, GCS_PROJECTID, nil); err != nil {
			log.Debug(gcsinstance.ctx, "***CreateError****")
			log.Debug(gcsinstance.ctx, err)
		}
	}
}

func (gcsinstance GCSInstance) Put(bucket string, path string, file io.ReadSeeker) error {
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

func (gcsinstance GCSInstance) Get(bucket string, key string) (string, []byte, error) {
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

func (gcsinstance GCSInstance) Delete(bucket string, key string) error {
	err := gcsinstance.client.Bucket(bucket).Object(key).Delete(gcsinstance.ctx)
	return err
}
