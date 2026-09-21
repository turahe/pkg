package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/turahe/pkg/config"
)

// memDriver is an in-memory Driver used to exercise the package facade without a backend.
type memDriver struct {
	objects map[string][]byte
	types   map[string]string
	closed  bool
	fail    error // if set, all ops return this error
}

func newMemDriver() *memDriver {
	return &memDriver{
		objects: map[string][]byte{},
		types:   map[string]string{},
	}
}

func (m *memDriver) ReadObject(ctx context.Context, objectName string) ([]byte, error) {
	r, err := m.ReadObjectAsReader(ctx, objectName)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func (m *memDriver) ReadObjectAsReader(ctx context.Context, objectName string) (io.ReadCloser, error) {
	if m.fail != nil {
		return nil, m.fail
	}
	data, ok := m.objects[objectName]
	if !ok {
		return nil, errors.New("not found")
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *memDriver) WriteObject(ctx context.Context, objectName string, data []byte, contentType string) error {
	if m.fail != nil {
		return m.fail
	}
	cp := append([]byte(nil), data...)
	m.objects[objectName] = cp
	m.types[objectName] = contentType
	return nil
}

func (m *memDriver) DeleteObject(ctx context.Context, objectName string) error {
	if m.fail != nil {
		return m.fail
	}
	delete(m.objects, objectName)
	delete(m.types, objectName)
	return nil
}

func (m *memDriver) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	if m.fail != nil {
		return false, m.fail
	}
	_, ok := m.objects[objectName]
	return ok, nil
}

func (m *memDriver) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	if m.fail != nil {
		return nil, m.fail
	}
	var names []string
	for k := range m.objects {
		if len(prefix) == 0 || (len(k) >= len(prefix) && k[:len(prefix)] == prefix) {
			names = append(names, k)
		}
	}
	return names, nil
}

func (m *memDriver) PresignUpload(ctx context.Context, objectName string, opts PresignUploadOptions) (PresignedUpload, error) {
	if m.fail != nil {
		return PresignedUpload{}, m.fail
	}
	headers := map[string]string{}
	if opts.ContentType != "" {
		headers["Content-Type"] = opts.ContentType
	}
	return PresignedUpload{
		URL:     "https://example.test/" + objectName,
		Method:  "PUT",
		Headers: headers,
	}, nil
}

func (m *memDriver) Close() error {
	m.closed = true
	return nil
}

func setTestDriver(t *testing.T, d Driver, bucket string) {
	t.Helper()
	t.Cleanup(func() {
		mu.Lock()
		driver, bucketName = nil, ""
		mu.Unlock()
	})
	mu.Lock()
	driver = d
	bucketName = bucket
	mu.Unlock()
}

func TestFacade_OpsWithMemDriver(t *testing.T) {
	m := newMemDriver()
	setTestDriver(t, m, "test-bucket")

	ctx := context.Background()
	if got := GetBucketName(); got != "test-bucket" {
		t.Fatalf("GetBucketName = %q", got)
	}

	if err := WriteObject(ctx, "a/b.txt", []byte("hello"), "text/plain"); err != nil {
		t.Fatal(err)
	}
	data, err := ReadObject(ctx, "a/b.txt")
	if err != nil || string(data) != "hello" {
		t.Fatalf("ReadObject: %v %q", err, data)
	}
	r, err := ReadObjectAsReader(ctx, "a/b.txt")
	if err != nil {
		t.Fatal(err)
	}
	buf, _ := io.ReadAll(r)
	_ = r.Close()
	if string(buf) != "hello" {
		t.Fatalf("ReadObjectAsReader: %q", buf)
	}
	exists, err := ObjectExists(ctx, "a/b.txt")
	if err != nil || !exists {
		t.Fatalf("ObjectExists: %v %v", exists, err)
	}
	names, err := ListObjects(ctx, "a/")
	if err != nil || len(names) != 1 || names[0] != "a/b.txt" {
		t.Fatalf("ListObjects: %v %v", names, err)
	}
	out, err := PresignUpload(ctx, "a/b.txt", PresignUploadOptions{ContentType: "text/plain"})
	if err != nil || out.Method != "PUT" || out.Headers["Content-Type"] != "text/plain" {
		t.Fatalf("PresignUpload: %+v %v", out, err)
	}
	err = DeleteObject(ctx, "a/b.txt")
	if err != nil {
		t.Fatal(err)
	}
	exists, err = ObjectExists(ctx, "a/b.txt")
	if err != nil || exists {
		t.Fatalf("after delete exists=%v err=%v", exists, err)
	}
}

func TestFacade_PropagatesDriverErrors(t *testing.T) {
	want := errors.New("boom")
	m := newMemDriver()
	m.fail = want
	setTestDriver(t, m, "b")

	ctx := context.Background()
	if _, err := ReadObject(ctx, "x"); !errors.Is(err, want) {
		t.Fatalf("ReadObject: %v", err)
	}
	if _, err := ReadObjectAsReader(ctx, "x"); !errors.Is(err, want) {
		t.Fatalf("ReadObjectAsReader: %v", err)
	}
	if err := WriteObject(ctx, "x", nil, ""); !errors.Is(err, want) {
		t.Fatalf("WriteObject: %v", err)
	}
	if err := DeleteObject(ctx, "x"); !errors.Is(err, want) {
		t.Fatalf("DeleteObject: %v", err)
	}
	if _, err := ObjectExists(ctx, "x"); !errors.Is(err, want) {
		t.Fatalf("ObjectExists: %v", err)
	}
	if _, err := ListObjects(ctx, ""); !errors.Is(err, want) {
		t.Fatalf("ListObjects: %v", err)
	}
	if _, err := PresignUpload(ctx, "x", PresignUploadOptions{}); !errors.Is(err, want) {
		t.Fatalf("PresignUpload: %v", err)
	}
}

func TestSetup_ClosesPreviousDriverWhenDisabled(t *testing.T) {
	m := newMemDriver()
	setTestDriver(t, m, "b")

	config.Config = &config.Configuration{}
	t.Cleanup(func() { config.Config = nil })
	if err := Setup(); err != nil {
		t.Fatal(err)
	}
	if !m.closed {
		t.Fatal("expected previous driver Close on disable")
	}
	if GetBucketName() != "" {
		t.Fatalf("bucket = %q", GetBucketName())
	}
}
