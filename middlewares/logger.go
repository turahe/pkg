package middlewares

import (
	"time"

	"github.com/turahe/pkg/logger"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware returns a Gin middleware that logs each request after Next(): method, path, status,
// latency, client IP, user-agent. Uses logger.WithContext(ctx) so trace_id and correlation_id appear
// in JSON when RequestID or TraceMiddleware is used. Log level: 5xx Error, 4xx Warn, 2xx/3xx Info.
func LoggerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Start timer
		start := time.Now()
		path := ctx.Request.URL.Path
		raw := ctx.Request.URL.RawQuery

		// Process request
		ctx.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get client IP
		clientIP := ctx.ClientIP()

		// Get method and status code
		method := ctx.Request.Method
		statusCode := ctx.Writer.Status()

		// Build query string if exists
		if raw != "" {
			path = path + "?" + raw
		}

		// Log request details; use context-bound logger so trace_id/correlation_id appear in JSON
		log := logger.WithContext(ctx.Request.Context())
		errorMsg := ctx.Errors.String()
		if statusCode >= 500 {
			// Server errors
			if errorMsg != "" {
				log.Errorf("[%s] %s %s %d %v %s - Error: %s",
					method, path, clientIP, statusCode, latency, ctx.Request.UserAgent(), errorMsg)
			} else {
				log.Errorf("[%s] %s %s %d %v %s",
					method, path, clientIP, statusCode, latency, ctx.Request.UserAgent())
			}
		} else if statusCode >= 400 {
			// Client errors
			if errorMsg != "" {
				log.Warnf("[%s] %s %s %d %v %s - Error: %s",
					method, path, clientIP, statusCode, latency, ctx.Request.UserAgent(), errorMsg)
			} else {
				log.Warnf("[%s] %s %s %d %v %s",
					method, path, clientIP, statusCode, latency, ctx.Request.UserAgent())
			}
		} else {
			// Success (2xx, 3xx)
			log.Infof("[%s] %s %s %d %v %s",
				method, path, clientIP, statusCode, latency, ctx.Request.UserAgent())
		}
	}
}
