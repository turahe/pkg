/*
Package logger provides structured logging (log/slog) with Google Cloud Logging-compatible JSON output: severity, time, message, trace_id, correlation_id, sourceLocation, and optional fields.

Role in architecture:
  - Infrastructure: used by handlers, middleware, and other packages for request-scoped and plain logging.

Responsibilities:
  - Package-level functions: Debugf, Infof, Warnf, Errorf, Fatalf and structured Debug/Info/Warn/Error with Fields.
  - Context-bound logging: WithTraceID, WithCorrelationID, GetTraceID, GetCorrelationID; WithContext(ctx) returns a logger that includes IDs in JSON.
  - Context-aware functions: DebugfContext, InfofContext, WarnfContext, ErrorfContext.
  - Lazy writer: log file is opened on first write so init() does not block.
  - SetLogLevel, GetLogger, GetWriter for configuration and testing.

Constraints:
  - Single global logger instance; no per-request logger creation beyond WithContext.
  - Output format is fixed (GCP-style JSON); no pluggable formatters.

This package must NOT:
  - Depend on database or HTTP packages; only standard library and slog.
*/
package logger
