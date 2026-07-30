# logger

Structured logging on Go `slog` with GCP Cloud Trace–compatible fields, correlation IDs, and optional redaction.

**Import:** `github.com/turahe/pkg/logger`

## Role

Infrastructure logging used by middleware and other packages. Bridge to OTel is documented; this package does not implement tracing.

## Init

```go
logger.Init(logger.Config{
    Level:     "info",
    Format:    "json", // or "text"
    ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
})
```

## Usage

```go
logger.Infof("started port=%s", port)
logger.ErrorContext(ctx, "failed", "err", err)

ctx = logger.WithTraceID(ctx, traceID)
ctx = logger.WithCorrelationID(ctx, corrID)
logger.InfofContext(ctx, "handled") // includes trace/correlation fields
```

Package-level: `Debugf` / `Infof` / `Warnf` / `Errorf` / `Fatalf`, plus `*Context` variants and structured `Debug` / `Info` / `Warn` / `Error`.

`Ctx` helpers and `ErrorStructured` / `ErrorStructuredContext` for richer payloads. `HTTPRequest` fields for request logging.

## Constraints

- Never overwrite an existing `correlation_id` in context.
- Use `Fatalf` only for fatal startup failures.
- Redact sensitive fields via `Config.Redact` / `RedactFunc`.
- Runnable examples: `ExampleInit`, `ExampleWithContext`, `ExampleConfig_Redact` in `logger/example_test.go`.

## See also

- [middlewares — LoggerMiddleware](middlewares.md)
- [otelx](otelx.md)
