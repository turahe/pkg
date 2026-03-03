package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/turahe/pkg/logger"
)

// GCP format: X-Cloud-Trace-Context: TRACE_ID/SPAN_ID;o=TRACE_TRUE
// See: https://cloud.google.com/trace/docs/setup#force-trace
const (
	HeaderCloudTraceContext = "X-Cloud-Trace-Context"
	HeaderTraceID           = "X-Trace-Id"
	HeaderCorrelationID     = "X-Correlation-Id"
	HeaderRequestID         = "X-Request-Id"
)

// ParseCloudTraceContext parses X-Cloud-Trace-Context header.
// Format: TRACE_ID/SPAN_ID;o=TRACE_TRUE (o=1 means sampled).
// Returns traceID, spanID, and whether the header was present.
func ParseCloudTraceContext(header string) (traceID, spanID string, ok bool) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", "", false
	}
	parts := strings.SplitN(header, ";", 2)
	main := strings.TrimSpace(parts[0])
	slash := strings.Index(main, "/")
	if slash <= 0 || slash >= len(main)-1 {
		if main != "" {
			return main, "", true
		}
		return "", "", false
	}
	traceID = strings.TrimSpace(main[:slash])
	spanID = strings.TrimSpace(main[slash+1:])
	return traceID, spanID, true
}

// CloudTraceMiddleware returns Gin middleware that extracts X-Cloud-Trace-Context,
// X-Request-ID, and X-Correlation-ID, injects traceID/spanID/correlationID into context,
// and echoes headers. Correlation ID is never overwritten when already set (business transaction ID).
func CloudTraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		ctx := req.Context()

		traceID := ""
		spanID := ""

		if v := req.Header.Get(HeaderCloudTraceContext); v != "" {
			if tid, sid, ok := ParseCloudTraceContext(v); ok {
				traceID = tid
				spanID = sid
			}
		}
		if traceID == "" {
			traceID = req.Header.Get(HeaderTraceID)
		}
		if traceID == "" {
			traceID = req.Header.Get(HeaderRequestID)
		}
		if traceID == "" {
			traceID = uuid.New().String()
		}

		ctx = logger.WithTraceID(ctx, traceID)
		if spanID != "" {
			ctx = logger.WithSpanID(ctx, spanID)
		}

		// Correlation ID: never overwrite when already set (e.g. business transaction ID).
		correlationID := logger.GetCorrelationID(ctx)
		if correlationID == "" {
			correlationID = req.Header.Get(HeaderCorrelationID)
		}
		if correlationID == "" {
			correlationID = traceID
		}
		ctx = logger.WithCorrelationID(ctx, correlationID)

		c.Request = req.WithContext(ctx)

		c.Header(HeaderTraceID, traceID)
		if spanID != "" {
			c.Header("X-Cloud-Trace-Context", traceID+"/"+spanID+";o=1")
		}
		c.Header(HeaderCorrelationID, correlationID)
		c.Header(HeaderRequestID, req.Header.Get(HeaderRequestID))
		if c.GetHeader(HeaderRequestID) == "" {
			c.Header(HeaderRequestID, traceID)
		}

		c.Next()
	}
}

// RequestID returns Gin middleware that ensures a request/correlation ID on every request.
// Prefer CloudTraceMiddleware for full GCP trace integration; use RequestID for minimal setup.
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

// TraceMiddleware returns Gin middleware that reads X-Trace-Id, X-Correlation-Id, X-Request-Id,
// and optionally X-Cloud-Trace-Context. Correlation ID is never overwritten when already set.
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		ctx := req.Context()

		traceID := ""
		spanID := ""
		if v := req.Header.Get(HeaderCloudTraceContext); v != "" {
			var ok bool
			traceID, spanID, ok = ParseCloudTraceContext(v)
			if !ok {
				traceID = ""
				spanID = ""
			}
		}
		if traceID == "" {
			traceID = req.Header.Get(HeaderTraceID)
		}
		if traceID == "" {
			traceID = req.Header.Get(HeaderRequestID)
		}
		if traceID == "" {
			traceID = uuid.New().String()
		}
		ctx = logger.WithTraceID(ctx, traceID)
		if spanID != "" {
			ctx = logger.WithSpanID(ctx, spanID)
		}

		correlationID := logger.GetCorrelationID(ctx)
		if correlationID == "" {
			correlationID = req.Header.Get(HeaderCorrelationID)
		}
		if correlationID == "" {
			correlationID = traceID
		}
		ctx = logger.WithCorrelationID(ctx, correlationID)

		c.Request = req.WithContext(ctx)

		c.Header(HeaderTraceID, traceID)
		c.Header(HeaderCorrelationID, correlationID)
		if req.Header.Get(HeaderRequestID) != "" {
			c.Header(HeaderRequestID, req.Header.Get(HeaderRequestID))
		} else {
			c.Header(HeaderRequestID, traceID)
		}

		c.Next()
	}
}
