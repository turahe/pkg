package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/turahe/pkg/logger"
)

func TestTraceMiddleware(t *testing.T) {
	t.Run("generates trace_id and correlation_id when headers absent", func(t *testing.T) {
		var gotTraceID, gotCorrelationID string
		router := setupRouter()
		router.Use(TraceMiddleware())
		router.GET("/", func(c *gin.Context) {
			gotTraceID = logger.GetTraceID(c.Request.Context())
			gotCorrelationID = logger.GetCorrelationID(c.Request.Context())
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, gotTraceID)
		assert.Equal(t, gotTraceID, gotCorrelationID)
		assert.NotEmpty(t, w.Header().Get(HeaderTraceID))
		assert.NotEmpty(t, w.Header().Get(HeaderCorrelationID))
	})

	t.Run("uses X-Trace-Id from request header", func(t *testing.T) {
		var gotTraceID string
		router := setupRouter()
		router.Use(TraceMiddleware())
		router.GET("/", func(c *gin.Context) {
			gotTraceID = logger.GetTraceID(c.Request.Context())
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(HeaderTraceID, "trace-abc-123")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "trace-abc-123", gotTraceID)
		assert.Equal(t, "trace-abc-123", w.Header().Get(HeaderTraceID))
	})

	t.Run("uses X-Request-Id as fallback for trace_id", func(t *testing.T) {
		var gotTraceID string
		router := setupRouter()
		router.Use(TraceMiddleware())
		router.GET("/", func(c *gin.Context) {
			gotTraceID = logger.GetTraceID(c.Request.Context())
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(HeaderRequestID, "req-xyz-456")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "req-xyz-456", gotTraceID)
		assert.Equal(t, "req-xyz-456", w.Header().Get(HeaderTraceID))
	})

	t.Run("uses X-Correlation-Id from request header", func(t *testing.T) {
		var gotCorrelationID string
		router := setupRouter()
		router.Use(TraceMiddleware())
		router.GET("/", func(c *gin.Context) {
			gotCorrelationID = logger.GetCorrelationID(c.Request.Context())
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(HeaderTraceID, "trace-1")
		req.Header.Set(HeaderCorrelationID, "corr-999")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "corr-999", gotCorrelationID)
		assert.Equal(t, "corr-999", w.Header().Get(HeaderCorrelationID))
	})
}

func TestRequestID(t *testing.T) {
	t.Run("generates request ID when headers absent", func(t *testing.T) {
		var gotTraceID, gotCorrelationID string
		router := setupRouter()
		router.Use(RequestID())
		router.GET("/", func(c *gin.Context) {
			gotTraceID = logger.GetTraceID(c.Request.Context())
			gotCorrelationID = logger.GetCorrelationID(c.Request.Context())
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, gotTraceID)
		assert.Equal(t, gotTraceID, gotCorrelationID)
		assert.NotEmpty(t, w.Header().Get(HeaderRequestID))
		assert.NotEmpty(t, w.Header().Get(HeaderTraceID))
	})

	t.Run("uses X-Request-ID from request header", func(t *testing.T) {
		var gotTraceID string
		router := setupRouter()
		router.Use(RequestID())
		router.GET("/", func(c *gin.Context) {
			gotTraceID = logger.GetTraceID(c.Request.Context())
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(HeaderRequestID, "req-from-header")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "req-from-header", gotTraceID)
		assert.Equal(t, "req-from-header", w.Header().Get(HeaderRequestID))
		assert.Equal(t, "req-from-header", w.Header().Get(HeaderTraceID))
	})

	t.Run("uses X-Trace-ID when X-Request-ID missing", func(t *testing.T) {
		var gotTraceID string
		router := setupRouter()
		router.Use(RequestID())
		router.GET("/", func(c *gin.Context) {
			gotTraceID = logger.GetTraceID(c.Request.Context())
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(HeaderTraceID, "trace-from-header")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "trace-from-header", gotTraceID)
		assert.Equal(t, "trace-from-header", w.Header().Get(HeaderRequestID))
		assert.Equal(t, "trace-from-header", w.Header().Get(HeaderTraceID))
	})

	t.Run("X-Request-ID takes precedence over X-Trace-ID", func(t *testing.T) {
		var gotTraceID string
		router := setupRouter()
		router.Use(RequestID())
		router.GET("/", func(c *gin.Context) {
			gotTraceID = logger.GetTraceID(c.Request.Context())
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(HeaderRequestID, "request-id-wins")
		req.Header.Set(HeaderTraceID, "trace-id-ignored")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "request-id-wins", gotTraceID)
	})

	t.Run("request ID available to logger via context", func(t *testing.T) {
		var loggedCtx context.Context
		router := setupRouter()
		router.Use(RequestID(), func(c *gin.Context) {
			loggedCtx = c.Request.Context()
			c.Next()
		})
		router.Use(LoggerMiddleware())
		router.GET("/", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(HeaderRequestID, "log-me")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "log-me", logger.GetTraceID(loggedCtx))
		assert.Equal(t, "log-me", logger.GetCorrelationID(loggedCtx))
	})
}
