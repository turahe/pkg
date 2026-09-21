package otelx

import (
	"context"
	"fmt"

	"github.com/turahe/pkg/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func buildResource(ctx context.Context, cfg config.OpenTelemetryConfiguration, serviceName string) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
}

func buildTraceExporter(ctx context.Context, cfg config.OpenTelemetryConfiguration) (sdktrace.SpanExporter, error) {
	if useOTLPGRPC(cfg) {
		return newOTLPGRPCTraceExporter(ctx, cfg)
	}
	return newOTLPHTTPTraceExporter(ctx, cfg)
}

func newOTLPHTTPTraceExporter(ctx context.Context, cfg config.OpenTelemetryConfiguration) (sdktrace.SpanExporter, error) {
	opts := []otlptracehttp.Option{}
	endpoint := normalizeEndpoint(firstNonEmpty(cfg.TracesEndpoint, cfg.Endpoint))
	opts = append(opts, otlptracehttp.WithEndpoint(endpoint))
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.Headers))
	}
	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("otlp http exporter: %w", err)
	}
	return exporter, nil
}

func newOTLPGRPCTraceExporter(ctx context.Context, cfg config.OpenTelemetryConfiguration) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{}
	endpoint := normalizeEndpoint(firstNonEmpty(cfg.TracesEndpoint, cfg.Endpoint))
	opts = append(opts, otlptracegrpc.WithEndpoint(endpoint))
	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(cfg.Headers))
	}
	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("otlp grpc exporter: %w", err)
	}
	return exporter, nil
}
