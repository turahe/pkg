package config

import (
	"os"
	"testing"
)

func TestGetConfig_WhenNil_BuildsFromEnv(t *testing.T) {
	// Reset so GetConfig uses buildConfigFromEnv
	Config = nil
	cfg := GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig must not return nil")
	}
	if cfg.Server.Port == "" {
		t.Error("Server.Port should have default")
	}
	if cfg.Database.Host == "" {
		t.Error("Database.Host should have default")
	}
}

func TestGetConfig_RespectsEnv(t *testing.T) {
	Config = nil
	os.Setenv("SERVER_PORT", "9999")
	defer os.Unsetenv("SERVER_PORT")
	cfg := GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig must not return nil")
	}
	if cfg.Server.Port != "9999" {
		t.Errorf("Server.Port = %q, want 9999", cfg.Server.Port)
	}
}

func TestSetup_InvalidDatabaseConfig(t *testing.T) {
	os.Setenv("DATABASE_DBNAME", "your_database_name")
	os.Setenv("DATABASE_USERNAME", "u")
	os.Setenv("DATABASE_PASSWORD", "p")
	defer func() {
		os.Unsetenv("DATABASE_DBNAME")
		os.Unsetenv("DATABASE_USERNAME")
		os.Unsetenv("DATABASE_PASSWORD")
	}()

	err := Setup("")
	if err == nil {
		t.Error("Setup should fail when DATABASE_DBNAME is placeholder")
	}
}

func TestSetup_ValidDatabaseConfig(t *testing.T) {
	os.Setenv("DATABASE_DBNAME", "testdb")
	os.Setenv("DATABASE_USERNAME", "testuser")
	os.Setenv("DATABASE_PASSWORD", "testpass")
	defer func() {
		os.Unsetenv("DATABASE_DBNAME")
		os.Unsetenv("DATABASE_USERNAME")
		os.Unsetenv("DATABASE_PASSWORD")
	}()

	err := Setup("")
	if err != nil {
		t.Errorf("Setup with valid DB config failed: %v", err)
	}
	if Config == nil {
		t.Error("Config should be set after successful Setup")
	}
}
