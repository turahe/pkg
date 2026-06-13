package otelx

import (
	"strings"

	"github.com/turahe/pkg/config"
)

const (
	exporterOTLP = "otlp"
	exporterGCP  = "gcp"
)

func normalizeExporter(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "otlp", "otlp_http", "http":
		return exporterOTLP
	case "gcp", "google", "googlecloud", "cloudtrace", "cloud_trace":
		return exporterGCP
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

func isGCPExporter(cfg config.OpenTelemetryConfiguration) bool {
	return normalizeExporter(cfg.Exporter) == exporterGCP
}

// TracingEnabled reports whether tracing export is configured.
func TracingEnabled(cfg config.OpenTelemetryConfiguration) bool {
	switch normalizeExporter(cfg.Exporter) {
	case exporterGCP:
		return true
	case exporterOTLP:
		return strings.TrimSpace(cfg.Endpoint) != "" || strings.TrimSpace(cfg.TracesEndpoint) != ""
	default:
		return false
	}
}

func useGCPPropagator(cfg config.OpenTelemetryConfiguration) bool {
	return cfg.GCPPropagator || isGCPExporter(cfg)
}

func exporterDescription(cfg config.OpenTelemetryConfiguration) string {
	if isGCPExporter(cfg) {
		if cfg.GCPProjectID != "" {
			return "gcp:project=" + cfg.GCPProjectID
		}
		return "gcp:adc"
	}
	return firstNonEmpty(cfg.TracesEndpoint, cfg.Endpoint)
}
