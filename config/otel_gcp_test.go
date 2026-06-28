package config

import (
	"os"
	"testing"
)

func TestBuildConfigFromEnv_OpenTelemetryGCP(t *testing.T) {
	envVars := map[string]string{
		"OTEL_TRACES_EXPORTER": "gcp",
		"OTEL_GCP_PROJECT_ID":  "my-gcp-project",
		"OTEL_GCP_PROPAGATOR":  "true",
		"OTEL_SERVICE_NAME":    "api",
	}
	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg := buildConfigFromEnv()
	if cfg.OpenTelemetry.Exporter != "gcp" {
		t.Errorf("Exporter = %q", cfg.OpenTelemetry.Exporter)
	}
	if cfg.OpenTelemetry.GCPProjectID != "my-gcp-project" {
		t.Errorf("GCPProjectID = %q", cfg.OpenTelemetry.GCPProjectID)
	}
	if !cfg.OpenTelemetry.GCPPropagator {
		t.Error("GCPPropagator = false, want true")
	}
}

func TestBuildConfigFromEnv_OpenTelemetryGCPProjectFallback(t *testing.T) {
	os.Setenv("GOOGLE_CLOUD_PROJECT", "from-env")
	defer os.Unsetenv("GOOGLE_CLOUD_PROJECT")

	cfg := buildConfigFromEnv()
	if cfg.OpenTelemetry.GCPProjectID != "from-env" {
		t.Errorf("GCPProjectID = %q, want from-env", cfg.OpenTelemetry.GCPProjectID)
	}
}
