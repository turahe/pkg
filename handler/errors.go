package handler

import "github.com/turahe/pkg/domain"

// ErrNotFound and ErrUnauthorized re-export domain sentinel errors for backward compatibility.
// Prefer domain.ErrNotFound and domain.ErrUnauthorized in use cases; handlers may use either with errors.Is.
var (
	ErrNotFound     = domain.ErrNotFound
	ErrUnauthorized = domain.ErrUnauthorized
)
