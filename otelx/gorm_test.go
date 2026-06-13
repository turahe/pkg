package otelx

import (
	"testing"

	"github.com/turahe/pkg/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTracingEnabled(t *testing.T) {
	if TracingEnabled(config.OpenTelemetryConfiguration{}) {
		t.Fatal("expected tracing disabled")
	}
	if !TracingEnabled(config.OpenTelemetryConfiguration{Endpoint: "localhost:4318"}) {
		t.Fatal("expected tracing enabled")
	}
	if !TracingEnabled(config.OpenTelemetryConfiguration{Exporter: "gcp"}) {
		t.Fatal("expected tracing enabled for gcp exporter")
	}
}

func TestGORMEnabled(t *testing.T) {
	if GORMEnabled(config.OpenTelemetryConfiguration{GORMEnabled: true}) {
		t.Fatal("expected gorm disabled without endpoint")
	}
	cfg := config.OpenTelemetryConfiguration{
		GORMEnabled: true,
		Endpoint:    "localhost:4318",
	}
	if !GORMEnabled(cfg) {
		t.Fatal("expected gorm enabled")
	}
}

func TestRegisterGORM_SQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Skipf("sqlite not available: %v", err)
	}
	if err := RegisterGORM(db, GORMOptions{DBSystem: "sqlite"}); err != nil {
		t.Fatalf("RegisterGORM() error = %v", err)
	}
}
