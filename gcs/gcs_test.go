package gcs

import (
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
