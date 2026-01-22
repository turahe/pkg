package config

import (
	"testing"
)

func TestConfiguration_StructFields(t *testing.T) {
	cfg := &Configuration{
		Server: ServerConfiguration{
			Port:               "8080",
			Secret:             "s",
			Mode:               "release",
			AccessTokenExpiry:  1,
			RefreshTokenExpiry: 7,
		},
		Cors: CorsConfiguration{
			Global: true,
			Ips:    "127.0.0.1",
		},
		Database: DatabaseConfiguration{
			Driver:           "mysql",
			Dbname:           "db",
			Username:         "u",
			Password:         "p",
			Host:             "localhost",
			Port:             "3306",
			Sslmode:          false,
			Logmode:          true,
			CloudSQLInstance: "p:r:i",
		},
		DatabaseSite: DatabaseConfiguration{},
		Redis: RedisConfiguration{
			Enabled:  true,
			Host:     "127.0.0.1",
			Port:     "6379",
			Password: "",
			DB:       0,
		},
		GCS: GCSConfiguration{
			Enabled:         false,
			BucketName:      "",
			CredentialsFile: "",
		},
	}

	if cfg.Server.Port != "8080" || cfg.Server.AccessTokenExpiry != 1 {
		t.Error("ServerConfiguration fields not set correctly")
	}
	if cfg.Database.Driver != "mysql" || cfg.Database.CloudSQLInstance != "p:r:i" {
		t.Error("DatabaseConfiguration fields not set correctly")
	}
	if !cfg.Redis.Enabled || cfg.Redis.DB != 0 {
		t.Error("RedisConfiguration fields not set correctly")
	}
}

func TestServerConfiguration_AllFields(t *testing.T) {
	server := ServerConfiguration{
		Port:               "8080",
		Secret:             "secret-key",
		Mode:               "production",
		AccessTokenExpiry:  24,
		RefreshTokenExpiry: 30,
	}

	if server.Port != "8080" {
		t.Errorf("Port = %q, want 8080", server.Port)
	}
	if server.Secret != "secret-key" {
		t.Errorf("Secret = %q, want secret-key", server.Secret)
	}
	if server.Mode != "production" {
		t.Errorf("Mode = %q, want production", server.Mode)
	}
	if server.AccessTokenExpiry != 24 {
		t.Errorf("AccessTokenExpiry = %v, want 24", server.AccessTokenExpiry)
	}
	if server.RefreshTokenExpiry != 30 {
		t.Errorf("RefreshTokenExpiry = %v, want 30", server.RefreshTokenExpiry)
	}
}

func TestCorsConfiguration_AllFields(t *testing.T) {
	cors := CorsConfiguration{
		Global: false,
		Ips:    "192.168.1.1,192.168.1.2",
	}

	if cors.Global != false {
		t.Errorf("Global = %v, want false", cors.Global)
	}
	if cors.Ips != "192.168.1.1,192.168.1.2" {
		t.Errorf("Ips = %q, want 192.168.1.1,192.168.1.2", cors.Ips)
	}
}

func TestDatabaseConfiguration_AllFields(t *testing.T) {
	db := DatabaseConfiguration{
		Driver:           "postgres",
		Dbname:           "mydb",
		Username:         "admin",
		Password:         "password123",
		Host:             "db.example.com",
		Port:             "5432",
		Sslmode:          true,
		Logmode:          false,
		CloudSQLInstance: "myproject:us-central1:mydb",
	}

	if db.Driver != "postgres" {
		t.Errorf("Driver = %q, want postgres", db.Driver)
	}
	if db.Dbname != "mydb" {
		t.Errorf("Dbname = %q, want mydb", db.Dbname)
	}
	if db.Username != "admin" {
		t.Errorf("Username = %q, want admin", db.Username)
	}
	if db.Password != "password123" {
		t.Errorf("Password = %q, want password123", db.Password)
	}
	if db.Host != "db.example.com" {
		t.Errorf("Host = %q, want db.example.com", db.Host)
	}
	if db.Port != "5432" {
		t.Errorf("Port = %q, want 5432", db.Port)
	}
	if db.Sslmode != true {
		t.Errorf("Sslmode = %v, want true", db.Sslmode)
	}
	if db.Logmode != false {
		t.Errorf("Logmode = %v, want false", db.Logmode)
	}
	if db.CloudSQLInstance != "myproject:us-central1:mydb" {
		t.Errorf("CloudSQLInstance = %q, want myproject:us-central1:mydb", db.CloudSQLInstance)
	}
}

