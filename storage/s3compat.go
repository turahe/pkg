package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/turahe/pkg/config"
)

type s3Driver struct {
	client *s3.Client
	bucket string
}

func resolveForcePathStyle(driver string, override *bool) bool {
	if override != nil {
		return *override
	}
	return driver != "r2"
}

func resolveRegion(driver, region string) string {
	if region != "" {
		return region
	}
	if driver == "r2" {
		return "auto"
	}
	return "us-east-1"
}

func newS3CompatDriver(ctx context.Context, cfg config.StorageConfiguration) (Driver, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("storage/%s: STORAGE_ENDPOINT required", cfg.Driver)
	}
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("storage/%s: STORAGE_ACCESS_KEY and STORAGE_SECRET_KEY required", cfg.Driver)
	}

	region := resolveRegion(cfg.Driver, cfg.Region)
	pathStyle := resolveForcePathStyle(cfg.Driver, cfg.ForcePathStyle)

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		// Non-AWS S3 (RustFS/MinIO/R2) often rejects AWS SDK v2 default CRC32 checksum headers.
		awsconfig.WithRequestChecksumCalculation(aws.RequestChecksumCalculationWhenRequired),
		awsconfig.WithResponseChecksumValidation(aws.ResponseChecksumValidationWhenRequired),
	)
	if err != nil {
		return nil, fmt.Errorf("storage/%s: %w", cfg.Driver, err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = pathStyle
	})

	return &s3Driver{client: client, bucket: cfg.Bucket}, nil
}

func (d *s3Driver) ReadObject(ctx context.Context, objectName string) ([]byte, error) {
	reader, err := d.ReadObjectAsReader(ctx, objectName)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("storage/s3: read object %s: %w", objectName, err)
	}

	return data, nil
}

func (d *s3Driver) ReadObjectAsReader(ctx context.Context, objectName string) (io.ReadCloser, error) {
	out, err := d.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return nil, fmt.Errorf("storage/s3: create reader for object %s: %w", objectName, err)
	}

	return out.Body, nil
}

func (d *s3Driver) WriteObject(ctx context.Context, objectName string, data []byte, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(objectName),
		Body:   bytes.NewReader(data),
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	if _, err := d.client.PutObject(ctx, input); err != nil {
		return fmt.Errorf("storage/s3: write object %s: %w", objectName, err)
	}

	return nil
}

func (d *s3Driver) DeleteObject(ctx context.Context, objectName string) error {
	if _, err := d.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(objectName),
	}); err != nil {
		return fmt.Errorf("storage/s3: delete object %s: %w", objectName, err)
	}

	return nil
}

func (d *s3Driver) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	if _, err := d.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(objectName),
	}); err != nil {
		if isS3NotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("storage/s3: check object existence %s: %w", objectName, err)
	}

	return true, nil
}

func (d *s3Driver) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	paginator := s3.NewListObjectsV2Paginator(d.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(d.bucket),
		Prefix: aws.String(prefix),
	})

	var objectNames []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("storage/s3: list objects with prefix %s: %w", prefix, err)
		}
		for _, object := range page.Contents {
			if object.Key != nil {
				objectNames = append(objectNames, *object.Key)
			}
		}
	}

	return objectNames, nil
}

func (d *s3Driver) Close() error {
	return nil
}

func isS3NotFound(err error) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}

	switch apiErr.ErrorCode() {
	case "NotFound", "NoSuchKey", "404":
		return true
	default:
		return false
	}
}
