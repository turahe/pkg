package domain

import "errors"

// Sentinel errors for use cases and handlers. Use with errors.Is so handlers can map to HTTP status.
//
// ErrNotFound: return when an entity is not found (e.g. repo.GetByID returns notFound).
// ErrUnauthorized: return when the operation requires authentication or the credentials are invalid.
//
// Wrap with fmt.Errorf("context: %w", ErrNotFound) when adding context; handlers should still use errors.Is(err, domain.ErrNotFound).
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
)
