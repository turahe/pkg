package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	gcsstorage "cloud.google.com/go/storage"
	"github.com/turahe/pkg/config"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type gcsDriver struct {
	client *gcsstorage.Client
	bucket string
}

func newGCSDriver(ctx context.Context, cfg config.StorageConfiguration) (Driver, error) {
	var opts []option.ClientOption
	if cfg.CredentialsFile != "" {
		opts = append(opts, option.WithAuthCredentialsFile(option.ServiceAccount, cfg.CredentialsFile))
	}

	client, err := gcsstorage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("storage/gcs: %w", err)
	}

	if cfg.Bucket != "" {
		if _, err := client.Bucket(cfg.Bucket).Attrs(ctx); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("storage/gcs: bucket %s: %w", cfg.Bucket, err)
		}
	}

	return &gcsDriver{client: client, bucket: cfg.Bucket}, nil
}

func (d *gcsDriver) bucketHandle() *gcsstorage.BucketHandle {
	return d.client.Bucket(d.bucket)
}

func (d *gcsDriver) ReadObject(ctx context.Context, objectName string) ([]byte, error) {
	reader, err := d.ReadObjectAsReader(ctx, objectName)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("storage/gcs: read object %s: %w", objectName, err)
	}

	return data, nil
}

func (d *gcsDriver) ReadObjectAsReader(ctx context.Context, objectName string) (io.ReadCloser, error) {
	reader, err := d.bucketHandle().Object(objectName).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage/gcs: create reader for object %s: %w", objectName, err)
	}

	return reader, nil
}

func (d *gcsDriver) WriteObject(ctx context.Context, objectName string, data []byte, contentType string) error {
	writer := d.bucketHandle().Object(objectName).NewWriter(ctx)
	if contentType != "" {
		writer.ContentType = contentType
	}

	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		return fmt.Errorf("storage/gcs: write object %s: %w", objectName, err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("storage/gcs: close writer for object %s: %w", objectName, err)
	}

	return nil
}

func (d *gcsDriver) DeleteObject(ctx context.Context, objectName string) error {
	if err := d.bucketHandle().Object(objectName).Delete(ctx); err != nil {
		return fmt.Errorf("storage/gcs: delete object %s: %w", objectName, err)
	}

	return nil
}

func (d *gcsDriver) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	_, err := d.bucketHandle().Object(objectName).Attrs(ctx)
	if errors.Is(err, gcsstorage.ErrObjectNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("storage/gcs: check object existence %s: %w", objectName, err)
	}

	return true, nil
}

func (d *gcsDriver) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	query := &gcsstorage.Query{Prefix: prefix}
	it := d.bucketHandle().Objects(ctx, query)

	var objectNames []string
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("storage/gcs: list objects with prefix %s: %w", prefix, err)
		}
		objectNames = append(objectNames, attrs.Name)
	}

	return objectNames, nil
}

func (d *gcsDriver) Close() error {
	return d.client.Close()
}
