// Package domain holds domain errors and ports. It has no dependencies on
// infrastructure (database, HTTP, config) so use cases can depend only on domain.
package domain

import "errors"

// Sentinel errors for use cases and handlers. Return these (e.g. with
// fmt.Errorf("context: %w", ErrNotFound)) so handlers can map to HTTP responses.
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
)
