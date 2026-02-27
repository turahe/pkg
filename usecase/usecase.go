package usecase

import "context"

// Runner runs a use case with the given context. Return domain.ErrNotFound, domain.ErrUnauthorized,
// or other errors for the handler to map to HTTP responses via BaseHandler.HandleServiceError.
type Runner interface {
	Run(ctx context.Context) error
}

// Func adapts a function to the Runner interface.
type Func func(ctx context.Context) error

// Run invokes f(ctx).
func (f Func) Run(ctx context.Context) error {
	return f(ctx)
}
