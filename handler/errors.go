package handler

import "github.com/turahe/pkg/domain"

// Re-export domain errors for backward compatibility. Prefer using domain.ErrNotFound
// and domain.ErrUnauthorized in use cases; handlers can use either handler or domain for errors.Is.
var (
	ErrNotFound     = domain.ErrNotFound
	ErrUnauthorized = domain.ErrUnauthorized
)
