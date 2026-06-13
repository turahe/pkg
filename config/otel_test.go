package config

import (
	"os"
	"testing"
	"time"
)

func TestParseKeyValueHeaders(t *testing.T) {
	got := parseKeyValueHeaders("Authorization=Bearer x, X-Scope=org=1")
	if got["Authorization"] != "Bearer x" {
		t.Fatalf("Authorization = %q", got["Authorization"])
	}
	if got["X-Scope"] != "org=1" {
		t.Fatalf("X-Scope = %q", got["X-Scope"])
	}
}

func TestParseFloatRatio(t *testing.T) {
	os.Setenv("_TEST_RATIO", "0.25")
	defer os.Unsetenv("_TEST_RATIO")
	if got := parseFloatRatio("_TEST_RATIO", 1.0); got != 0.25 {
		t.Fatalf("got %v, want 0.25", got)
	}
}

func TestParseDuration(t *testing.T) {
	os.Setenv("_TEST_DURATION", "10s")
	defer os.Unsetenv("_TEST_DURATION")
	if got := parseDuration("_TEST_DURATION", 5*time.Second); got != 10*time.Second {
		t.Fatalf("got %v, want 10s", got)
	}
}

func TestBuildConfigFromEnv_OpenTelemetry(t *testing.T) {
	envVars := map[string]string{
		"OTEL_EXPORTER_OTLP_ENDPOINT":        "localhost:4318",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT": "localhost:4319",
		"OTEL_EXPORTER_OTLP_INSECURE":        "false",
		"OTEL_EXPORTER_OTLP_HEADERS":         "Authorization=Bearer token",
		"OTEL_SERVICE_NAME":                  "my-service",
		"OTEL_ENVIRONMENT":                   "staging",
		"OTEL_SERVICE_VERSION":               "1.2.3",
		"OTEL_TRACES_SAMPLER_ARG":            "0.5",
		"OTEL_SHUTDOWN_TIMEOUT":              "8s",
		"OTEL_GORM_ENABLED":                  "false",
	}
	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg := buildConfigFromEnv()
	if cfg.OpenTelemetry.Endpoint != "localhost:4318" {
		t.Errorf("Endpoint = %q", cfg.OpenTelemetry.Endpoint)
	}
	if cfg.OpenTelemetry.TracesEndpoint != "localhost:4319" {
		t.Errorf("TracesEndpoint = %q", cfg.OpenTelemetry.TracesEndpoint)
	}
	if cfg.OpenTelemetry.Insecure {
		t.Error("Insecure = true, want false")
	}
	if cfg.OpenTelemetry.Headers["Authorization"] != "Bearer token" {
		t.Errorf("Headers = %#v", cfg.OpenTelemetry.Headers)
	}
	if cfg.OpenTelemetry.ServiceName != "my-service" {
		t.Errorf("ServiceName = %q", cfg.OpenTelemetry.ServiceName)
	}
	if cfg.OpenTelemetry.Environment != "staging" {
		t.Errorf("Environment = %q", cfg.OpenTelemetry.Environment)
	}
	if cfg.OpenTelemetry.ServiceVersion != "1.2.3" {
		t.Errorf("ServiceVersion = %q", cfg.OpenTelemetry.ServiceVersion)
	}
	if cfg.OpenTelemetry.TracesSamplerArg != 0.5 {
		t.Errorf("TracesSamplerArg = %v", cfg.OpenTelemetry.TracesSamplerArg)
	}
	if cfg.OpenTelemetry.ShutdownTimeout != 8*time.Second {
		t.Errorf("ShutdownTimeout = %v", cfg.OpenTelemetry.ShutdownTimeout)
	}
	if cfg.OpenTelemetry.GORMEnabled {
		t.Error("GORMEnabled = true, want false")
	}
}

func TestBuildConfigFromEnv_OpenTelemetryFallbacks(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Setenv("SENTRY_RELEASE", "abc123")
	defer os.Unsetenv("APP_ENV")
	defer os.Unsetenv("SENTRY_RELEASE")

	cfg := buildConfigFromEnv()
	if cfg.OpenTelemetry.Environment != "production" {
		t.Errorf("Environment = %q, want production", cfg.OpenTelemetry.Environment)
	}
	if cfg.OpenTelemetry.ServiceVersion != "abc123" {
		t.Errorf("ServiceVersion = %q, want abc123", cfg.OpenTelemetry.ServiceVersion)
	}
}
