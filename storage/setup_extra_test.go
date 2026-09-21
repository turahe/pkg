package storage

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/turahe/pkg/config"
)

func TestSetup_UnknownDriver(t *testing.T) {
	t.Cleanup(func() {
		config.Config = nil
		_ = Close()
	})
	config.Config = &config.Configuration{
		Storage: config.StorageConfiguration{Driver: "azure", Bucket: "b"},
	}
	err := Setup()
	if err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("got %v", err)
	}
}

func TestSetup_R2RequiresEndpoint(t *testing.T) {
	t.Cleanup(func() {
		config.Config = nil
		_ = Close()
	})
	config.Config = &config.Configuration{
		Storage: config.StorageConfiguration{
			Driver: "r2", Bucket: "b", AccessKey: "a", SecretKey: "b",
		},
	}
	if err := Setup(); err == nil {
		t.Fatal("expected endpoint required")
	}
}

func TestSetup_S3MissingKeys(t *testing.T) {
	t.Cleanup(func() {
		config.Config = nil
		_ = Close()
	})
	config.Config = &config.Configuration{
		Storage: config.StorageConfiguration{
			Driver: "s3", Bucket: "b", Endpoint: "http://127.0.0.1:9000",
		},
	}
	if err := Setup(); err == nil {
		t.Fatal("expected access key required")
	}
}

func TestS3Driver_OpsAgainstRefusedEndpoint(t *testing.T) {
	t.Cleanup(func() {
		config.Config = nil
		_ = Close()
	})
	config.Config = &config.Configuration{Storage: config.StorageConfiguration{
		Driver: "s3", Bucket: "pkg", Endpoint: "http://127.0.0.1:1",
		AccessKey: "a", SecretKey: "b", Region: "us-east-1", ForcePathStyle: boolPtr(true),
	}}
	if err := Setup(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := WriteObject(ctx, "k", []byte("v"), "text/plain"); err == nil {
		t.Fatal("WriteObject expected error")
	}
	if err := WriteObject(ctx, "k", []byte("v"), ""); err == nil {
		t.Fatal("WriteObject empty type expected error")
	}
	if _, err := ReadObject(ctx, "k"); err == nil {
		t.Fatal("ReadObject expected error")
	}
	if _, err := ReadObjectAsReader(ctx, "k"); err == nil {
		t.Fatal("ReadObjectAsReader expected error")
	}
	if err := DeleteObject(ctx, "k"); err == nil {
		t.Fatal("DeleteObject expected error")
	}
	if _, err := ObjectExists(ctx, "k"); err == nil {
		t.Fatal("ObjectExists expected error")
	}
	if _, err := ListObjects(ctx, "p/"); err == nil {
		t.Fatal("ListObjects expected error")
	}
}

func TestS3PresignUpload_WithoutContentType(t *testing.T) {
	t.Cleanup(func() {
		config.Config = nil
		_ = Close()
	})
	config.Config = &config.Configuration{Storage: config.StorageConfiguration{
		Driver: "s3", Bucket: "pkg", Endpoint: "http://127.0.0.1:9000",
		AccessKey: "rustfsadmin", SecretKey: "rustfsadmin",
		Region: "us-east-1", ForcePathStyle: boolPtr(true),
	}}
	if err := Setup(); err != nil {
		t.Fatal(err)
	}
	out, err := PresignUpload(context.Background(), "uploads/x.bin", PresignUploadOptions{Expiry: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if out.Method != "PUT" || out.URL == "" {
		t.Fatalf("%+v", out)
	}
	if len(out.Headers) != 0 {
		t.Fatalf("Headers = %#v", out.Headers)
	}
}

func TestClose_Idempotent(t *testing.T) {
	if err := Close(); err != nil {
		t.Fatal(err)
	}
	if err := Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGetBucketName_EmptyWhenUninitialized(t *testing.T) {
	_ = Close()
	if GetBucketName() != "" {
		t.Fatalf("got %q", GetBucketName())
	}
}

func TestSetup_ReplacesPreviousDriver(t *testing.T) {
	t.Cleanup(func() {
		config.Config = nil
		_ = Close()
	})
	config.Config = &config.Configuration{Storage: config.StorageConfiguration{
		Driver: "s3", Bucket: "pkg", Endpoint: "http://127.0.0.1:9000",
		AccessKey: "a", SecretKey: "b", ForcePathStyle: boolPtr(true),
	}}
	if err := Setup(); err != nil {
		t.Fatal(err)
	}
	if GetBucketName() != "pkg" {
		t.Fatalf("bucket = %q", GetBucketName())
	}
	// Second Setup with same shape should Close the previous driver and succeed.
	config.Config = &config.Configuration{Storage: config.StorageConfiguration{
		Driver: "s3", Bucket: "pkg2", Endpoint: "http://127.0.0.1:9000",
		AccessKey: "a", SecretKey: "b", ForcePathStyle: boolPtr(true),
	}}
	if err := Setup(); err != nil {
		t.Fatal(err)
	}
	if GetBucketName() != "pkg2" {
		t.Fatalf("bucket = %q", GetBucketName())
	}
}

func TestNewDriver_GCS(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/sa.json"
	if err := os.WriteFile(path, testServiceAccountJSON(t), 0o600); err != nil {
		t.Fatal(err)
	}
	d, err := newDriver(context.Background(), config.StorageConfiguration{
		Driver:          "gcs",
		Bucket:          "",
		CredentialsFile: path,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestFacade_OpsWhenUninitialized(t *testing.T) {
	_ = Close()
	ctx := context.Background()
	if _, err := ReadObjectAsReader(ctx, "x"); err == nil {
		t.Fatal("expected ErrNotInitialized")
	}
	if err := WriteObject(ctx, "x", nil, ""); err == nil {
		t.Fatal("expected ErrNotInitialized")
	}
	if err := DeleteObject(ctx, "x"); err == nil {
		t.Fatal("expected ErrNotInitialized")
	}
	if _, err := ObjectExists(ctx, "x"); err == nil {
		t.Fatal("expected ErrNotInitialized")
	}
	if _, err := ListObjects(ctx, ""); err == nil {
		t.Fatal("expected ErrNotInitialized")
	}
}
