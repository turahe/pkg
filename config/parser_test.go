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

func TestGetEnvOrDefault_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue string
		want         string
	}{
		{"empty string", "", "default", "default"},
		{"single char with double quotes", `"a"`, "default", "a"},
		{"single char with single quotes", "'a'", "default", "a"},
		{"only opening quote", `"value`, "default", `"value`},
		{"only closing quote", `value"`, "default", `value"`},
		{"quotes in middle", `val"ue`, "default", `val"ue`},
		{"empty quoted string", `""`, "default", ""},
		{"empty single quoted string", `''`, "default", ""},
		{"single character", "a", "default", "a"},
		{"value with spaces", "value with spaces", "default", "value with spaces"},
		{"quoted value with spaces", `"value with spaces"`, "default", "value with spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("_TEST_EDGE", tt.envValue)
				defer os.Unsetenv("_TEST_EDGE")
			} else {
				os.Unsetenv("_TEST_EDGE")
			}
			got := getEnvOrDefault("_TEST_EDGE", tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvOrDefault(%q, %q) = %q, want %q", tt.envValue, tt.defaultValue, got, tt.want)
			}
		})
	}
}

func TestParseInt_NegativeNumbers(t *testing.T) {
	os.Setenv("_TEST_NEG_INT", "-42")
	defer os.Unsetenv("_TEST_NEG_INT")
	if got := parseInt("_TEST_NEG_INT", 0); got != -42 {
		t.Errorf("parseInt(-42) = %v, want -42", got)
	}
}

func TestParseInt_Zero(t *testing.T) {
	os.Setenv("_TEST_ZERO_INT", "0")
	defer os.Unsetenv("_TEST_ZERO_INT")
	if got := parseInt("_TEST_ZERO_INT", 42); got != 0 {
		t.Errorf("parseInt(0) = %v, want 0", got)
	}
}

func TestParseInt_LargeNumber(t *testing.T) {
	os.Setenv("_TEST_LARGE_INT", "2147483647")
	defer os.Unsetenv("_TEST_LARGE_INT")
	if got := parseInt("_TEST_LARGE_INT", 0); got != 2147483647 {
		t.Errorf("parseInt(2147483647) = %v, want 2147483647", got)
	}
}

