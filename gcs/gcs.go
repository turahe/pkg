package gcs

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/turahe/pkg/config"
	"github.com/turahe/pkg/logger"
)

var (
	client     *storage.Client
	bucketName string
	ctx        = context.Background()
)

// Setup initializes the Google Cloud Storage client
func Setup() error {
	configuration := config.GetConfig()

	if !configuration.GCS.Enabled {
		logger.Infof("GCS is disabled, skipping setup")
		return nil
	}

	bucketName = configuration.GCS.BucketName

	var err error
	var opts []option.ClientOption

	// If credentials file is provided, use it
	if configuration.GCS.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(configuration.GCS.CredentialsFile))
	}
	// Otherwise, use Application Default Credentials (ADC)
	// This will work when running on GCP or when GOOGLE_APPLICATION_CREDENTIALS env var is set

	client, err = storage.NewClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create GCS client: %w", err)
	}

	// Verify bucket access
	if bucketName != "" {
		bucket := client.Bucket(bucketName)
		_, err = bucket.Attrs(ctx)
		if err != nil {
			return fmt.Errorf("failed to access bucket %s: %w", bucketName, err)
		}
		logger.Infof("GCS client initialized successfully with bucket: %s", bucketName)
	} else {
		logger.Infof("GCS client initialized successfully (no bucket specified)")
	}

	return nil
}

// GetClient returns the GCS client instance
func GetClient() *storage.Client {
	if client == nil {
		panic("GCS client is not initialized. Call Setup() first.")
	}
	return client
}

// GetBucket returns a bucket handle
func GetBucket() *storage.BucketHandle {
	if bucketName == "" {
		panic("GCS bucket name is not configured")
	}
	return GetClient().Bucket(bucketName)
}

// GetBucketName returns the configured bucket name
func GetBucketName() string {
	return bucketName
}

// ReadObject reads an object from GCS bucket
func ReadObject(objectName string) ([]byte, error) {
	bucket := GetBucket()
	obj := bucket.Object(objectName)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader for object %s: %w", objectName, err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read object %s: %w", objectName, err)
	}

	return data, nil
}

// ReadObjectAsReader returns a reader for an object from GCS bucket
func ReadObjectAsReader(objectName string) (io.ReadCloser, error) {
	bucket := GetBucket()
	obj := bucket.Object(objectName)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader for object %s: %w", objectName, err)
	}

	return reader, nil
}

// WriteObject writes data to an object in GCS bucket
func WriteObject(objectName string, data []byte, contentType string) error {
	bucket := GetBucket()
	obj := bucket.Object(objectName)

	writer := obj.NewWriter(ctx)
	if contentType != "" {
		writer.ContentType = contentType
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return fmt.Errorf("failed to write object %s: %w", objectName, err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer for object %s: %w", objectName, err)
	}

	return nil
}

// DeleteObject deletes an object from GCS bucket
func DeleteObject(objectName string) error {
	bucket := GetBucket()
	obj := bucket.Object(objectName)

	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete object %s: %w", objectName, err)
	}

	return nil
}

// ObjectExists checks if an object exists in the bucket
func ObjectExists(objectName string) (bool, error) {
	bucket := GetBucket()
	obj := bucket.Object(objectName)

	_, err := obj.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check object existence %s: %w", objectName, err)
	}

	return true, nil
}

// ListObjects lists objects in the bucket with the given prefix
func ListObjects(prefix string) ([]string, error) {
	bucket := GetBucket()
	query := &storage.Query{
		Prefix: prefix,
	}

	var objectNames []string
	it := bucket.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == storage.ErrObjectNotExist || err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list objects with prefix %s: %w", prefix, err)
		}
		objectNames = append(objectNames, attrs.Name)
	}

	return objectNames, nil
}

// Close closes the GCS client
func Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}
