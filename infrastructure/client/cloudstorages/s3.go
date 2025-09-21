package cloudstorages

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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
	region := os.Getenv("AWS_S3_REGION")
	endpoint := os.Getenv("AWS_S3_ENDPOINT")
	log.Debug(ctx, "----S3 DNS----")
	log.Debug(ctx, region)
	log.Debug(ctx, endpoint)

	var cfg aws.Config
	var err error

	configOptions := []func(*config.LoadOptions) error{}

	if region != "" {
		configOptions = append(configOptions, config.WithRegion(region))
	}

	// 認証情報の設定
	accessKey := os.Getenv("AWS_S3_ACCESSKEY")
	secretKey := os.Getenv("AWS_S3_SECRETKEY")
	if accessKey != "" && secretKey != "" {
		configOptions = append(configOptions, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
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
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
	} else {
		client = s3.NewFromConfig(cfg)
	}

	instance := &S3Instance{client: client}
	instance.init(ctx)
	return instance
}

// Put puts to storage.
func (s3instance *S3Instance) Put(ctx context.Context, bucket string, path string, file io.ReadSeeker) error {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("put seek start: %w", err)
	}

	mimetype, err := s3instance.getContentType(ctx, file)
	if err != nil {
		return fmt.Errorf("detect content type: %w", err)
	}

	// getContentType 内で先頭へ戻しているため再Seek不要
	result, err := s3instance.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(path),
		Body:        file,
		ContentType: aws.String(mimetype),
	})
	if err != nil {
		return fmt.Errorf("put object bucket=%s key=%s: %w", bucket, path, err)
	}
	log.Debug(ctx, result)
	return nil
}

// Get gets from storage.
func (s3instance *S3Instance) Get(ctx context.Context, bucket string, key string) (string, []byte, error) { // バイト配列取得
	log.Debug(ctx, bucket)
	log.Debug(ctx, key)

	response, err := s3instance.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", nil, fmt.Errorf("get object bucket=%s key=%s: %w", bucket, key, err)
	}

	contenttype := ""
	if response.ContentType != nil {
		contenttype = *response.ContentType
	}
	defer response.Body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, response.Body); err != nil {
		return "", nil, fmt.Errorf("read object body bucket=%s key=%s: %w", bucket, key, err)
	}
	log.Debug(ctx, contenttype)
	return contenttype, buf.Bytes(), nil
}

// GetByStreaming gets from storage by streaming.
func (s3instance *S3Instance) GetByStreaming(ctx context.Context, bucket string, key string) (string, io.ReadCloser, error) { // ストリーミング取得
	log.Debug(ctx, bucket)
	log.Debug(ctx, key)

	response, err := s3instance.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", nil, fmt.Errorf("get object(stream) bucket=%s key=%s: %w", bucket, key, err)
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
		return so, fmt.Errorf("head object bucket=%s key=%s: %w", bucket, key, err)
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
	if key != "" {
		key = strings.TrimPrefix(key, "/")
	}

	var names []string
	paginator := s3.NewListObjectsV2Paginator(s3instance.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return names, fmt.Errorf("list objects bucket=%s prefix=%s: %w", bucket, key, err)
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
	if err != nil {
		return fmt.Errorf("delete object bucket=%s key=%s: %w", bucket, key, err)
	}
	log.Debug(ctx, result)
	return nil
}

func (s3instance *S3Instance) getContentType(ctx context.Context, rs io.ReadSeeker) (string, error) {
	// 常に先頭から判定し、終了後は先頭へ戻す（呼び出し側の余計なSeekを不要に）
	if _, err := rs.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("content-type seek start: %w", err)
	}
	//nolint:mnd // 512 bytes for http.DetectContentType
	buf := make([]byte, 512)
	n, err := rs.Read(buf)
	if err != nil && err != io.EOF { // EOFは許容
		log.Warn(ctx, "data read error for content-type detection")
		log.Warn(ctx, err)
	}
	buf = buf[:n]
	contentType := http.DetectContentType(buf)
	if contentType == "" || contentType == mimeOctetStream {
		if _, err := rs.Seek(0, io.SeekStart); err != nil {
			return "", fmt.Errorf("content-type reseek: %w", err)
		}
		if mtype, err := extramimetype.DetectReader(rs); err == nil {
			log.Debug(ctx, mtype)
			contentType = mtype.String()
		}
	}
	if _, err := rs.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("content-type final seek: %w", err)
	}
	log.Debug(ctx, contentType)
	return contentType, nil
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