func TestBuildConfigFromEnv_AllSections(t *testing.T) {
	// Set comprehensive environment variables
	envVars := map[string]string{
		"SERVER_PORT":                  "3000",
		"SERVER_SECRET":                "test-secret",
		"SERVER_MODE":                  "release",
		"SERVER_ACCESS_TOKEN_EXPIRY":   "2",
		"SERVER_REFRESH_TOKEN_EXPIRY":  "14",
		"SERVER_TIMEZONE":              "Asia/Jakarta",
		"CORS_GLOBAL":                  "false",
		"CORS_IPS":                     "192.168.1.1",
		"DATABASE_DRIVER":              "postgres",
		"DATABASE_DBNAME":              "testdb",
		"DATABASE_USERNAME":            "testuser",
		"DATABASE_PASSWORD":            "testpass",
		"DATABASE_HOST":                "localhost",
		"DATABASE_PORT":                "5432",
		"DATABASE_SSLMODE":             "true",
		"DATABASE_LOGMODE":             "false",
		"DATABASE_CLOUD_SQL_INSTANCE":  "project:region:instance",
		"DATABASE_DRIVER_SITE":         "mysql",
		"DATABASE_DBNAME_SITE":         "sitedb",
		"DATABASE_USERNAME_SITE":       "siteuser",
		"DATABASE_PASSWORD_SITE":       "sitepass",
		"DATABASE_HOST_SITE":           "sitehost",
		"DATABASE_PORT_SITE":           "3307",
		"DATABASE_SSLMODE_SITE":        "true",
		"DATABASE_LOGMODE_SITE":        "false",
		"DATABASE_CLOUD_SQL_INSTANCE_SITE": "site:region:instance",
		"REDIS_ENABLED":                "true",
		"REDIS_HOST":                   "redis-host",
		"REDIS_PORT":                   "6380",
		"REDIS_PASSWORD":                "redis-pass",
		"REDIS_DB":                      "5",
		"GCS_ENABLED":                  "true",
		"GCS_BUCKET_NAME":              "test-bucket",
		"GCS_CREDENTIALS_FILE":          "/path/to/creds.json",
		"RATE_LIMITER_ENABLED":         "true",
		"RATE_LIMITER_REQUESTS":        "200",
		"RATE_LIMITER_WINDOW":          "120",
		"RATE_LIMITER_KEY_BY":          "user",
		"RATE_LIMITER_SKIP_PATHS":      "/health,/metrics",
	}

	// Set all environment variables
	for k, v := range envVars {
		os.Setenv(k, v)
	}
	defer func() {
		// Clean up
		for k := range envVars {
			os.Unsetenv(k)
		}
	}()

	cfg := buildConfigFromEnv()

	// Test Server configuration
	if cfg.Server.Port != "3000" {
		t.Errorf("Server.Port = %q, want 3000", cfg.Server.Port)
	}
	if cfg.Server.Secret != "test-secret" {
		t.Errorf("Server.Secret = %q, want test-secret", cfg.Server.Secret)
	}
	if cfg.Server.Mode != "release" {
		t.Errorf("Server.Mode = %q, want release", cfg.Server.Mode)
	}
	if cfg.Server.AccessTokenExpiry != 2 {
		t.Errorf("Server.AccessTokenExpiry = %v, want 2", cfg.Server.AccessTokenExpiry)
	}
	if cfg.Server.RefreshTokenExpiry != 14 {
		t.Errorf("Server.RefreshTokenExpiry = %v, want 14", cfg.Server.RefreshTokenExpiry)
	}

	// Test CORS configuration
	if cfg.Cors.Global != false {
		t.Errorf("Cors.Global = %v, want false", cfg.Cors.Global)
	}
	if cfg.Cors.Ips != "192.168.1.1" {
		t.Errorf("Cors.Ips = %q, want 192.168.1.1", cfg.Cors.Ips)
	}

	// Test Database configuration
	if cfg.Database.Driver != "postgres" {
		t.Errorf("Database.Driver = %q, want postgres", cfg.Database.Driver)
	}
	if cfg.Database.Dbname != "testdb" {
		t.Errorf("Database.Dbname = %q, want testdb", cfg.Database.Dbname)
	}
	if cfg.Database.Username != "testuser" {
		t.Errorf("Database.Username = %q, want testuser", cfg.Database.Username)
	}
	if cfg.Database.Password != "testpass" {
		t.Errorf("Database.Password = %q, want testpass", cfg.Database.Password)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %q, want localhost", cfg.Database.Host)
	}
	if cfg.Database.Port != "5432" {
		t.Errorf("Database.Port = %q, want 5432", cfg.Database.Port)
	}
	if cfg.Database.Sslmode != true {
		t.Errorf("Database.Sslmode = %v, want true", cfg.Database.Sslmode)
	}
	if cfg.Database.Logmode != false {
		t.Errorf("Database.Logmode = %v, want false", cfg.Database.Logmode)
	}
	if cfg.Database.CloudSQLInstance != "project:region:instance" {
		t.Errorf("Database.CloudSQLInstance = %q, want project:region:instance", cfg.Database.CloudSQLInstance)
	}

	// Test DatabaseSite configuration
	if cfg.DatabaseSite.Driver != "mysql" {
		t.Errorf("DatabaseSite.Driver = %q, want mysql", cfg.DatabaseSite.Driver)
	}
	if cfg.DatabaseSite.Dbname != "sitedb" {
		t.Errorf("DatabaseSite.Dbname = %q, want sitedb", cfg.DatabaseSite.Dbname)
	}
	if cfg.DatabaseSite.Username != "siteuser" {
		t.Errorf("DatabaseSite.Username = %q, want siteuser", cfg.DatabaseSite.Username)
	}
	if cfg.DatabaseSite.Password != "sitepass" {
		t.Errorf("DatabaseSite.Password = %q, want sitepass", cfg.DatabaseSite.Password)
	}
	if cfg.DatabaseSite.Host != "sitehost" {
		t.Errorf("DatabaseSite.Host = %q, want sitehost", cfg.DatabaseSite.Host)
	}
	if cfg.DatabaseSite.Port != "3307" {
		t.Errorf("DatabaseSite.Port = %q, want 3307", cfg.DatabaseSite.Port)
	}
	if cfg.DatabaseSite.Sslmode != true {
		t.Errorf("DatabaseSite.Sslmode = %v, want true", cfg.DatabaseSite.Sslmode)
	}
	if cfg.DatabaseSite.Logmode != false {
		t.Errorf("DatabaseSite.Logmode = %v, want false", cfg.DatabaseSite.Logmode)
	}
	if cfg.DatabaseSite.CloudSQLInstance != "site:region:instance" {
		t.Errorf("DatabaseSite.CloudSQLInstance = %q, want site:region:instance", cfg.DatabaseSite.CloudSQLInstance)
	}

	// Test Redis configuration
	if cfg.Redis.Enabled != true {
		t.Errorf("Redis.Enabled = %v, want true", cfg.Redis.Enabled)
	}
	if cfg.Redis.Host != "redis-host" {
		t.Errorf("Redis.Host = %q, want redis-host", cfg.Redis.Host)
	}
	if cfg.Redis.Port != "6380" {
		t.Errorf("Redis.Port = %q, want 6380", cfg.Redis.Port)
	}
	if cfg.Redis.Password != "redis-pass" {
		t.Errorf("Redis.Password = %q, want redis-pass", cfg.Redis.Password)
	}
	if cfg.Redis.DB != 5 {
		t.Errorf("Redis.DB = %v, want 5", cfg.Redis.DB)
	}

	// Test GCS configuration
	if cfg.GCS.Enabled != true {
		t.Errorf("GCS.Enabled = %v, want true", cfg.GCS.Enabled)
	}
	if cfg.GCS.BucketName != "test-bucket" {
		t.Errorf("GCS.BucketName = %q, want test-bucket", cfg.GCS.BucketName)
	}
	if cfg.GCS.CredentialsFile != "/path/to/creds.json" {
		t.Errorf("GCS.CredentialsFile = %q, want /path/to/creds.json", cfg.GCS.CredentialsFile)
	}

	// Test RateLimiter configuration
	if cfg.RateLimiter.Enabled != true {
		t.Errorf("RateLimiter.Enabled = %v, want true", cfg.RateLimiter.Enabled)
	}
	if cfg.RateLimiter.Requests != 200 {
		t.Errorf("RateLimiter.Requests = %v, want 200", cfg.RateLimiter.Requests)
	}
	if cfg.RateLimiter.Window != 120 {
		t.Errorf("RateLimiter.Window = %v, want 120", cfg.RateLimiter.Window)
	}
	if cfg.RateLimiter.KeyBy != "user" {
		t.Errorf("RateLimiter.KeyBy = %q, want user", cfg.RateLimiter.KeyBy)
	}
	if cfg.RateLimiter.SkipPaths != "/health,/metrics" {
		t.Errorf("RateLimiter.SkipPaths = %q, want /health,/metrics", cfg.RateLimiter.SkipPaths)
	}

	// Test Timezone configuration
	if cfg.Timezone.Timezone != "Asia/Jakarta" {
		t.Errorf("Timezone.Timezone = %q, want Asia/Jakarta", cfg.Timezone.Timezone)
	}
}
