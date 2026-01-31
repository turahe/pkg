package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/turahe/pkg/logger"
)

// Header names for trace and correlation IDs (configurable via options if needed)
const (
	HeaderTraceID       = "X-Trace-Id"
	HeaderCorrelationID = "X-Correlation-Id"
	HeaderRequestID     = "X-Request-Id"
)

// TraceMiddleware injects trace_id and correlation_id into the request context
// so they appear in structured logs. It reads X-Trace-Id and X-Correlation-Id
// from incoming headers, or X-Request-Id as fallback for trace ID. If no trace ID
// is provided, one is generated. Downstream handlers and the logger middleware
// can use the request context to get these values.
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		ctx := req.Context()

		traceID := req.Header.Get(HeaderTraceID)
		if traceID == "" {
			traceID = req.Header.Get(HeaderRequestID)
		}
		if traceID == "" {
			traceID = uuid.New().String()
		}
		ctx = logger.WithTraceID(ctx, traceID)

		correlationID := req.Header.Get(HeaderCorrelationID)
		if correlationID == "" {
			correlationID = traceID
		}
		ctx = logger.WithCorrelationID(ctx, correlationID)

		c.Request = req.WithContext(ctx)

		// Expose trace ID in response header for clients to correlate
		c.Header(HeaderTraceID, traceID)
		c.Header(HeaderCorrelationID, correlationID)

		c.Next()
	}
}