func TestRedisConfiguration_AllFields(t *testing.T) {
	redis := RedisConfiguration{
		Enabled:  true,
		Host:     "redis.example.com",
		Port:     "6380",
		Password: "redis-password",
		DB:       3,
	}

	if redis.Enabled != true {
		t.Errorf("Enabled = %v, want true", redis.Enabled)
	}
	if redis.Host != "redis.example.com" {
		t.Errorf("Host = %q, want redis.example.com", redis.Host)
	}
	if redis.Port != "6380" {
		t.Errorf("Port = %q, want 6380", redis.Port)
	}
	if redis.Password != "redis-password" {
		t.Errorf("Password = %q, want redis-password", redis.Password)
	}
	if redis.DB != 3 {
		t.Errorf("DB = %v, want 3", redis.DB)
	}
}

func TestGCSConfiguration_AllFields(t *testing.T) {
	gcs := GCSConfiguration{
		Enabled:         true,
		BucketName:      "my-bucket",
		CredentialsFile: "/path/to/credentials.json",
	}

	if gcs.Enabled != true {
		t.Errorf("Enabled = %v, want true", gcs.Enabled)
	}
	if gcs.BucketName != "my-bucket" {
		t.Errorf("BucketName = %q, want my-bucket", gcs.BucketName)
	}
	if gcs.CredentialsFile != "/path/to/credentials.json" {
		t.Errorf("CredentialsFile = %q, want /path/to/credentials.json", gcs.CredentialsFile)
	}
}

func TestRateLimiterConfiguration_AllFields(t *testing.T) {
	rl := RateLimiterConfiguration{
		Enabled:   true,
		Requests:  100,
		Window:    60,
		KeyBy:     "ip",
		SkipPaths: "/health,/metrics,/status",
	}

	if rl.Enabled != true {
		t.Errorf("Enabled = %v, want true", rl.Enabled)
	}
	if rl.Requests != 100 {
		t.Errorf("Requests = %v, want 100", rl.Requests)
	}
	if rl.Window != 60 {
		t.Errorf("Window = %v, want 60", rl.Window)
	}
	if rl.KeyBy != "ip" {
		t.Errorf("KeyBy = %q, want ip", rl.KeyBy)
	}
	if rl.SkipPaths != "/health,/metrics,/status" {
		t.Errorf("SkipPaths = %q, want /health,/metrics,/status", rl.SkipPaths)
	}
}

func TestTimezoneConfiguration_AllFields(t *testing.T) {
	tz := TimezoneConfiguration{
		Timezone: "America/New_York",
	}

	if tz.Timezone != "America/New_York" {
		t.Errorf("Timezone = %q, want America/New_York", tz.Timezone)
	}
}

func TestConfiguration_ZeroValues(t *testing.T) {
	cfg := &Configuration{}

	// Test that zero values are valid
	if cfg.Server.Port != "" {
		t.Errorf("Server.Port zero value = %q, want empty string", cfg.Server.Port)
	}
	if cfg.Server.AccessTokenExpiry != 0 {
		t.Errorf("Server.AccessTokenExpiry zero value = %v, want 0", cfg.Server.AccessTokenExpiry)
	}
	if cfg.Cors.Global != false {
		t.Errorf("Cors.Global zero value = %v, want false", cfg.Cors.Global)
	}
	if cfg.Database.Driver != "" {
		t.Errorf("Database.Driver zero value = %q, want empty string", cfg.Database.Driver)
	}
	if cfg.Redis.Enabled != false {
		t.Errorf("Redis.Enabled zero value = %v, want false", cfg.Redis.Enabled)
	}
	if cfg.Redis.DB != 0 {
		t.Errorf("Redis.DB zero value = %v, want 0", cfg.Redis.DB)
	}
	if cfg.GCS.Enabled != false {
		t.Errorf("GCS.Enabled zero value = %v, want false", cfg.GCS.Enabled)
	}
	if cfg.RateLimiter.Enabled != false {
		t.Errorf("RateLimiter.Enabled zero value = %v, want false", cfg.RateLimiter.Enabled)
	}
	if cfg.Timezone.Timezone != "" {
		t.Errorf("Timezone.Timezone zero value = %q, want empty string", cfg.Timezone.Timezone)
	}
}
