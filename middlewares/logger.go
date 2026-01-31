package middlewares

import (
	"time"

	"github.com/turahe/pkg/logger"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware logs HTTP request details including method, path, status code, latency, and client IP
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

		// Log request details (use request context so trace_id/correlation_id appear in JSON)
		reqCtx := ctx.Request.Context()
		errorMsg := ctx.Errors.String()
		if statusCode >= 500 {
			// Server errors
			if errorMsg != "" {
				logger.ErrorfContext(reqCtx,
					"[%s] %s %s %d %v %s - Error: %s",
					method,
					path,
					clientIP,
					statusCode,
					latency,
					ctx.Request.UserAgent(),
					errorMsg,
				)
			} else {
				logger.ErrorfContext(reqCtx,
					"[%s] %s %s %d %v %s",
					method,
					path,
					clientIP,
					statusCode,
					latency,
					ctx.Request.UserAgent(),
				)
			}
		} else if statusCode >= 400 {
			// Client errors
			if errorMsg != "" {
				logger.WarnfContext(reqCtx,
					"[%s] %s %s %d %v %s - Error: %s",
					method,
					path,
					clientIP,
					statusCode,
					latency,
					ctx.Request.UserAgent(),
					errorMsg,
				)
			} else {
				logger.WarnfContext(reqCtx,
					"[%s] %s %s %d %v %s",
					method,
					path,
					clientIP,
					statusCode,
					latency,
					ctx.Request.UserAgent(),
				)
			}
		} else {
			// Success (2xx, 3xx)
			logger.InfofContext(reqCtx,
				"[%s] %s %s %d %v %s",
				method,
				path,
				clientIP,
				statusCode,
				latency,
				ctx.Request.UserAgent(),
			)
		}
	}
}
