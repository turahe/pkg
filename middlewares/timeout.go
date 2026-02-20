package middlewares

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestTimeout returns a middleware that sets a deadline on the request context.
// Handlers and use cases that pass this context to repositories will have queries
// bounded by the timeout; when the timeout is exceeded, context is cancelled and
// operations should return context.DeadlineExceeded.
// Use after trace/request-id so the timeout applies to the full handler chain.
func RequestTimeout(d time.Duration) gin.HandlerFunc {
	if d <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
