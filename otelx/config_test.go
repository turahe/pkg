package otelx

import (
	"testing"

	"github.com/turahe/pkg/config"
)

func TestNormalizeExporter(t *testing.T) {
	tests := map[string]string{
		"":           exporterOTLP,
		"otlp":       exporterOTLP,
		"OTLP_HTTP":  exporterOTLP,
		"otlp_grpc":  exporterOTLPGRPC,
		"grpc":       exporterOTLPGRPC,
		"gcp":        exporterGCP,
		"google":     exporterGCP,
		"cloudtrace": exporterGCP,
	}
	for in, want := range tests {
		if got := normalizeExporter(in); got != want {
			t.Errorf("normalizeExporter(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeProtocol(t *testing.T) {
	tests := map[string]string{
		"":              protocolHTTP,
		"http/protobuf": protocolHTTP,
		"http":          protocolHTTP,
		"grpc":          protocolGRPC,
		"otlp_grpc":     protocolGRPC,
	}
	for in, want := range tests {
		if got := normalizeProtocol(in); got != want {
			t.Errorf("normalizeProtocol(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUseOTLPGRPC(t *testing.T) {
	if useOTLPGRPC(config.OpenTelemetryConfiguration{Exporter: "otlp"}) {
		t.Fatal("expected http for default otlp")
	}
	if !useOTLPGRPC(config.OpenTelemetryConfiguration{Exporter: "otlp_grpc"}) {
		t.Fatal("expected grpc for otlp_grpc exporter")
	}
	if !useOTLPGRPC(config.OpenTelemetryConfiguration{Exporter: "otlp", Protocol: "grpc"}) {
		t.Fatal("expected grpc when protocol=grpc")
	}
}

func TestTracingEnabled_GCPExporter(t *testing.T) {
	cfg := config.OpenTelemetryConfiguration{Exporter: "gcp"}
	if !TracingEnabled(cfg) {
		t.Fatal("expected tracing enabled for gcp exporter")
	}
}

func TestTracingEnabled_OTLPRequiresEndpoint(t *testing.T) {
	if TracingEnabled(config.OpenTelemetryConfiguration{}) {
		t.Fatal("expected tracing disabled without endpoint")
	}
	if !TracingEnabled(config.OpenTelemetryConfiguration{Endpoint: "localhost:4318"}) {
		t.Fatal("expected tracing enabled with endpoint")
	}
	if !TracingEnabled(config.OpenTelemetryConfiguration{Exporter: "otlp_grpc", Endpoint: "localhost:4317"}) {
		t.Fatal("expected tracing enabled for otlp_grpc with endpoint")
	}
}

func TestUseGCPPropagator(t *testing.T) {
	if !useGCPPropagator(config.OpenTelemetryConfiguration{Exporter: "gcp"}) {
		t.Fatal("expected gcp propagator for gcp exporter")
	}
	if useGCPPropagator(config.OpenTelemetryConfiguration{Exporter: "otlp"}) {
		t.Fatal("expected gcp propagator off for otlp by default")
	}
	if !useGCPPropagator(config.OpenTelemetryConfiguration{Exporter: "otlp", GCPPropagator: true}) {
		t.Fatal("expected gcp propagator when explicitly enabled")
	}
}

func TestExporterDescription(t *testing.T) {
	got := exporterDescription(config.OpenTelemetryConfiguration{
		Exporter:     "gcp",
		GCPProjectID: "my-project",
	})
	if got != "gcp:project=my-project" {
		t.Fatalf("got %q", got)
	}

	got = exporterDescription(config.OpenTelemetryConfiguration{
		Exporter: "otlp",
		Endpoint: "localhost:4318",
	})
	if got != "http:localhost:4318" {
		t.Fatalf("got %q", got)
	}

	got = exporterDescription(config.OpenTelemetryConfiguration{
		Exporter: "otlp_grpc",
		Endpoint: "localhost:4317",
	})
	if got != "grpc:localhost:4317" {
		t.Fatalf("got %q", got)
	}
}
