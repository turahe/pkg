package storage

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/turahe/pkg/config"
)

func TestResolvePresignUploadOptions(t *testing.T) {
	t.Parallel()
	got, err := resolvePresignUploadOptions(PresignUploadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Expiry != DefaultPresignUploadExpiry {
		t.Fatalf("Expiry = %v, want %v", got.Expiry, DefaultPresignUploadExpiry)
	}

	got, err = resolvePresignUploadOptions(PresignUploadOptions{Expiry: time.Hour, ContentType: "image/png"})
	if err != nil || got.Expiry != time.Hour || got.ContentType != "image/png" {
		t.Fatalf("got %+v err %v", got, err)
	}

	if _, err := resolvePresignUploadOptions(PresignUploadOptions{Expiry: -time.Second}); err == nil {
		t.Fatal("expected negative expiry error")
	}
	if _, err := resolvePresignUploadOptions(PresignUploadOptions{Expiry: MaxPresignUploadExpiry + time.Second}); err == nil {
		t.Fatal("expected max expiry error")
	}
}

func TestPresignUpload_NotInitialized(t *testing.T) {
	t.Cleanup(func() {
		config.Config = nil
		_ = Close()
	})
	config.Config = &config.Configuration{}
	_ = Setup()
	_, err := PresignUpload(context.Background(), "a/b", PresignUploadOptions{})
	if !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("got %v, want ErrNotInitialized", err)
	}
}

func TestPresignUpload_EmptyObjectName(t *testing.T) {
	t.Cleanup(func() {
		config.Config = nil
		_ = Close()
	})
	config.Config = &config.Configuration{
		Storage: config.StorageConfiguration{
			Driver: "s3", Bucket: "pkg", Endpoint: "http://127.0.0.1:9000",
			AccessKey: "a", SecretKey: "b", ForcePathStyle: boolPtr(true),
		},
	}
	if err := Setup(); err != nil {
		t.Fatal(err)
	}
	_, err := PresignUpload(context.Background(), "  ", PresignUploadOptions{})
	if err == nil || !strings.Contains(err.Error(), "object name required") {
		t.Fatalf("got %v", err)
	}
}
