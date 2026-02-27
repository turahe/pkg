package port

import "context"

// GetByID is a minimal read port for use cases that fetch an entity by ID.
//
// Implementations adapt infrastructure (e.g. repositories.BaseRepository) to this interface.
// Use in use cases so they depend only on domain; wire a concrete implementation in main.
//
// Contract:
//   - ctx: used for cancellation and timeouts; implementations must pass it to storage calls.
//   - id: non-empty identifier (format is implementation-defined).
//   - Returns (value, true, nil) when found, (nil, false, nil) when not found, (nil, false, err) on error.
//
// Error contract: do not return (nil, true, err). On storage errors return (nil, false, err).
type GetByID interface {
	GetByID(ctx context.Context, id string) (interface{}, bool, error)
}
