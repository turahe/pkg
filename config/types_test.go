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
