package database

import (
	"testing"

	"github.com/turahe/pkg/config"
)

func TestDBSystemForDriver(t *testing.T) {
	tests := map[string]string{
		"mysql":            "mysql",
		"cloudsql-mysql":   "mysql",
		"postgres":         "postgresql",
		"cloudsql-postgres": "postgresql",
		"sqlite":           "sqlite",
		"sqlserver":        "mssql",
	}
	for driver, want := range tests {
		if got := dbSystemForDriver(driver); got != want {
			t.Errorf("dbSystemForDriver(%q) = %q, want %q", driver, got, want)
		}
	}
}

func TestApplyOpenTelemetryDefaults(t *testing.T) {
	original := config.Config
	defer func() { config.Config = original }()

	config.Config = &config.Configuration{
		OpenTelemetry: config.OpenTelemetryConfiguration{
			GORMEnabled: true,
			Endpoint:    "localhost:4318",
		},
	}

	got := applyOpenTelemetryDefaults(Options{})
	if !got.EnableOpenTelemetry {
		t.Fatal("expected EnableOpenTelemetry=true from config")
	}

	explicit := applyOpenTelemetryDefaults(Options{EnableOpenTelemetry: true})
	if !explicit.EnableOpenTelemetry {
		t.Fatal("expected explicit EnableOpenTelemetry to remain true")
	}

	config.Config.OpenTelemetry.GORMEnabled = false
	got = applyOpenTelemetryDefaults(Options{})
	if got.EnableOpenTelemetry {
		t.Fatal("expected EnableOpenTelemetry=false when OTEL_GORM_ENABLED=false")
	}
}
