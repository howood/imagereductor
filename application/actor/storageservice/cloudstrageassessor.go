package storageservice

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/howood/imagereductor/infrastructure/client/cloudstorages"
	log "github.com/howood/imagereductor/infrastructure/logger"
)

const RecordNotFoundMsg = "status code: 404"

type CloudStorageAssessor struct {
	instance cloudstorages.StorageInstance
	bucket   string
	ctx      context.Context
}

// インスタンス作成用のメソッド
func NewCloudStorageAssessor(ctx context.Context) *CloudStorageAssessor {
	var I *CloudStorageAssessor
	log.Debug(ctx, "use:"+os.Getenv("STORAGE_TYPE"))
	switch os.Getenv("STORAGE_TYPE") {
	case "s3":
		I = &CloudStorageAssessor{
			instance: cloudstorages.NewS3(ctx),
			bucket:   cloudstorages.S3_BUCKET_UPLOADFILES,
			ctx:      ctx,
		}
	case "gcs":
		I = &CloudStorageAssessor{
			instance: cloudstorages.NewGCS(ctx),
			bucket:   cloudstorages.GCS_BUCKET_UPLOADFILES,
			ctx:      ctx,
		}
	default:
		panic(errors.New("Invalid STORAGE_TYPE"))
	}
	return I
}

func (csa *CloudStorageAssessor) Get(key string) (string, []byte, error) {
	return csa.instance.Get(csa.bucket, key)
}

func (csa *CloudStorageAssessor) Put(path string, file io.ReadSeeker) error {
	return csa.instance.Put(csa.bucket, path, file)
}

func (csa *CloudStorageAssessor) Delete(key string) error {
	return csa.instance.Delete(csa.bucket, key)
}
