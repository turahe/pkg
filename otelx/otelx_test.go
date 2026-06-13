package otelx

import (
	"testing"

	"github.com/turahe/pkg/config"
)

func TestNormalizeEndpoint(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"localhost:4318", "localhost:4318"},
		{"http://localhost:4318", "localhost:4318"},
		{"https://collector.example.com:443/v1/traces", "collector.example.com:443"},
	}
	for _, tt := range tests {
		if got := normalizeEndpoint(tt.in); got != tt.want {
			t.Errorf("normalizeEndpoint(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestInit_DisabledWhenEndpointEmpty(t *testing.T) {
	shutdown, enabled := Init(t.Context(), config.OpenTelemetryConfiguration{})
	if enabled {
		t.Fatal("expected tracing disabled")
	}
	if err := shutdown(t.Context()); err != nil {
		t.Fatalf("noop shutdown: %v", err)
	}
}
