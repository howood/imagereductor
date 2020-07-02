package storageservice

import (
	"os"

	"github.com/howood/imagereductor/infrastructure/client/cloudstorages"
	log "github.com/howood/imagereductor/infrastructure/logger"
)

type CloudStorageAssessor struct {
	instance cloudstorages.StorageInstance
	bucket   string
}

// インスタンス作成用のメソッド
func NewCloudStorageAssessor() *CloudStorageAssessor {
	var I *CloudStorageAssessor
	log.Debug("", "use:"+os.Getenv("STORAGE_TYPE"))
	if os.Getenv("STORAGE_TYPE") == "s3" {
		I = &CloudStorageAssessor{
			instance: cloudstorages.NewS3(),
			bucket:   cloudstorages.S3_BUCKET_UPLOADFILES,
		}
	} else if os.Getenv("STORAGE_TYPE") == "gcs" {
		I = &CloudStorageAssessor{
			instance: cloudstorages.NewGCS(),
			bucket:   cloudstorages.GCS_BUCKET_UPLOADFILES,
		}
	}
	return I
}

func (csa *CloudStorageAssessor) Get(key string) (string, []byte, error) {
	return csa.instance.Get(csa.bucket, key)
}

func (csa *CloudStorageAssessor) Put(path string, file *os.File) error {
	return csa.instance.Put(csa.bucket, path, file)
}

func (csa *CloudStorageAssessor) Delete(key string) error {
	return csa.instance.Delete(csa.bucket, key)
}
