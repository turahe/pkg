// Package usecase provides a minimal pattern for application use cases (application service layer).
// Use cases depend only on domain (domain errors, domain/port interfaces). Handlers call use cases
// with request context; use cases call repository ports and return domain errors.
// Dependency flow: handler → usecase → domain.port (implementations live in repositories or app).
package usecase

import "context"

// Runner runs a use case with the given context. Return domain.ErrNotFound, domain.ErrUnauthorized,
// or other errors for the handler to map to HTTP responses via BaseHandler.HandleServiceError.
type Runner interface {
	Run(ctx context.Context) error
}

// Func adapts a function to Runner.
type Func func(ctx context.Context) error

// Run calls f(ctx).
func (f Func) Run(ctx context.Context) error {
	return f(ctx)
}
