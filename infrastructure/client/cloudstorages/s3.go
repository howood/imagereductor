package cloudstorages

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	extramimetype "github.com/gabriel-vasile/mimetype"
	"github.com/howood/imagereductor/domain/entity"
	log "github.com/howood/imagereductor/infrastructure/logger"
)

// S3BucketUploadfiles is bucket to upload.
//
//nolint:gochecknoglobals
var S3BucketUploadfiles = os.Getenv("AWS_S3_BUKET")

// S3Instance struct.
type S3Instance struct {
	client *s3.Client
}

// NewS3 creates a new S3Instance.
func NewS3() *S3Instance {
	ctx := context.Background()
	log.Debug(ctx, "----S3 DNS----")
	log.Debug(ctx, os.Getenv("AWS_S3_REGION"))
	log.Debug(ctx, os.Getenv("AWS_S3_ENDPOINT"))

	var cfg aws.Config
	var err error

	configOptions := []func(*config.LoadOptions) error{
		config.WithRegion(os.Getenv("AWS_S3_REGION")),
	}

	// 認証情報の設定
	if os.Getenv("AWS_S3_ACCESSKEY") != "" && os.Getenv("AWS_S3_SECRETKEY") != "" {
		configOptions = append(configOptions, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				os.Getenv("AWS_S3_ACCESSKEY"),
				os.Getenv("AWS_S3_SECRETKEY"),
				"",
			),
		))
	}

	cfg, err = config.LoadDefaultConfig(ctx, configOptions...)
	if err != nil {
		log.Debug(ctx, "Failed to load AWS config:")
		log.Debug(ctx, err)
		panic(err)
	}

	var client *s3.Client

	if os.Getenv("AWS_S3_LOCALUSE") != "" {
		log.Debug(ctx, "-----use local-----")
		client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(os.Getenv("AWS_S3_ENDPOINT"))
			o.UsePathStyle = true
		})
	} else {
		client = s3.NewFromConfig(cfg)
	}

	instance := &S3Instance{client: client}
	instance.init(ctx)
	return instance
}

func (s3instance *S3Instance) init(ctx context.Context) {
	_, bucketerr := s3instance.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(S3BucketUploadfiles),
	})
	if bucketerr != nil {
		result, err := s3instance.client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(S3BucketUploadfiles),
		})
		if err != nil {
			log.Debug(ctx, "***CreateError****")
			log.Debug(ctx, err)
			log.Debug(ctx, result)
		}
	}
}

// Put puts to storage.
func (s3instance *S3Instance) Put(ctx context.Context, bucket string, path string, file io.ReadSeeker) error {
	// ファイルのオフセットを先頭に戻す
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	mimetype, errfile := s3instance.getContentType(ctx, file)
	if errfile != nil {
		return errfile
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	result, err := s3instance.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(path),
		Body:        file,
		ContentType: aws.String(mimetype),
	})
	log.Debug(ctx, result)
	return err
}

// Get gets from storage.
func (s3instance *S3Instance) Get(ctx context.Context, bucket string, key string) (string, []byte, error) {
	log.Debug(ctx, bucket)
	log.Debug(ctx, key)

	response, err := s3instance.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", nil, err
	}

	contenttype := ""
	if response.ContentType != nil {
		contenttype = *response.ContentType
	}
	defer response.Body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, response.Body); err != nil {
		return "", nil, err
	}
	log.Debug(ctx, contenttype)
	return contenttype, buf.Bytes(), nil
}

// GetByStreaming gets from storage by streaming.
func (s3instance *S3Instance) GetByStreaming(ctx context.Context, bucket string, key string) (string, io.ReadCloser, error) {
	log.Debug(ctx, bucket)
	log.Debug(ctx, key)

	response, err := s3instance.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", nil, err
	}

	contenttype := ""
	if response.ContentType != nil {
		contenttype = *response.ContentType
	}
	log.Debug(ctx, contenttype)
	return contenttype, response.Body, nil
}

// GetObjectInfo gets from storage.
func (s3instance *S3Instance) GetObjectInfo(ctx context.Context, bucket string, key string) (entity.StorageObjectInfo, error) {
	log.Debug(ctx, bucket)
	log.Debug(ctx, key)

	so := entity.StorageObjectInfo{}
	response, err := s3instance.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return so, err
	}

	if response.ContentType != nil {
		so.ContentType = *response.ContentType
	}
	if response.ContentLength != nil {
		so.ContentLength = int(*response.ContentLength)
	}
	return so, nil
}

// List get list from storage.
func (s3instance *S3Instance) List(ctx context.Context, bucket string, key string) ([]string, error) {
	log.Debug(ctx, fmt.Sprintf("ListDirectory %s : %s", bucket, key))
	if key[0:1] == "/" {
		key = key[1:]
	}

	var names []string
	paginator := s3.NewListObjectsV2Paginator(s3instance.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return names, err
		}

		for _, obj := range page.Contents {
			if obj.Key == nil {
				continue
			}
			names = append(names, *obj.Key)
		}
	}

	return names, nil
}

// Delete deletes from storage.
func (s3instance *S3Instance) Delete(ctx context.Context, bucket string, key string) error {
	result, err := s3instance.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	log.Debug(ctx, result)
	return err
}

func (s3instance *S3Instance) getContentType(ctx context.Context, out io.ReadSeeker) (string, error) {
	//nolint:mnd
	buffer := make([]byte, 512)
	_, err := out.Read(buffer)
	if err != nil {
		log.Warn(ctx, "Date Read Error!")
		log.Warn(ctx, err)
	}

	contentType := http.DetectContentType(buffer)
	if contentType == "" || contentType == mimeOctetStream {
		if _, err := out.Seek(0, io.SeekStart); err != nil {
			return "", err
		}
		if mtype, err := extramimetype.DetectReader(out); err == nil {
			log.Debug(ctx, mtype)
			contentType = mtype.String()
		}
	}
	log.Debug(ctx, contentType)
	return contentType, nil
}
