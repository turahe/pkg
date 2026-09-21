package storage

import (
	"context"
	"errors"
	"io"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/turahe/pkg/config"
)

func TestResolveForcePathStyle(t *testing.T) {
	tr, fa := true, false
	tests := []struct {
		name   string
		driver string
		in     *bool
		want   bool
	}{
		{name: "s3 defaults to path style", driver: "s3", in: nil, want: true},
		{name: "r2 defaults to virtual host style", driver: "r2", in: nil, want: false},
		{name: "s3 honors false override", driver: "s3", in: &fa, want: false},
		{name: "r2 honors true override", driver: "r2", in: &tr, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveForcePathStyle(tt.driver, tt.in); got != tt.want {
				t.Fatalf("resolveForcePathStyle(%q, %v) = %v, want %v", tt.driver, tt.in, got, tt.want)
			}
		})
	}
}

func TestResolveRegion(t *testing.T) {
	tests := []struct {
		name   string
		driver string
		region string
		want   string
	}{
		{name: "explicit region", driver: "s3", region: "ap-southeast-1", want: "ap-southeast-1"},
		{name: "s3 default", driver: "s3", region: "", want: "us-east-1"},
		{name: "r2 default", driver: "r2", region: "", want: "auto"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveRegion(tt.driver, tt.region); got != tt.want {
				t.Fatalf("resolveRegion(%q, %q) = %q, want %q", tt.driver, tt.region, got, tt.want)
			}
		})
	}
}

func TestSetup_S3RequiresEndpoint(t *testing.T) {
	t.Cleanup(func() {
		config.Config = nil
		_ = Close()
	})

	config.Config = &config.Configuration{
		Storage: config.StorageConfiguration{
			Driver:    "s3",
			Bucket:    "pkg",
			AccessKey: "a",
			SecretKey: "b",
		},
	}

	if err := Setup(); err == nil {
		t.Fatal("expected endpoint required")
	}
}

func TestS3Compat_RustFSIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping RustFS integration in short mode")
	}
	t.Cleanup(func() {
		config.Config = nil
		_ = Close()
	})

	endpoint := "http://127.0.0.1:9000"
	if configEndpoint := config.GetConfig().Storage.Endpoint; configEndpoint != "" {
		endpoint = configEndpoint
	}
	if !canDialEndpoint(endpoint) {
		t.Skipf("RustFS unreachable at %s", endpoint)
	}

	config.Config = &config.Configuration{Storage: config.StorageConfiguration{
		Driver:         "s3",
		Bucket:         "pkg",
		Endpoint:       endpoint,
		AccessKey:      "rustfsadmin",
		SecretKey:      "rustfsadmin",
		ForcePathStyle: boolPtr(true),
	}}

	if err := Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if err := ensureS3CompatBucket(context.Background(), "pkg"); err != nil {
		if isS3AccessDenied(err) {
			t.Skipf("RustFS at %s rejected credentials (not this project's compose service?): %v", endpoint, err)
		}
		t.Fatalf("CreateBucket: %v", err)
	}

	ctx := context.Background()
	key := "storage-test/hello.txt"
	if err := WriteObject(ctx, key, []byte("hi"), "text/plain"); err != nil {
		t.Fatal(err)
	}
	data, err := ReadObject(ctx, key)
	if err != nil || string(data) != "hi" {
		t.Fatalf("read: %v %q", err, data)
	}
	r, err := ReadObjectAsReader(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	streamed, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil || string(streamed) != "hi" {
		t.Fatalf("ReadObjectAsReader: %v %q", err, streamed)
	}
	exists, existsErr := ObjectExists(ctx, key)
	if existsErr != nil || !exists {
		t.Fatalf("exists: %v %v", exists, existsErr)
	}
	exists, existsErr = ObjectExists(ctx, key+"-missing")
	if existsErr != nil || exists {
		t.Fatalf("missing exists: %v %v", exists, existsErr)
	}
	names, err := ListObjects(ctx, "storage-test/")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, n := range names {
		if n == key {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListObjects missing %s: %v", key, names)
	}
	presigned, err := PresignUpload(ctx, key, PresignUploadOptions{ContentType: "text/plain", Expiry: time.Minute})
	if err != nil || presigned.Method != "PUT" || presigned.URL == "" {
		t.Fatalf("PresignUpload: %+v %v", presigned, err)
	}
	if err := DeleteObject(ctx, key); err != nil {
		t.Fatal(err)
	}
}

func TestS3PresignUpload_ReturnsPUTURL(t *testing.T) {
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
	out, err := PresignUpload(context.Background(), "uploads/hello.txt", PresignUploadOptions{
		Expiry:      time.Minute,
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Method != "PUT" {
		t.Fatalf("Method = %q", out.Method)
	}
	if out.URL == "" {
		t.Fatal("empty URL")
	}
	if out.Headers["Content-Type"] != "text/plain" {
		t.Fatalf("Headers = %#v", out.Headers)
	}
	if !out.ExpiresAt.After(time.Now()) {
		t.Fatalf("ExpiresAt = %v", out.ExpiresAt)
	}
}

func boolPtr(v bool) *bool {
	return &v
}

func canDialEndpoint(endpoint string) bool {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" {
		return false
	}

	conn, err := net.DialTimeout("tcp", u.Host, time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func ensureS3CompatBucket(ctx context.Context, bucket string) error {
	d, err := current()
	if err != nil {
		return err
	}
	s3d, ok := d.(*s3Driver)
	if !ok {
		return nil
	}

	_, err = s3d.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	if err != nil && !isS3BucketAlreadyExists(err) {
		return err
	}
	return nil
}

func isS3BucketAlreadyExists(err error) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}

	switch apiErr.ErrorCode() {
	case "BucketAlreadyOwnedByYou", "BucketAlreadyExists":
		return true
	default:
		return false
	}
}

func TestIsS3NotFound(t *testing.T) {
	if isS3NotFound(errors.New("nope")) {
		t.Fatal("expected false for plain error")
	}
	for _, code := range []string{"NotFound", "NoSuchKey", "404"} {
		err := &fakeAPIError{code: code}
		if !isS3NotFound(err) {
			t.Fatalf("code %s: expected true", code)
		}
	}
	if isS3NotFound(&fakeAPIError{code: "AccessDenied"}) {
		t.Fatal("AccessDenied should not be not-found")
	}
}

type fakeAPIError struct {
	code string
}

func (e *fakeAPIError) Error() string                 { return e.code }
func (e *fakeAPIError) ErrorCode() string             { return e.code }
func (e *fakeAPIError) ErrorMessage() string          { return e.code }
func (e *fakeAPIError) ErrorFault() smithy.ErrorFault { return smithy.FaultUnknown }

func isS3AccessDenied(err error) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}

	switch apiErr.ErrorCode() {
	case "InvalidAccessKeyId", "AccessDenied", "SignatureDoesNotMatch", "InvalidToken":
		return true
	default:
		return false
	}
}
