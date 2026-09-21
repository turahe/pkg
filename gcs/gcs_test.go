package gcs

import (
	"context"
	"testing"

	"github.com/turahe/pkg/config"
)

func TestSetup_GCSDisabled(t *testing.T) {
	config.Config = &config.Configuration{
		GCS: config.GCSConfiguration{Enabled: false},
	}
	err := Setup()
	if err != nil {
		t.Errorf("Setup with GCS disabled: %v", err)
	}
}

func TestSetupContext_GCSDisabled(t *testing.T) {
	config.Config = &config.Configuration{
		GCS: config.GCSConfiguration{Enabled: false},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// Disabled path must not touch GCS, even with a cancelled context.
	if err := SetupContext(ctx); err != nil {
		t.Errorf("SetupContext with GCS disabled: %v", err)
	}
}
