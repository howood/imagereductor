package cloudstorages

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"golang.org/x/net/context"

	"bytes"
	"io"
	"net/http"
	"os"

	log "github.com/howood/imagereductor/infrastructure/logger"
)

var S3BucketUploadfiles = os.Getenv("AWS_S3_BUKET")

// S3Instance struct
type S3Instance struct {
	client *s3.S3
	ctx    context.Context
}

// downloader struct
type downloader struct {
	*s3manager.Downloader
	bucket string
	file   string
	dir    string
}

// NewS3 creates a new S3Instance
func NewS3(ctx context.Context) *S3Instance {
	log.Debug(ctx, "----S3 DNS----")
	log.Debug(ctx, os.Getenv("AWS_S3_REGION"))
	log.Debug(ctx, os.Getenv("AWS_S3_ENDPOINT"))
	//	log.Debug(ctx, os.Getenv("AWS_S3_ACCESSKEY"))
	//	log.Debug(ctx, os.Getenv("AWS_S3_SECRETKEY"))
	var I *S3Instance
	var cred *credentials.Credentials
	if os.Getenv("AWS_S3_ACCESSKEY") != "" && os.Getenv("AWS_S3_SECRETKEY") != "" {
		cred = credentials.NewStaticCredentials(os.Getenv("AWS_S3_ACCESSKEY"), os.Getenv("AWS_S3_SECRETKEY"), "")
	}
	if os.Getenv("AWS_S3_LOCALUSE") != "" {
		log.Debug(ctx, "-----use local-----")
		I = &S3Instance{
			client: s3.New(session.New(), &aws.Config{
				Region:           aws.String(os.Getenv("AWS_S3_REGION")),
				Endpoint:         aws.String(os.Getenv("AWS_S3_ENDPOINT")),
				Credentials:      cred,
				DisableSSL:       aws.Bool(true),
				S3ForcePathStyle: aws.Bool(true),
			}),
			ctx: ctx,
		}
	} else {
		I = &S3Instance{
			client: s3.New(session.New(), &aws.Config{
				Region:      aws.String(os.Getenv("AWS_S3_REGION")),
				Credentials: cred,
			}),
			ctx: ctx,
		}
	}
	I.init()
	return I
}

func (s3instance S3Instance) init() {
	if _, bucketerr := s3instance.client.HeadBucket(&s3.HeadBucketInput{Bucket: aws.String(S3BucketUploadfiles)}); bucketerr != nil {
		if result, err := s3instance.client.CreateBucket(&s3.CreateBucketInput{Bucket: aws.String(S3BucketUploadfiles)}); err != nil {
			log.Debug(s3instance.ctx, "***CreateError****")
			log.Debug(s3instance.ctx, err)
			log.Debug(s3instance.ctx, result)
		}
	}
}

// Put puts to storage
func (s3instance S3Instance) Put(bucket string, path string, file io.ReadSeeker) error {
	//ファイルのオフセットを先頭に戻す
	file.Seek(0, os.SEEK_SET)
	mimetype, errfile := s3instance.getContentType(file)
	if errfile != nil {
		return errfile
	}
	file.Seek(0, os.SEEK_SET)
	result, err := s3instance.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(path),
		Body:        file,
		ContentType: aws.String(mimetype),
	})
	log.Debug(s3instance.ctx, result)
	return err
}

// Get gets from storage
func (s3instance S3Instance) Get(bucket string, key string) (string, []byte, error) {
	log.Debug(s3instance.ctx, bucket)
	log.Debug(s3instance.ctx, key)
	response, err := s3instance.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return "", nil, err
	}
	contenttype := *response.ContentType
	defer response.Body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, response.Body); err != nil {
		return "", nil, err
	}
	log.Debug(s3instance.ctx, contenttype)
	return contenttype, buf.Bytes(), nil
}

// Delete deletes from storage
func (s3instance S3Instance) Delete(bucket string, key string) error {
	result, err := s3instance.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	log.Debug(s3instance.ctx, result)
	return err
}

func (s3instance S3Instance) getContentType(out io.ReadSeeker) (string, error) {
	buffer := make([]byte, 512)
	_, err := out.Read(buffer)
	if err != nil {
		log.Warn(s3instance.ctx, "Date Read Error!")
		log.Warn(s3instance.ctx, err)
	}
	contentType := http.DetectContentType(buffer)
	log.Debug(s3instance.ctx, contentType)
	return contentType, nil
}
