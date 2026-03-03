package logger

import "log/slog"

// Config configures the enterprise observability logger.
// Use Init(cfg) once at startup (e.g. in main or wire).
type Config struct {
	// LogLevel is the minimum level to log (default: slog.LevelInfo).
	LogLevel slog.Level
	// EnableCaller enables source location (file, line, function). Disable in production for performance.
	EnableCaller bool
	// EnableHTTPLogging enables inclusion of httpRequest in log entries when present in context.
	EnableHTTPLogging bool
	// ProjectID is the GCP project ID for trace format "projects/{ProjectID}/traces/{traceID}". Set from GOOGLE_CLOUD_PROJECT when empty.
	ProjectID string
	// ServiceName is the service name for log labels (optional).
	ServiceName string
	// ServiceVersion is the service version for log labels (optional).
	ServiceVersion string
	// Environment is the deployment environment (e.g. production, staging) for log labels (optional).
	Environment string
	// Redact is called for each field value before logging; return redacted value or the original. Optional.
	Redact RedactFunc
	// ErrorStacktrace enables stack traces in structured error logging. Optional.
	ErrorStacktrace bool
}

// RedactFunc redacts sensitive values. Return the redacted string or the original.
// Used for audit-safe logs in fintech (e.g. card numbers, tokens).
type RedactFunc func(key string, value interface{}) interface{}
