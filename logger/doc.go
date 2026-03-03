/*
Package logger provides an enterprise observability foundation for fintech microservices:
structured logging (log/slog), Cloud Trace integration, correlation, and request instrumentation.

## Google Cloud Compliance

Logs follow the official Google Cloud structured logging format:

  - severity, message, timestamp (RFC3339Nano)
  - logging.googleapis.com/trace (format: "projects/{PROJECT_ID}/traces/{TRACE_ID}")
  - logging.googleapis.com/spanId, logging.googleapis.com/sourceLocation
  - httpRequest (GCP HttpRequest shape when in HTTP context)
  - correlation_id (business transaction ID; never overwritten by middleware)

Trace is only injected when a valid trace ID and project ID exist (env GOOGLE_CLOUD_PROJECT or Config.ProjectID).

## Initialization

Call Init once at startup (e.g. in main or wire):

  cfg := logger.Config{
      LogLevel:           slog.LevelInfo,
      EnableCaller:       false, // set true in dev for file/line
      EnableHTTPLogging:  true,
      ProjectID:          os.Getenv("GOOGLE_CLOUD_PROJECT"),
      ServiceName:       "payments-api",
      ServiceVersion:     version.Version,
      Environment:        "production",
      Redact:             myRedactFunc, // optional: redact sensitive fields
      ErrorStacktrace:    false,
  }
  logger.Init(cfg)

## Usage in Handlers

After trace and HTTP instrumentation middleware have run:

  log := logger.WithContext(c.Request.Context())
  log.Infof("payment processed: %s", paymentID)
  log.Info("payment", logger.Fields{"payment_id": paymentID, "amount": 100})

Structured errors (type, cause chain, optional stack):

  log := logger.WithContext(c.Request.Context())
  if err != nil {
      log.ErrorStructured(err)
  }

## Correlation Strategy

  - traceID: distributed tracing (X-Cloud-Trace-Context, X-Trace-Id, X-Request-ID)
  - spanID: request span (from X-Cloud-Trace-Context)
  - correlationID: business transaction (order_id, payment_id, etc.); never overwritten when already set

## OpenTelemetry Integration Point

This package does not implement OpenTelemetry. To bridge with OTel:

  - Extract trace/span from otel span context and set them via WithTraceID/WithSpanID in context.
  - Or use go.opentelemetry.io/contrib/instrumentation to populate X-Cloud-Trace-Context
    and let CloudTraceMiddleware parse it; logs will then link to the same trace in GCP.

## Production Safety

  - No os.Exit in request context; Fatalf is for startup failures only.
  - Recovery middleware logs CRITICAL with trace/correlation preserved.
  - Redact hook prevents sensitive data in logs (audit-grade for fintech).
*/

package logger
