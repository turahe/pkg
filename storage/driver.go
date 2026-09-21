package storage

import (
	"context"
	"errors"
	"io"
)

// ErrNotInitialized is returned by package operations before Setup succeeds
// (or when STORAGE_DRIVER is empty / Setup disabled the driver).
var ErrNotInitialized = errors.New("storage: not initialized (call Setup or set STORAGE_DRIVER)")

// Driver is the provider-neutral object storage contract implemented by GCS and S3-compatible backends.
type Driver interface {
	ReadObject(ctx context.Context, objectName string) ([]byte, error)
	ReadObjectAsReader(ctx context.Context, objectName string) (io.ReadCloser, error)
	WriteObject(ctx context.Context, objectName string, data []byte, contentType string) error
	DeleteObject(ctx context.Context, objectName string) error
	ObjectExists(ctx context.Context, objectName string) (bool, error)
	ListObjects(ctx context.Context, prefix string) ([]string, error)
	PresignUpload(ctx context.Context, objectName string, opts PresignUploadOptions) (PresignedUpload, error)
	Close() error
}
