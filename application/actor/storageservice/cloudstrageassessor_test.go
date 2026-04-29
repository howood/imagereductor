package storageservice_test

import (
	"errors"
	"testing"

	"github.com/howood/imagereductor/application/actor/storageservice"
)

func Test_NewCloudStorageAssessorWithConfig_EmptyType(t *testing.T) {
	t.Setenv("STORAGE_TYPE", "")
	_, err := storageservice.NewCloudStorageAssessorWithConfig(t.Context())
	if !errors.Is(err, storageservice.ErrStorageTypeEmpty) {
		t.Fatalf("expected ErrStorageTypeEmpty, got %v", err)
	}
}

func Test_NewCloudStorageAssessorWithConfig_InvalidType(t *testing.T) {
	t.Setenv("STORAGE_TYPE", "unknown")
	_, err := storageservice.NewCloudStorageAssessorWithConfig(t.Context())
	if !errors.Is(err, storageservice.ErrInvalidStorageType) {
		t.Fatalf("expected ErrInvalidStorageType, got %v", err)
	}
}

func Test_NewCloudStorageAssessorWithConfig_S3MissingBucket(t *testing.T) {
	t.Setenv("STORAGE_TYPE", "s3")
	t.Setenv("AWS_S3_BUKET", "")
	_, err := storageservice.NewCloudStorageAssessorWithConfig(t.Context())
	if err == nil {
		t.Fatal("expected error for missing s3 bucket, got nil")
	}
}

func Test_NewCloudStorageAssessorWithConfig_GCSMissingBucket(t *testing.T) {
	t.Setenv("STORAGE_TYPE", "gcs")
	t.Setenv("GCS_BUKET", "")
	_, err := storageservice.NewCloudStorageAssessorWithConfig(t.Context())
	if err == nil {
		t.Fatal("expected error for missing gcs bucket, got nil")
	}
}
