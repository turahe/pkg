package database

import (
	"strings"
	"testing"

	"github.com/turahe/pkg/config"
)

func TestResolveConnectionTimezone_FromConfig(t *testing.T) {
	original := config.Config
	defer func() { config.Config = original }()

	config.Config = &config.Configuration{
		Timezone: config.TimezoneConfiguration{Timezone: "UTC"},
	}

	cfg := &config.DatabaseConfiguration{ConnectionTimezone: "Asia/Jakarta"}
	got, err := resolveConnectionTimezone(cfg)
	if err != nil {
		t.Fatalf("resolveConnectionTimezone() error = %v", err)
	}
	if got != "Asia/Jakarta" {
		t.Errorf("got %q, want Asia/Jakarta", got)
	}
}

func TestResolveConnectionTimezone_FallbackServerTimezone(t *testing.T) {
	original := config.Config
	defer func() { config.Config = original }()

	config.Config = &config.Configuration{
		Timezone: config.TimezoneConfiguration{Timezone: "Asia/Singapore"},
	}

	got, err := resolveConnectionTimezone(&config.DatabaseConfiguration{})
	if err != nil {
		t.Fatalf("resolveConnectionTimezone() error = %v", err)
	}
	if got != "Asia/Singapore" {
		t.Errorf("got %q, want Asia/Singapore", got)
	}
}

func TestResolveConnectionTimezone_Invalid(t *testing.T) {
	_, err := resolveConnectionTimezone(&config.DatabaseConfiguration{
		ConnectionTimezone: "Not/A/Timezone",
	})
	if err == nil {
		t.Fatal("expected error for invalid timezone")
	}
}

func TestBuildDSN_MySQL_ConnectionTimezone(t *testing.T) {
	cfg := &config.DatabaseConfiguration{
		Driver:             "mysql",
		Host:               "localhost",
		Port:               "3306",
		Username:           "u",
		Password:           "p",
		Dbname:             "db",
		ConnectionTimezone: "Asia/Jakarta",
	}
	got, err := buildDSN(cfg)
	if err != nil {
		t.Fatalf("buildDSN() error = %v", err)
	}
	if !strings.Contains(got, "loc=Asia%2FJakarta") {
		t.Errorf("buildDSN() = %q, want loc=Asia%%2FJakarta", got)
	}
}

func TestBuildDSN_Postgres_ConnectionTimezone(t *testing.T) {
	cfg := &config.DatabaseConfiguration{
		Driver:             "postgres",
		Host:               "localhost",
		Port:               "5432",
		Username:           "u",
		Password:           "p",
		Dbname:             "db",
		ConnectionTimezone: "UTC",
	}
	got, err := buildDSN(cfg)
	if err != nil {
		t.Fatalf("buildDSN() error = %v", err)
	}
	if !strings.Contains(got, "timezone=UTC") {
		t.Errorf("buildDSN() = %q, want timezone=UTC", got)
	}
}
