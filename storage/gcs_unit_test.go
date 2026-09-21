package storage

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	gcsstorage "cloud.google.com/go/storage"
	"github.com/turahe/pkg/config"
	"google.golang.org/api/option"
)

func testServiceAccountJSON(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der := x509.MarshalPKCS1PrivateKey(key)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})
	payload := map[string]string{
		"type":         "service_account",
		"project_id":   "test-project",
		"private_key":  string(pemBytes),
		"client_email": "tester@test-project.iam.gserviceaccount.com",
		"token_uri":    "https://oauth2.googleapis.com/token",
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestGCSPresignUpload_WithServiceAccountJSON(t *testing.T) {
	ctx := context.Background()
	client, err := gcsstorage.NewClient(ctx, option.WithAuthCredentialsJSON(option.ServiceAccount, testServiceAccountJSON(t)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	d := &gcsDriver{client: client, bucket: "test-bucket"}
	out, err := d.PresignUpload(ctx, "uploads/hello.txt", PresignUploadOptions{
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
	if !strings.Contains(out.URL, "test-bucket") {
		t.Fatalf("URL missing bucket: %s", out.URL)
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGCSPresignUpload_WithoutContentType(t *testing.T) {
	ctx := context.Background()
	client, err := gcsstorage.NewClient(ctx, option.WithAuthCredentialsJSON(option.ServiceAccount, testServiceAccountJSON(t)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	d := &gcsDriver{client: client, bucket: "test-bucket"}
	out, err := d.PresignUpload(ctx, "obj", PresignUploadOptions{Expiry: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Headers) != 0 {
		t.Fatalf("Headers = %#v, want empty", out.Headers)
	}
}

func TestNewGCSDriver_MissingCredentialsFile(t *testing.T) {
	_, err := newGCSDriver(context.Background(), config.StorageConfiguration{
		Driver:          "gcs",
		Bucket:          "b",
		CredentialsFile: "/nonexistent/gcs-creds.json",
	})
	if err == nil {
		t.Fatal("expected error for missing credentials file")
	}
}

func TestNewGCSDriver_EmptyBucketSkipsAttrs(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/sa.json"
	if err := os.WriteFile(path, testServiceAccountJSON(t), 0o600); err != nil {
		t.Fatal(err)
	}
	d, err := newGCSDriver(context.Background(), config.StorageConfiguration{
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

func TestNewGCSDriver_BucketAttrsFails(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/sa.json"
	if err := os.WriteFile(path, testServiceAccountJSON(t), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := newGCSDriver(ctx, config.StorageConfiguration{
		Driver:          "gcs",
		Bucket:          "definitely-not-a-real-bucket-xyz",
		CredentialsFile: path,
	})
	if err == nil {
		t.Fatal("expected bucket attrs error")
	}
}

type errTransport struct{}

func (errTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, context.Canceled
}

func TestGCSDriver_ImmediateRPCErrors(t *testing.T) {
	ctx := context.Background()
	client, err := gcsstorage.NewClient(ctx,
		option.WithoutAuthentication(),
		option.WithHTTPClient(&http.Client{Transport: errTransport{}}),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	d := &gcsDriver{client: client, bucket: "b"}
	// Only call methods that issue an RPC immediately (avoid Writer.Close hang/retry).
	if _, err := d.ReadObjectAsReader(ctx, "x"); err == nil {
		t.Fatal("ReadObjectAsReader expected error")
	}
	if _, err := d.ReadObject(ctx, "x"); err == nil {
		t.Fatal("ReadObject expected error")
	}
	if err := d.DeleteObject(ctx, "x"); err == nil {
		t.Fatal("DeleteObject expected error")
	}
	if _, err := d.ObjectExists(ctx, "x"); err == nil {
		t.Fatal("ObjectExists expected error")
	}
	if _, err := d.ListObjects(ctx, "p/"); err == nil {
		t.Fatal("ListObjects expected error")
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := d.WriteObject(canceled, "x", []byte("y"), "text/plain"); err == nil {
		t.Fatal("WriteObject expected error on canceled context")
	}
	if err := d.WriteObject(canceled, "x", []byte("y"), ""); err == nil {
		t.Fatal("WriteObject empty content-type expected error")
	}
}
