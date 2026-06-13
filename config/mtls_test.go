package config

import (
	"os"
	"testing"
)

func TestBuildConfigFromEnv_MTLS(t *testing.T) {
	envVars := map[string]string{
		"MTLS_ENABLED":        "true",
		"MTLS_CA_CERT":        "/custom/ca.crt",
		"MTLS_SERVER_CERT":    "/custom/server.crt",
		"MTLS_SERVER_KEY":     "/custom/server.key",
		"MTLS_CLIENT_CERT":    "/custom/client.crt",
		"MTLS_CLIENT_KEY":     "/custom/client.key",
		"MTLS_SKIP_PATHS":     "/live,/ready",
	}
	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg := buildConfigFromEnv()
	if !cfg.MTLS.Enabled {
		t.Fatal("expected MTLS enabled")
	}
	if cfg.MTLS.CACertFile != "/custom/ca.crt" {
		t.Errorf("CACertFile = %q", cfg.MTLS.CACertFile)
	}
	if cfg.MTLS.ServerCertFile != "/custom/server.crt" {
		t.Errorf("ServerCertFile = %q", cfg.MTLS.ServerCertFile)
	}
	if cfg.MTLS.ClientCertFile != "/custom/client.crt" {
		t.Errorf("ClientCertFile = %q", cfg.MTLS.ClientCertFile)
	}
	if cfg.MTLS.SkipPaths != "/live,/ready" {
		t.Errorf("SkipPaths = %q", cfg.MTLS.SkipPaths)
	}
}

func TestBuildConfigFromEnv_MTLSDefaults(t *testing.T) {
	os.Unsetenv("MTLS_ENABLED")
	cfg := buildConfigFromEnv()
	if cfg.MTLS.Enabled {
		t.Fatal("expected MTLS disabled by default")
	}
	if cfg.MTLS.CACertFile != "/etc/mtls/ca.crt" {
		t.Errorf("CACertFile = %q", cfg.MTLS.CACertFile)
	}
	if cfg.MTLS.SkipPaths != "/live,/ready,/metrics" {
		t.Errorf("SkipPaths = %q", cfg.MTLS.SkipPaths)
	}
}
