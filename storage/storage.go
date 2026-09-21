package storage

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/turahe/pkg/config"
	"github.com/turahe/pkg/logger"
)

var (
	mu         sync.RWMutex
	driver     Driver
	bucketName string
)

// Setup initializes the storage driver from config using a background context.
// No-op when STORAGE_DRIVER is empty.
func Setup() error { return SetupContext(context.Background()) }

// SetupContext initializes the storage driver from config.
// When STORAGE_DRIVER is empty, any previously configured driver is closed and
// the package returns to the uninitialized state.
func SetupContext(ctx context.Context) error {
	cfg := config.GetConfig().Storage
	if cfg.Driver == "" {
		logger.Infof("storage: disabled (STORAGE_DRIVER empty)")
		mu.Lock()
		if driver != nil {
			_ = driver.Close()
		}
		driver, bucketName = nil, ""
		mu.Unlock()
		return nil
	}
	if cfg.Bucket == "" {
		return fmt.Errorf("storage: STORAGE_BUCKET required when STORAGE_DRIVER=%s", cfg.Driver)
	}
	d, err := newDriver(ctx, cfg)
	if err != nil {
		return err
	}
	mu.Lock()
	if driver != nil {
		_ = driver.Close()
	}
	driver = d
	bucketName = cfg.Bucket
	mu.Unlock()
	logger.Infof("storage: enabled driver=%s bucket=%s", cfg.Driver, cfg.Bucket)
	return nil
}

func newDriver(ctx context.Context, cfg config.StorageConfiguration) (Driver, error) {
	switch cfg.Driver {
	case "gcs":
		return newGCSDriver(ctx, cfg)
	case "s3", "r2":
		return newS3CompatDriver(ctx, cfg)
	default:
		return nil, fmt.Errorf("storage: unknown or unimplemented driver %q", cfg.Driver)
	}
}

// Close shuts down the active driver and clears package state.
func Close() error {
	mu.Lock()
	defer mu.Unlock()
	if driver == nil {
		return nil
	}
	err := driver.Close()
	driver = nil
	bucketName = ""
	return err
}

// GetBucketName returns the configured bucket name, or empty if not initialized.
func GetBucketName() string {
	mu.RLock()
	defer mu.RUnlock()
	return bucketName
}

func current() (Driver, error) {
	mu.RLock()
	defer mu.RUnlock()
	if driver == nil {
		return nil, ErrNotInitialized
	}
	return driver, nil
}

// ReadObject downloads an object and returns its full contents.
func ReadObject(ctx context.Context, objectName string) ([]byte, error) {
	d, err := current()
	if err != nil {
		return nil, err
	}
	return d.ReadObject(ctx, objectName)
}

// ReadObjectAsReader returns a streaming reader for an object. The caller must Close it.
func ReadObjectAsReader(ctx context.Context, objectName string) (io.ReadCloser, error) {
	d, err := current()
	if err != nil {
		return nil, err
	}
	return d.ReadObjectAsReader(ctx, objectName)
}

// WriteObject uploads object bytes with an optional content type.
func WriteObject(ctx context.Context, objectName string, data []byte, contentType string) error {
	d, err := current()
	if err != nil {
		return err
	}
	return d.WriteObject(ctx, objectName, data, contentType)
}

// DeleteObject removes an object.
func DeleteObject(ctx context.Context, objectName string) error {
	d, err := current()
	if err != nil {
		return err
	}
	return d.DeleteObject(ctx, objectName)
}

// ObjectExists reports whether an object exists.
func ObjectExists(ctx context.Context, objectName string) (bool, error) {
	d, err := current()
	if err != nil {
		return false, err
	}
	return d.ObjectExists(ctx, objectName)
}

// ListObjects lists object names under the given prefix.
func ListObjects(ctx context.Context, prefix string) ([]string, error) {
	d, err := current()
	if err != nil {
		return nil, err
	}
	return d.ListObjects(ctx, prefix)
}
