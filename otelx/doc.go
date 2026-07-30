/*
Package otelx wraps OpenTelemetry trace setup with configuration from config
so the rest of the service does not need to know about OTel's surface area.

Role in architecture:
  - Infrastructure: configures the global TracerProvider and optional GORM plugin.
  - Settings are loaded via config.GetConfig().OpenTelemetry (or config.Setup).

Responsibilities:
  - Init: build resource, exporter (OTLP HTTP or Google Cloud Trace), sampler, and propagator.
  - TracingEnabled: report whether tracing should start (OTLP requires a non-empty endpoint).
  - RegisterGORM / GORMEnabled: optional SQL spans when OTEL_GORM_ENABLED is true.
  - Return a shutdown func for graceful TracerProvider teardown.

Constraints:
  - Tracing is disabled when OTEL_TRACES_EXPORTER=otlp and the OTLP endpoint is empty.
  - Set OTEL_TRACES_EXPORTER=gcp to export to Google Cloud Trace via Application Default Credentials.
  - Invalid or failed Init is logged and treated as off; the process must keep running.
  - Sampler uses parent-based TraceIDRatioBased from OTEL_TRACES_SAMPLER_ARG (default 1.0).

This package must NOT:
  - Contain use-case or domain logic.
  - Panic or abort startup when tracing cannot be enabled.
*/
package otelx
