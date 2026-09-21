package otelx

import (
	"strings"

	"github.com/turahe/pkg/config"
)

const (
	exporterOTLP     = "otlp"
	exporterOTLPGRPC = "otlp_grpc"

	protocolHTTP = "http/protobuf"
	protocolGRPC = "grpc"
)

func normalizeExporter(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "otlp", "otlp_http", "http", "http/protobuf", "http/json":
		return exporterOTLP
	case "otlp_grpc", "grpc":
		return exporterOTLPGRPC
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

func normalizeProtocol(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "http", "http/protobuf", "http/json", "otlp_http":
		return protocolHTTP
	case "grpc", "otlp_grpc":
		return protocolGRPC
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

// useOTLPGRPC reports whether OTLP should use the gRPC transport.
// True when Exporter is otlp_grpc/grpc, or Exporter is otlp and Protocol is grpc.
func useOTLPGRPC(cfg config.OpenTelemetryConfiguration) bool {
	switch normalizeExporter(cfg.Exporter) {
	case exporterOTLPGRPC:
		return true
	case exporterOTLP:
		return normalizeProtocol(cfg.Protocol) == protocolGRPC
	default:
		return false
	}
}

// TracingEnabled reports whether tracing export is configured.
func TracingEnabled(cfg config.OpenTelemetryConfiguration) bool {
	switch normalizeExporter(cfg.Exporter) {
	case exporterOTLP, exporterOTLPGRPC:
		return strings.TrimSpace(cfg.Endpoint) != "" || strings.TrimSpace(cfg.TracesEndpoint) != ""
	default:
		return false
	}
}

func exporterDescription(cfg config.OpenTelemetryConfiguration) string {
	endpoint := firstNonEmpty(cfg.TracesEndpoint, cfg.Endpoint)
	if useOTLPGRPC(cfg) {
		return "grpc:" + endpoint
	}
	return "http:" + endpoint
}

func exporterName(cfg config.OpenTelemetryConfiguration) string {
	if useOTLPGRPC(cfg) {
		return exporterOTLPGRPC
	}
	return exporterOTLP
}
