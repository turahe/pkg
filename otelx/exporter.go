package otelx

import (
	"context"
	"fmt"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/turahe/pkg/config"
	"github.com/turahe/pkg/logger"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/api/option"
)

func buildResource(ctx context.Context, cfg config.OpenTelemetryConfiguration, serviceName string) (*resource.Resource, error) {
	opts := []resource.Option{
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	}
	if isGCPExporter(cfg) {
		opts = append(opts, resource.WithDetectors(gcp.NewDetector()))
	}
	return resource.New(ctx, opts...)
}

func buildTraceExporter(ctx context.Context, cfg config.OpenTelemetryConfiguration) (sdktrace.SpanExporter, error) {
	if isGCPExporter(cfg) {
		return newGCPTraceExporter(ctx, cfg)
	}
	return newOTLPTraceExporter(ctx, cfg)
}

func newOTLPTraceExporter(ctx context.Context, cfg config.OpenTelemetryConfiguration) (sdktrace.SpanExporter, error) {
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
		return nil, fmt.Errorf("otlp exporter: %w", err)
	}
	return exporter, nil
}

func newGCPTraceExporter(ctx context.Context, cfg config.OpenTelemetryConfiguration) (sdktrace.SpanExporter, error) {
	opts := []texporter.Option{
		texporter.WithTraceClientOptions([]option.ClientOption{option.WithTelemetryDisabled()}),
	}
	if cfg.GCPProjectID != "" {
		opts = append(opts, texporter.WithProjectID(cfg.GCPProjectID))
	}
	exporter, err := texporter.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("gcp trace exporter: %w", err)
	}
	logger.Infof("otel: gcp trace exporter configured project=%s", cfg.GCPProjectID)
	return exporter, nil
}
