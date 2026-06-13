package otelx

import (
	"testing"

	"github.com/turahe/pkg/config"
)

func TestNormalizeExporter(t *testing.T) {
	tests := map[string]string{
		"":            exporterOTLP,
		"otlp":        exporterOTLP,
		"OTLP_HTTP":   exporterOTLP,
		"gcp":         exporterGCP,
		"google":      exporterGCP,
		"cloudtrace":  exporterGCP,
	}
	for in, want := range tests {
		if got := normalizeExporter(in); got != want {
			t.Errorf("normalizeExporter(%q) = %q, want %q", in, got, want)
		}
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
}
