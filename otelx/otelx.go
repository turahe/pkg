package otelx

import (
	"context"
	"strings"
	"time"

	"github.com/turahe/pkg/config"
	"github.com/turahe/pkg/logger"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const defaultServiceName = "app"

// LoadConfig returns OpenTelemetry settings from the global config package.
func LoadConfig() config.OpenTelemetryConfiguration {
	return config.GetConfig().OpenTelemetry
}

// Init configures the global OpenTelemetry TracerProvider. Returns a shutdown
// function and a boolean (enabled). When tracing is not configured we report "off"
// and proceed silently; when configured but invalid we log and treat as "off".
func Init(ctx context.Context, cfg config.OpenTelemetryConfiguration) (func(context.Context) error, bool) {
	noopShutdown := func(context.Context) error { return nil }

	if !TracingEnabled(cfg) {
		logger.Infof("otel: disabled (exporter=%s endpoint empty)", normalizeExporter(cfg.Exporter))
		return noopShutdown, false
	}

	serviceName := firstNonEmpty(cfg.ServiceName, defaultServiceName)
	shutdownTimeout := cfg.ShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = 5 * time.Second
	}

	res, err := buildResource(ctx, cfg, serviceName)
	if err != nil {
		logger.Errorf("otel: resource init failed: %v", err)
		return noopShutdown, false
	}

	exporter, err := buildTraceExporter(ctx, cfg)
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
	otel.SetTextMapPropagator(buildTextMapPropagator(useGCPPropagator(cfg)))

	logger.Infof("otel: enabled exporter=%s service=%s environment=%s version=%s sampler=%.2f target=%s gcp_propagator=%t",
		normalizeExporter(cfg.Exporter), serviceName, cfg.Environment, cfg.ServiceVersion, samplerArg,
		exporterDescription(cfg), useGCPPropagator(cfg))

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
			return strings.TrimSpace(v)
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
