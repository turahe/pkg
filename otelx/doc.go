/*
Package otelx wraps OpenTelemetry trace setup with configuration from config
so the rest of the service does not need to know about OTel's surface area.

Role in architecture:
  - Infrastructure: configures the global TracerProvider and optional GORM/gRPC instrumentation.
  - Settings are loaded via config.GetConfig().OpenTelemetry (or config.Setup).

Responsibilities:
  - Init: build resource, exporter (OTLP HTTP, OTLP gRPC, or Google Cloud Trace), sampler, and propagator.
  - TracingEnabled: report whether tracing should start (OTLP requires a non-empty endpoint).
  - RegisterGORM / GORMEnabled: optional SQL spans when OTEL_GORM_ENABLED is true.
  - GRPCServerOption / GRPCDialOption: otelgrpc stats handlers for gRPC servers and clients.
  - Return a shutdown func for graceful TracerProvider teardown.

Constraints:
  - Tracing is disabled when OTLP exporter is selected and the OTLP endpoint is empty.
  - Set OTEL_TRACES_EXPORTER=gcp to export to Google Cloud Trace via Application Default Credentials.
  - Set OTEL_TRACES_EXPORTER=otlp_grpc or OTEL_EXPORTER_OTLP_PROTOCOL=grpc for OTLP over gRPC (default port 4317).
  - Invalid or failed Init is logged and treated as off; the process must keep running.
  - Sampler uses parent-based TraceIDRatioBased from OTEL_TRACES_SAMPLER_ARG (default 1.0).

This package must NOT:
  - Contain use-case or domain logic.
  - Panic or abort startup when tracing cannot be enabled.
*/
package otelx
