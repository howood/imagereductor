package cloudstorages_test

import (
	"errors"
	"testing"
	"time"

	"github.com/howood/imagereductor/infrastructure/client/cloudstorages"
)

func Test_LoadS3ConfigFromEnv_Defaults(t *testing.T) {
	t.Setenv("AWS_S3_REGION", "")
	t.Setenv("AWS_S3_ENDPOINT", "")
	t.Setenv("AWS_S3_LOCALUSE", "")
	t.Setenv("AWS_S3_ACCESSKEY", "")
	t.Setenv("AWS_S3_SECRETKEY", "")
	t.Setenv("AWS_S3_BUKET", "")
	t.Setenv("AWS_S3_TIMEOUT", "")

	cfg := cloudstorages.LoadS3ConfigFromEnv()
	if cfg.Region != "" || cfg.Bucket != "" {
		t.Fatalf("expected empty defaults, got %+v", cfg)
	}
	if cfg.Timeout != 30*time.Second {
		t.Fatalf("expected 30s default timeout, got %v", cfg.Timeout)
	}
}

func Test_LoadS3ConfigFromEnv_Values(t *testing.T) {
	t.Setenv("AWS_S3_REGION", "us-east-1")
	t.Setenv("AWS_S3_ENDPOINT", "http://localhost:9000")
	t.Setenv("AWS_S3_LOCALUSE", "1")
	t.Setenv("AWS_S3_ACCESSKEY", "ak")
	t.Setenv("AWS_S3_SECRETKEY", "sk")
	t.Setenv("AWS_S3_BUKET", "mybucket")
	t.Setenv("AWS_S3_TIMEOUT", "5s")

	cfg := cloudstorages.LoadS3ConfigFromEnv()
	if cfg.Region != "us-east-1" || cfg.Endpoint != "http://localhost:9000" ||
		!cfg.UseLocal || cfg.AccessKey != "ak" || cfg.SecretKey != "sk" ||
		cfg.Bucket != "mybucket" || cfg.Timeout != 5*time.Second {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func Test_LoadS3ConfigFromEnv_InvalidTimeout(t *testing.T) {
	t.Setenv("AWS_S3_TIMEOUT", "not-a-duration")
	t.Setenv("AWS_S3_BUKET", "x")
	cfg := cloudstorages.LoadS3ConfigFromEnv()
	if cfg.Timeout != 30*time.Second {
		t.Fatalf("expected fallback timeout 30s, got %v", cfg.Timeout)
	}
}

func Test_NewS3WithConfig_EmptyBucket(t *testing.T) {
	t.Parallel()

	_, err := cloudstorages.NewS3WithConfig(t.Context(), cloudstorages.S3Config{})
	if !errors.Is(err, cloudstorages.ErrS3BucketEmpty) {
		t.Fatalf("expected ErrS3BucketEmpty, got %v", err)
	}
}

func Test_LoadGCSConfigFromEnv_Defaults(t *testing.T) {
	t.Setenv("GCS_PROJECTID", "")
	t.Setenv("GCS_BUKET", "")
	t.Setenv("GCS_TIMEOUT", "")
	cfg := cloudstorages.LoadGCSConfigFromEnv()
	if cfg.ProjectID != "" || cfg.Bucket != "" {
		t.Fatalf("expected empty defaults, got %+v", cfg)
	}
	if cfg.Timeout != 30*time.Second {
		t.Fatalf("expected 30s default, got %v", cfg.Timeout)
	}
}

func Test_LoadGCSConfigFromEnv_Values(t *testing.T) {
	t.Setenv("GCS_PROJECTID", "my-project")
	t.Setenv("GCS_BUKET", "my-bucket")
	t.Setenv("GCS_TIMEOUT", "10s")
	cfg := cloudstorages.LoadGCSConfigFromEnv()
	if cfg.ProjectID != "my-project" || cfg.Bucket != "my-bucket" || cfg.Timeout != 10*time.Second {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func Test_LoadGCSConfigFromEnv_InvalidTimeout(t *testing.T) {
	t.Setenv("GCS_TIMEOUT", "bogus")
	cfg := cloudstorages.LoadGCSConfigFromEnv()
	if cfg.Timeout != 30*time.Second {
		t.Fatalf("expected fallback 30s, got %v", cfg.Timeout)
	}
}

func Test_NewGCSWithConfig_EmptyBucket(t *testing.T) {
	t.Parallel()

	_, err := cloudstorages.NewGCSWithConfig(t.Context(), cloudstorages.GCSConfig{})
	if !errors.Is(err, cloudstorages.ErrGCSBucketEmpty) {
		t.Fatalf("expected ErrGCSBucketEmpty, got %v", err)
	}
}

func Test_NewGCSWithConfig_EmptyProjectID(t *testing.T) {
	t.Parallel()

	_, err := cloudstorages.NewGCSWithConfig(t.Context(), cloudstorages.GCSConfig{Bucket: "b"})
	if !errors.Is(err, cloudstorages.ErrGCSProjectIDEmpty) {
		t.Fatalf("expected ErrGCSProjectIDEmpty, got %v", err)
	}
}
