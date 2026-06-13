package otelx

import (
	"fmt"
	"strings"

	"github.com/turahe/pkg/config"
	"github.com/turahe/pkg/logger"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

// GORMOptions configures OpenTelemetry instrumentation for a *gorm.DB instance.
type GORMOptions struct {
	// DBSystem sets the db.system semantic convention (e.g. mysql, postgresql).
	DBSystem string
	// Metrics enables DB pool metrics from the plugin. Default false (tracing only).
	Metrics bool
}

// TracingEnabled reports whether OTLP tracing is configured (endpoint set).
func TracingEnabled(cfg config.OpenTelemetryConfiguration) bool {
	return strings.TrimSpace(cfg.Endpoint) != "" || strings.TrimSpace(cfg.TracesEndpoint) != ""
}

// GORMEnabled reports whether GORM instrumentation should be registered from config.
func GORMEnabled(cfg config.OpenTelemetryConfiguration) bool {
	return cfg.GORMEnabled && TracingEnabled(cfg)
}

// RegisterGORM installs the official GORM OpenTelemetry plugin on db.
// Call after otelx.Init so spans use the configured TracerProvider.
func RegisterGORM(db *gorm.DB, opts GORMOptions) error {
	if db == nil {
		return fmt.Errorf("gorm db is nil")
	}

	pluginOpts := []tracing.Option{
		tracing.WithTracerProvider(otel.GetTracerProvider()),
	}
	if opts.DBSystem != "" {
		pluginOpts = append(pluginOpts, tracing.WithDBSystem(opts.DBSystem))
	}
	if !opts.Metrics {
		pluginOpts = append(pluginOpts, tracing.WithoutMetrics())
	}
	if err := db.Use(tracing.NewPlugin(pluginOpts...)); err != nil {
		return fmt.Errorf("register gorm otel plugin: %w", err)
	}

	logger.Infof("otel: gorm instrumentation enabled db.system=%s metrics=%t", opts.DBSystem, opts.Metrics)
	return nil
}
