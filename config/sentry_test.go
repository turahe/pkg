package config

import (
	"os"
	"testing"
	"time"
)

func TestBuildConfigFromEnv_Sentry(t *testing.T) {
	envVars := map[string]string{
		"SENTRY_DSN":                "https://example@sentry.io/1",
		"SENTRY_ENVIRONMENT":        "staging",
		"SENTRY_RELEASE":            "abc123",
		"SENTRY_SERVER_NAME":        "my-service",
		"SENTRY_DEBUG":              "true",
		"SENTRY_ATTACH_STACKTRACE":  "false",
		"SENTRY_SAMPLE_RATE":        "0.75",
		"SENTRY_TRACES_SAMPLE_RATE": "0.1",
		"SENTRY_FLUSH_TIMEOUT":      "3s",
	}
	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg := buildConfigFromEnv()
	if cfg.Sentry.DSN != "https://example@sentry.io/1" {
		t.Errorf("DSN = %q", cfg.Sentry.DSN)
	}
	if cfg.Sentry.Environment != "staging" {
		t.Errorf("Environment = %q", cfg.Sentry.Environment)
	}
	if cfg.Sentry.Release != "abc123" {
		t.Errorf("Release = %q", cfg.Sentry.Release)
	}
	if cfg.Sentry.ServerName != "my-service" {
		t.Errorf("ServerName = %q", cfg.Sentry.ServerName)
	}
	if !cfg.Sentry.Debug {
		t.Error("Debug = false, want true")
	}
	if cfg.Sentry.AttachStacktrace {
		t.Error("AttachStacktrace = true, want false")
	}
	if cfg.Sentry.SampleRate != 0.75 {
		t.Errorf("SampleRate = %v", cfg.Sentry.SampleRate)
	}
	if cfg.Sentry.TracesSampleRate != 0.1 {
		t.Errorf("TracesSampleRate = %v", cfg.Sentry.TracesSampleRate)
	}
	if cfg.Sentry.FlushTimeout != 3*time.Second {
		t.Errorf("FlushTimeout = %v", cfg.Sentry.FlushTimeout)
	}
}

func TestBuildConfigFromEnv_SentryEnvironmentFallback(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	cfg := buildConfigFromEnv()
	if cfg.Sentry.Environment != "production" {
		t.Errorf("Environment = %q, want production", cfg.Sentry.Environment)
	}
}

func TestBuildConfigFromEnv_SentryDefaults(t *testing.T) {
	cfg := buildConfigFromEnv()
	if cfg.Sentry.DSN != "" {
		t.Errorf("DSN = %q, want empty", cfg.Sentry.DSN)
	}
	if !cfg.Sentry.AttachStacktrace {
		t.Error("AttachStacktrace = false, want true")
	}
	if cfg.Sentry.SampleRate != 1.0 {
		t.Errorf("SampleRate = %v, want 1.0", cfg.Sentry.SampleRate)
	}
	if cfg.Sentry.TracesSampleRate != 0.0 {
		t.Errorf("TracesSampleRate = %v, want 0.0", cfg.Sentry.TracesSampleRate)
	}
	if cfg.Sentry.FlushTimeout != 2*time.Second {
		t.Errorf("FlushTimeout = %v, want 2s", cfg.Sentry.FlushTimeout)
	}
}
