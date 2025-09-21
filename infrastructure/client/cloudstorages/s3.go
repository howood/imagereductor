package cloudstorages

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	extramimetype "github.com/gabriel-vasile/mimetype"
	"github.com/howood/imagereductor/domain/entity"
	log "github.com/howood/imagereductor/infrastructure/logger"
)

// （後方互換目的で残されていたグローバルバケット変数は廃止。config経由に統一）

// S3Config defines configuration for S3Instance.
type S3Config struct {
	Region    string
	Endpoint  string
	UseLocal  bool
	AccessKey string
	SecretKey string
	Bucket    string
	Timeout   time.Duration // 0 means no timeout
}

// LoadS3ConfigFromEnv builds config from environment variables (backward compatibility helper).
func LoadS3ConfigFromEnv() S3Config {
	return S3Config{
		Region:    os.Getenv("AWS_S3_REGION"),
		Endpoint:  os.Getenv("AWS_S3_ENDPOINT"),
		UseLocal:  os.Getenv("AWS_S3_LOCALUSE") != "",
		AccessKey: os.Getenv("AWS_S3_ACCESSKEY"),
		SecretKey: os.Getenv("AWS_S3_SECRETKEY"),
		Bucket:    os.Getenv("AWS_S3_BUKET"),
		Timeout:   0,
	}
}

// S3Instance struct.
type S3Instance struct {
	client *s3.Client
	cfg    S3Config
}

// GetBucket returns configured bucket name.
func (s3instance *S3Instance) GetBucket() string {
	return s3instance.cfg.Bucket
}

// NewS3WithConfig is the new constructor returning error.
func NewS3WithConfig(ctx context.Context, cfgIn S3Config) (*S3Instance, error) { //nolint:cyclop
	if cfgIn.Bucket == "" {
		return nil, errors.New("s3 bucket name is empty")
	}
	log.Debug(ctx, "----S3 DNS----")
	log.Debug(ctx, cfgIn.Region)
	log.Debug(ctx, cfgIn.Endpoint)
	opts := []func(*config.LoadOptions) error{}
	if cfgIn.Region != "" {
		opts = append(opts, config.WithRegion(cfgIn.Region))
	}
	if cfgIn.AccessKey != "" && cfgIn.SecretKey != "" {
		opts = append(opts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfgIn.AccessKey, cfgIn.SecretKey, "")))
	}
	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	var client *s3.Client
	if cfgIn.UseLocal {
		log.Debug(ctx, "-----use local-----")
		client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			if cfgIn.Endpoint != "" {
				o.BaseEndpoint = aws.String(cfgIn.Endpoint)
			}
			o.UsePathStyle = true
		})
	} else {
		client = s3.NewFromConfig(awsCfg)
	}
	inst := &S3Instance{client: client, cfg: cfgIn}
	inst.init(ctx)
	return inst, nil
}

// NewS3 keeps backward compatible panic behavior.
func NewS3() *S3Instance { //nolint:forcetypeassert
	ctx := context.Background()
	inst, err := NewS3WithConfig(ctx, LoadS3ConfigFromEnv())
	if err != nil {
		panic(err)
	}
	return inst
}

// withTimeout attaches timeout if configured.
func (s *S3Instance) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if s.cfg.Timeout > 0 {
		return context.WithTimeout(ctx, s.cfg.Timeout)
	}
	return ctx, func() {}
}

// Put puts to storage.
func (s3instance *S3Instance) Put(ctx context.Context, bucket, path string, file io.ReadSeeker) error {
	ctx, cancel := s3instance.withTimeout(ctx)
	defer cancel()
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("put seek start: %w", err)
	}
	mimetype, err := s3instance.getContentType(ctx, file)
	if err != nil {
		return fmt.Errorf("detect content type: %w", err)
	}
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

// Get gets from storage and returns bytes.
func (s3instance *S3Instance) Get(ctx context.Context, bucket, key string) (string, []byte, error) {
	ctx, cancel := s3instance.withTimeout(ctx)
	defer cancel()
	log.Debug(ctx, bucket)
	log.Debug(ctx, key)
	response, err := s3instance.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
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

// GetByStreaming gets from storage by streaming (no timeout).
func (s3instance *S3Instance) GetByStreaming(ctx context.Context, bucket, key string) (string, io.ReadCloser, error) {
	log.Debug(ctx, bucket)
	log.Debug(ctx, key)
	response, err := s3instance.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
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

// GetObjectInfo gets object head metadata.
func (s3instance *S3Instance) GetObjectInfo(ctx context.Context, bucket, key string) (entity.StorageObjectInfo, error) {
	ctx, cancel := s3instance.withTimeout(ctx)
	defer cancel()
	log.Debug(ctx, bucket)
	log.Debug(ctx, key)
	so := entity.StorageObjectInfo{}
	response, err := s3instance.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
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
func (s3instance *S3Instance) List(ctx context.Context, bucket, key string) ([]string, error) {
	ctx, cancel := s3instance.withTimeout(ctx)
	defer cancel()
	log.Debug(ctx, fmt.Sprintf("ListDirectory %s : %s", bucket, key))
	if key != "" {
		key = strings.TrimPrefix(key, "/")
	}
	var names []string
	paginator := s3.NewListObjectsV2Paginator(s3instance.client, &s3.ListObjectsV2Input{Bucket: aws.String(bucket), Prefix: aws.String(key)})
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
func (s3instance *S3Instance) Delete(ctx context.Context, bucket, key string) error {
	ctx, cancel := s3instance.withTimeout(ctx)
	defer cancel()
	result, err := s3instance.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	if err != nil {
		return fmt.Errorf("delete object bucket=%s key=%s: %w", bucket, key, err)
	}
	log.Debug(ctx, result)
	return nil
}

func (s3instance *S3Instance) getContentType(ctx context.Context, rs io.ReadSeeker) (string, error) {
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

func (s3instance *S3Instance) init(ctx context.Context) { // bucket存在確認と必要なら作成
	bucket := s3instance.cfg.Bucket
	if bucket == "" { // config上未設定なら何もしない(後方互換のenv fallbackは削除)
		log.Warn(ctx, "bucket name empty; skip init")
		return
	}
	if _, err := s3instance.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)}); err != nil {
		result, createErr := s3instance.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
		if createErr != nil {
			log.Debug(ctx, "***CreateError****")
			log.Debug(ctx, createErr)
			log.Debug(ctx, result)
		}
	}
}
