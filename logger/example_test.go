// Package logger examples: initialization and handler usage.
// See doc.go for full documentation.
package logger_test

import (
	"log/slog"
	"os"

	"github.com/turahe/pkg/logger"
)

func ExampleInit() {
	cfg := logger.Config{
		LogLevel:          slog.LevelInfo,
		EnableCaller:      false,
		EnableHTTPLogging: true,
		ProjectID:         os.Getenv("GOOGLE_CLOUD_PROJECT"),
		ServiceName:       "my-service",
		ServiceVersion:    "1.0.0",
		Environment:       "production",
	}
	logger.Init(cfg)
}

func ExampleWithContext() {
	// In a Gin handler, after middlewares (CloudTraceMiddleware, HTTPInstrumentation):
	//   log := logger.WithContext(c.Request.Context())
	//   log.Infof("request processed: %s", id)
	//   log.Info("event", logger.Fields{"key": "value"})
	_ = logger.WithContext
}

func ExampleConfig_Redact() {
	cfg := logger.Config{
		LogLevel: slog.LevelInfo,
		Redact: func(key string, value interface{}) interface{} {
			if key == "card_number" || key == "token" {
				return "***REDACTED***"
			}
			return value
		},
	}
	logger.Init(cfg)
}
