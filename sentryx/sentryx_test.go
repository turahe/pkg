package sentryx

import (
	"testing"

	"github.com/turahe/pkg/config"
)

func TestInit_DisabledWhenDSNEmpty(t *testing.T) {
	if Init(config.SentryConfiguration{}) {
		t.Fatal("expected Sentry disabled")
	}
}

func TestLoadConfig_FromGlobalConfig(t *testing.T) {
	original := config.Config
	defer func() { config.Config = original }()

	config.Config = &config.Configuration{
		Sentry: config.SentryConfiguration{
			DSN:         "https://example@sentry.io/1",
			Environment: "test",
			Release:     "v1",
		},
	}

	cfg := LoadConfig()
	if cfg.DSN != "https://example@sentry.io/1" {
		t.Fatalf("DSN = %q", cfg.DSN)
	}
	if cfg.Environment != "test" {
		t.Fatalf("Environment = %q", cfg.Environment)
	}
}
