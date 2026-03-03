package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/turahe/pkg/logger"
)

// responseWriter wraps gin.ResponseWriter to capture status code and response size.
type responseWriter struct {
	gin.ResponseWriter
	status int
	size   int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

// HTTPInstrumentation returns Gin middleware that captures request/response metrics
// and injects a GCP-format httpRequest into context so every log in the request includes it.
// Place after CloudTraceMiddleware or TraceMiddleware.
func HTTPInstrumentation() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		req := c.Request

		reqSize := int64(0)
		if req.ContentLength > 0 {
			reqSize = req.ContentLength
		}

		hr := &logger.HTTPRequest{
			RequestMethod: req.Method,
			RequestURL:    req.URL.String(),
			RequestSize:   reqSize,
			UserAgent:     req.UserAgent(),
			RemoteIP:      c.ClientIP(),
		}

		ctx := logger.WithHTTPRequest(req.Context(), hr)
		c.Request = req.WithContext(ctx)

		wrap := &responseWriter{
			ResponseWriter: c.Writer,
			status:         0,
		}
		c.Writer = wrap

		c.Next()

		hr.Status = wrap.status
		if hr.Status == 0 {
			hr.Status = http.StatusOK
		}
		hr.ResponseSize = int64(wrap.size)
		hr.Latency = strconv.FormatFloat(time.Since(start).Seconds(), 'f', -1, 64) + "s"
	}
}
