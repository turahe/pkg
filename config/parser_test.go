package config

import (
	"os"
	"testing"
)

func TestGetEnvOrDefault_EmptyReturnsDefault(t *testing.T) {
	os.Unsetenv("_TEST_GETENV_KEY")
	if got := getEnvOrDefault("_TEST_GETENV_KEY", "default"); got != "default" {
		t.Errorf("getEnvOrDefault(_, %q) = %q, want %q", "default", got, "default")
	}
}

func TestGetEnvOrDefault_SetReturnsValue(t *testing.T) {
	os.Setenv("_TEST_GETENV_KEY", "value")
	defer os.Unsetenv("_TEST_GETENV_KEY")
	if got := getEnvOrDefault("_TEST_GETENV_KEY", "default"); got != "value" {
		t.Errorf("getEnvOrDefault = %q, want %q", got, "value")
	}
}

func TestGetEnvOrDefault_StripsDoubleQuotes(t *testing.T) {
	os.Setenv("_TEST_GETENV_KEY", `"quoted"`)
	defer os.Unsetenv("_TEST_GETENV_KEY")
	if got := getEnvOrDefault("_TEST_GETENV_KEY", "x"); got != "quoted" {
		t.Errorf("getEnvOrDefault = %q, want %q", got, "quoted")
	}
}

func TestGetEnvOrDefault_StripsSingleQuotes(t *testing.T) {
	os.Setenv("_TEST_GETENV_KEY", "'quoted'")
	defer os.Unsetenv("_TEST_GETENV_KEY")
	if got := getEnvOrDefault("_TEST_GETENV_KEY", "x"); got != "quoted" {
		t.Errorf("getEnvOrDefault = %q, want %q", got, "quoted")
	}
}

func TestGetEnvOrDefault_NoStripWhenMismatchedQuotes(t *testing.T) {
	os.Setenv("_TEST_GETENV_KEY", `"a`)
	defer os.Unsetenv("_TEST_GETENV_KEY")
	want := "\"a"
	if got := getEnvOrDefault("_TEST_GETENV_KEY", "x"); got != want {
		t.Errorf("getEnvOrDefault = %q, want %q (no strip when only leading quote)", got, want)
	}
}

func TestParseBool_EmptyReturnsDefault(t *testing.T) {
	os.Unsetenv("_TEST_PARSE_BOOL")
	if got := parseBool("_TEST_PARSE_BOOL", true); got != true {
		t.Errorf("parseBool(_, true) = %v, want true", got)
	}
	if got := parseBool("_TEST_PARSE_BOOL", false); got != false {
		t.Errorf("parseBool(_, false) = %v, want false", got)
	}
}

func TestParseBool_ValidTrue(t *testing.T) {
	for _, v := range []string{"true", "TRUE", "1", "t"} {
		os.Setenv("_TEST_PARSE_BOOL", v)
		if got := parseBool("_TEST_PARSE_BOOL", false); got != true {
			t.Errorf("parseBool(%q, false) = %v, want true", v, got)
		}
	}
	os.Unsetenv("_TEST_PARSE_BOOL")
}

func TestParseBool_ValidFalse(t *testing.T) {
	for _, v := range []string{"false", "FALSE", "0", "f"} {
		os.Setenv("_TEST_PARSE_BOOL", v)
		if got := parseBool("_TEST_PARSE_BOOL", true); got != false {
			t.Errorf("parseBool(%q, true) = %v, want false", v, got)
		}
	}
	os.Unsetenv("_TEST_PARSE_BOOL")
}

func TestParseBool_InvalidReturnsDefault(t *testing.T) {
	os.Setenv("_TEST_PARSE_BOOL", "yes")
	defer os.Unsetenv("_TEST_PARSE_BOOL")
	if got := parseBool("_TEST_PARSE_BOOL", false); got != false {
		t.Errorf("parseBool(\"yes\", false) = %v, want false (invalid)", got)
	}
	os.Setenv("_TEST_PARSE_BOOL", "invalid")
	if got := parseBool("_TEST_PARSE_BOOL", true); got != true {
		t.Errorf("parseBool(\"invalid\", true) = %v, want true (default)", got)
	}
}

func TestParseInt_EmptyReturnsDefault(t *testing.T) {
	os.Unsetenv("_TEST_PARSE_INT")
	if got := parseInt("_TEST_PARSE_INT", 42); got != 42 {
		t.Errorf("parseInt(_, 42) = %v, want 42", got)
	}
	os.Unsetenv("_TEST_PARSE_INT")
}

