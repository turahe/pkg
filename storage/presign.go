package storage

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultPresignUploadExpiry = 15 * time.Minute
	MaxPresignUploadExpiry     = 7 * 24 * time.Hour
)

// PresignUploadOptions configures a client-side PUT upload URL.
type PresignUploadOptions struct {
	// Expiry is how long the URL remains valid.
	// Zero means DefaultPresignUploadExpiry (15m).
	Expiry time.Duration

	// ContentType, when non-empty, is included in the signature.
	// The client MUST send the same Content-Type header on PUT.
	ContentType string
}

// PresignedUpload is a one-shot client upload grant.
type PresignedUpload struct {
	URL       string
	Method    string
	Headers   map[string]string
	ExpiresAt time.Time
}

func resolvePresignUploadOptions(opts PresignUploadOptions) (PresignUploadOptions, error) {
	if opts.Expiry < 0 {
		return PresignUploadOptions{}, fmt.Errorf("storage: expiry must be >= 0")
	}
	if opts.Expiry > MaxPresignUploadExpiry {
		return PresignUploadOptions{}, fmt.Errorf("storage: expiry exceeds maximum %s", MaxPresignUploadExpiry)
	}
	if opts.Expiry == 0 {
		opts.Expiry = DefaultPresignUploadExpiry
	}
	return opts, nil
}

// PresignUpload returns a PUT URL for objectName using the active driver.
func PresignUpload(ctx context.Context, objectName string, opts PresignUploadOptions) (PresignedUpload, error) {
	objectName = strings.TrimSpace(objectName)
	if objectName == "" {
		return PresignedUpload{}, fmt.Errorf("storage: object name required")
	}
	resolved, err := resolvePresignUploadOptions(opts)
	if err != nil {
		return PresignedUpload{}, err
	}
	d, err := current()
	if err != nil {
		return PresignedUpload{}, err
	}
	return d.PresignUpload(ctx, objectName, resolved)
}
