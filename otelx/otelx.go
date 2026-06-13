// Package otelx wraps OpenTelemetry trace setup with configuration from config
// so the rest of the service doesn't need to know about OTel's surface area.
//
// Settings are loaded via config.GetConfig().OpenTelemetry (or config.Setup).
// OTEL_EXPORTER_OTLP_ENDPOINT being empty is the explicit "tracing is off" signal —
// Init becomes a no-op and Shutdown returns immediately.
package otelx

import (
	"context"
	"strings"
	"time"

	"github.com/turahe/pkg/config"
	"github.com/turahe/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const defaultServiceName = "app"

// LoadConfig returns OpenTelemetry settings from the global config package.
func LoadConfig() config.OpenTelemetryConfiguration {
	return config.GetConfig().OpenTelemetry
}

// Init configures the global OpenTelemetry TracerProvider. Returns a shutdown
// function and a boolean (enabled). When Endpoint is empty we report "off" and
// proceed silently; when set but invalid we log and treat as "off".
func Init(ctx context.Context, cfg config.OpenTelemetryConfiguration) (func(context.Context) error, bool) {
	noopShutdown := func(context.Context) error { return nil }

	if strings.TrimSpace(cfg.Endpoint) == "" && strings.TrimSpace(cfg.TracesEndpoint) == "" {
		logger.Infof("otel: disabled (OTEL_EXPORTER_OTLP_ENDPOINT is empty)")
		return noopShutdown, false
	}

	serviceName := firstNonEmpty(cfg.ServiceName, defaultServiceName)
	shutdownTimeout := cfg.ShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = 5 * time.Second
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		logger.Errorf("otel: resource init failed: %v", err)
		return noopShutdown, false
	}

	opts := []otlptracehttp.Option{}
	endpoint := firstNonEmpty(cfg.TracesEndpoint, cfg.Endpoint)
	endpoint = normalizeEndpoint(endpoint)
	opts = append(opts, otlptracehttp.WithEndpoint(endpoint))
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.Headers))
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		logger.Errorf("otel: exporter init failed: %v", err)
		return noopShutdown, false
	}

	samplerArg := cfg.TracesSamplerArg
	if samplerArg <= 0 || samplerArg > 1 {
		samplerArg = 1.0
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(samplerArg))),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Infof("otel: enabled service=%s environment=%s version=%s sampler=%.2f endpoint=%s",
		serviceName, cfg.Environment, cfg.ServiceVersion, samplerArg,
		firstNonEmpty(cfg.TracesEndpoint, cfg.Endpoint))

	shutdown := func(shutdownCtx context.Context) error {
		if shutdownCtx == nil {
			var cancel context.CancelFunc
			shutdownCtx, cancel = context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()
		}
		return tp.Shutdown(shutdownCtx)
	}

	return shutdown, true
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// normalizeEndpoint strips scheme and path from OTLP endpoint values so both
// "localhost:4318" and "http://localhost:4318" work with otlptracehttp.
func normalizeEndpoint(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")
	if i := strings.Index(raw, "/"); i >= 0 {
		raw = raw[:i]
	}
	return raw
}
