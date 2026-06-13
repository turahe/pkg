package mtls

import (
	"net/http"
	"strings"
	"testing"

	"github.com/turahe/pkg/config"
)

func TestConfigureServer_disabled(t *testing.T) {
	srv := &http.Server{Addr: ":0"}
	if err := ConfigureServerWithConfig(srv, Config{Enabled: false}); err != nil {
		t.Fatalf("ConfigureServerWithConfig() err = %v", err)
	}
	if srv.TLSConfig != nil {
		t.Fatal("expected TLSConfig unset when disabled")
	}
}

func TestConfigureServer_enabled(t *testing.T) {
	cfg := testConfig(t)

	srv := &http.Server{Addr: ":0"}
	if err := ConfigureServerWithConfig(srv, cfg); err != nil {
		t.Fatalf("ConfigureServerWithConfig() err = %v", err)
	}
	if srv.TLSConfig == nil {
		t.Fatal("expected TLSConfig when enabled")
	}
}

func TestConfigureServer_enabled_missingCert(t *testing.T) {
	cfg := Config{
		Enabled:        true,
		CACertFile:     "/missing/ca.crt",
		ServerCertFile: "/etc/mtls/server.crt",
		ServerKeyFile:  "/etc/mtls/server.key",
	}

	srv := &http.Server{Addr: ":0"}
	err := ConfigureServerWithConfig(srv, cfg)
	if err == nil || !strings.Contains(err.Error(), "mtls configure server") {
		t.Fatalf("ConfigureServerWithConfig() err = %v", err)
	}
}

func TestLoadConfig(t *testing.T) {
	original := config.Config
	defer func() { config.Config = original }()

	config.Config = &config.Configuration{
		MTLS: config.MTLSConfiguration{
			Enabled:    true,
			CACertFile: "/etc/mtls/ca.crt",
		},
	}

	cfg := LoadConfig()
	if !cfg.Enabled {
		t.Fatal("expected enabled")
	}
	if cfg.CACertFile != "/etc/mtls/ca.crt" {
		t.Fatalf("CACertFile = %q", cfg.CACertFile)
	}
}
