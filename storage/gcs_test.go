package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/turahe/pkg/config"
)

func TestGCSPresignUpload_ReturnsPUTURL(t *testing.T) {
	credFile := os.Getenv("GCS_CREDENTIALS_FILE")
	bucket := os.Getenv("STORAGE_BUCKET")
	if credFile == "" || bucket == "" {
		t.Skip("set GCS_CREDENTIALS_FILE and STORAGE_BUCKET to run GCS presign test")
	}
	t.Cleanup(func() {
		config.Config = nil
		_ = Close()
	})
	config.Config = &config.Configuration{Storage: config.StorageConfiguration{
		Driver: "gcs", Bucket: bucket, CredentialsFile: credFile,
	}}
	if err := Setup(); err != nil {
		t.Fatal(err)
	}
	out, err := PresignUpload(context.Background(), "uploads/hello.txt", PresignUploadOptions{
		Expiry:      time.Minute,
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Method != "PUT" || out.URL == "" {
		t.Fatalf("%+v", out)
	}
	if out.Headers["Content-Type"] != "text/plain" {
		t.Fatalf("Headers = %#v", out.Headers)
	}
}