func TestParseInt_ValidReturnsParsed(t *testing.T) {
	os.Setenv("_TEST_PARSE_INT", "24")
	defer os.Unsetenv("_TEST_PARSE_INT")
	if got := parseInt("_TEST_PARSE_INT", 1); got != 24 {
		t.Errorf("parseInt = %v, want 24", got)
	}
}

func TestParseInt_InvalidReturnsDefault(t *testing.T) {
	os.Setenv("_TEST_PARSE_INT", "not a number")
	defer os.Unsetenv("_TEST_PARSE_INT")
	if got := parseInt("_TEST_PARSE_INT", 7); got != 7 {
		t.Errorf("parseInt(invalid, 7) = %v, want 7", got)
	}
}

func TestBuildConfigFromEnv_Defaults(t *testing.T) {
	// Clear env for keys that have non-empty defaults so we can assert defaults
	keys := []string{
		"SERVER_PORT", "SERVER_MODE", "DATABASE_DRIVER", "DATABASE_HOST", "DATABASE_PORT",
		"REDIS_HOST", "REDIS_PORT", "REDIS_ENABLED", "REDIS_DB",
		"CORS_GLOBAL", "GCS_ENABLED",
	}
	for _, k := range keys {
		os.Unsetenv(k)
	}
	cfg := buildConfigFromEnv()
	if cfg == nil {
		t.Fatal("buildConfigFromEnv must not return nil")
	}
	if cfg.Server.Port != "8080" {
		t.Errorf("Server.Port = %q, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Mode != "debug" {
		t.Errorf("Server.Mode = %q, want debug", cfg.Server.Mode)
	}
	if cfg.Database.Driver != "mysql" {
		t.Errorf("Database.Driver = %q, want mysql", cfg.Database.Driver)
	}
	if cfg.Database.Host != "127.0.0.1" {
		t.Errorf("Database.Host = %q, want 127.0.0.1", cfg.Database.Host)
	}
	if cfg.Database.Port != "3306" {
		t.Errorf("Database.Port = %q, want 3306", cfg.Database.Port)
	}
	if cfg.Redis.Host != "127.0.0.1" {
		t.Errorf("Redis.Host = %q, want 127.0.0.1", cfg.Redis.Host)
	}
	if cfg.Redis.Port != "6379" {
		t.Errorf("Redis.Port = %q, want 6379", cfg.Redis.Port)
	}
	if cfg.Redis.Enabled != false {
		t.Errorf("Redis.Enabled = %v, want false", cfg.Redis.Enabled)
	}
	if cfg.Redis.DB != 1 {
		t.Errorf("Redis.DB = %v, want 1", cfg.Redis.DB)
	}
}

func TestBuildConfigFromEnv_RespectsEnv(t *testing.T) {
	os.Setenv("SERVER_PORT", "9999")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("REDIS_ENABLED", "true")
	os.Setenv("SERVER_ACCESS_TOKEN_EXPIRY", "24")
	os.Setenv("REDIS_DB", "3")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("DATABASE_DRIVER")
		os.Unsetenv("REDIS_ENABLED")
		os.Unsetenv("SERVER_ACCESS_TOKEN_EXPIRY")
		os.Unsetenv("REDIS_DB")
	}()

	cfg := buildConfigFromEnv()
	if cfg.Server.Port != "9999" {
		t.Errorf("Server.Port = %q, want 9999", cfg.Server.Port)
	}
	if cfg.Database.Driver != "postgres" {
		t.Errorf("Database.Driver = %q, want postgres", cfg.Database.Driver)
	}
	if !cfg.Redis.Enabled {
		t.Error("Redis.Enabled want true")
	}
	if cfg.Server.AccessTokenExpiry != 24 {
		t.Errorf("Server.AccessTokenExpiry = %v, want 24", cfg.Server.AccessTokenExpiry)
	}
	if cfg.Redis.DB != 3 {
		t.Errorf("Redis.DB = %v, want 3", cfg.Redis.DB)
	}
}

func TestInvalidPlaceholders(t *testing.T) {
	for _, p := range []string{"your_database_name", "your_database_user", "your_database_password"} {
		if !invalidPlaceholders[p] {
			t.Errorf("invalidPlaceholders[%q] = false, want true", p)
		}
	}
}
