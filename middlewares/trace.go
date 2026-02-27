package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/turahe/pkg/logger"
)

// HeaderTraceID, HeaderCorrelationID, and HeaderRequestID are the HTTP header names used for trace and correlation IDs.
const (
	HeaderTraceID       = "X-Trace-Id"
	HeaderCorrelationID = "X-Correlation-Id"
	HeaderRequestID     = "X-Request-Id"
)

// RequestID returns Gin middleware that ensures a request/correlation ID on every request.
// It reads X-Request-ID or X-Trace-ID from headers; generates a new UUID if missing.
// The ID is stored in context (trace_id and correlation_id) so logger.WithContext
// and LoggerMiddleware include it in structured logs. Response headers X-Request-ID
// and X-Trace-ID are set for client correlation.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		ctx := req.Context()

		requestID := req.Header.Get(HeaderRequestID)
		if requestID == "" {
			requestID = req.Header.Get(HeaderTraceID)
		}
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx = logger.WithTraceID(ctx, requestID)
		ctx = logger.WithCorrelationID(ctx, requestID)
		c.Request = req.WithContext(ctx)

		c.Header(HeaderRequestID, requestID)
		c.Header(HeaderTraceID, requestID)

		c.Next()
	}
}

// TraceMiddleware returns a Gin middleware that reads X-Trace-Id and X-Correlation-Id (or X-Request-Id for trace),
// stores them in the request context via logger.WithTraceID/WithCorrelationID, and echoes them in response headers.
// If trace ID is missing, generates a UUID. Use when upstream sends distinct trace and correlation IDs.
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

		c.Header(HeaderTraceID, traceID)
		c.Header(HeaderCorrelationID, correlationID)

		c.Next()
	}
}
