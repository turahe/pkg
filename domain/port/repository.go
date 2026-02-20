// Package port defines repository and other ports (interfaces) used by use cases.
// Implementations live in infrastructure (e.g. repositories package); domain does not depend on them.
package port

import "context"

// GetByID is a minimal read port for use cases that fetch an entity by ID.
// Implementations adapt infrastructure (e.g. repositories.BaseRepository) to this interface.
// Use this in use cases so they depend only on domain; wire a concrete implementation in main.
type GetByID interface {
	GetByID(ctx context.Context, id string) (interface{}, bool, error)
}
