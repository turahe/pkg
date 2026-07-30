# otelx

OpenTelemetry TracerProvider setup from config: OTLP HTTP or Google Cloud Trace, plus optional GORM instrumentation.

**Import:** `github.com/turahe/pkg/otelx`

## Role

Infrastructure wrapper so services do not touch OTel SDK surface area directly.

## API

```go
cfg := otelx.LoadConfig() // or config.GetConfig().OpenTelemetry
shutdown, enabled := otelx.Init(ctx, cfg)
defer func() { _ = shutdown(context.Background()) }()

if otelx.GORMEnabled(cfg) {
    _ = otelx.RegisterGORM(gormDB, otelx.GORMOptions{})
}
```

`TracingEnabled(cfg)` — false when exporter is OTLP and endpoint is empty. Invalid config is logged and treated as off (service keeps running).

## Exporters

| `OTEL_TRACES_EXPORTER` | Behavior |
|------------------------|----------|
| `otlp` (default) | OTLP HTTP; requires `OTEL_EXPORTER_OTLP_ENDPOINT` |
| `gcp` | Cloud Trace via ADC; optional `OTEL_GCP_PROJECT_ID` |

Sampler: parent-based ratio from `OTEL_TRACES_SAMPLER_ARG` (default `1.0`).

## GORM

When tracing is on and `OTEL_GORM_ENABLED=true` (default), register the GORM plugin so SQL spans appear in traces. Also available via `database.WithOpenTelemetry`.

## Constraints

- Init failures must not crash the process — log and continue disabled.
- Call shutdown on graceful exit (after HTTP drain).

## See also

- [Environment — OpenTelemetry](../environment.md#opentelemetry)
- [logger](logger.md)
- [middlewares — Trace](middlewares.md)
