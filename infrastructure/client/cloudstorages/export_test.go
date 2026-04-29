package cloudstorages

import "cloud.google.com/go/storage"

// NewGCSInstanceForTest creates a GCSInstance with a custom client for testing purposes.
func NewGCSInstanceForTest(client *storage.Client, cfg GCSConfig) *GCSInstance {
	return &GCSInstance{client: client, cfg: cfg}
}
