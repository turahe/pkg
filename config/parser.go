package config

import (
	"os"
	"strconv"
)

// Placeholder values that should be treated as invalid
var invalidPlaceholders = map[string]bool{
	"your_database_name":     true,
	"your_database_user":     true,
	"your_database_password": true,
}

// buildConfigFromEnv builds configuration directly from environment variables
func buildConfigFromEnv() *Configuration {
	return &Configuration{
		Server: ServerConfiguration{
			Port:               getEnvOrDefault("SERVER_PORT", "8080"),
			Secret:             getEnvOrDefault("SERVER_SECRET", ""),
			Mode:               getEnvOrDefault("SERVER_MODE", "debug"),
			AccessTokenExpiry:  parseInt("SERVER_ACCESS_TOKEN_EXPIRY", 1),
			RefreshTokenExpiry: parseInt("SERVER_REFRESH_TOKEN_EXPIRY", 7),
		},
		Cors: CorsConfiguration{
			Global: parseBool("CORS_GLOBAL", true),
			Ips:    getEnvOrDefault("CORS_IPS", ""),
		},
		Database: DatabaseConfiguration{
			Driver:           getEnvOrDefault("DATABASE_DRIVER", "mysql"),
			Dbname:           getEnvOrDefault("DATABASE_DBNAME", ""),
			Username:         getEnvOrDefault("DATABASE_USERNAME", ""),
			Password:         getEnvOrDefault("DATABASE_PASSWORD", ""),
			Host:             getEnvOrDefault("DATABASE_HOST", "127.0.0.1"),
			Port:             getEnvOrDefault("DATABASE_PORT", "3306"),
			Sslmode:          parseBool("DATABASE_SSLMODE", false),
			Logmode:          parseBool("DATABASE_LOGMODE", true),
			CloudSQLInstance: getEnvOrDefault("DATABASE_CLOUD_SQL_INSTANCE", ""),
		},
		DatabaseSite: DatabaseConfiguration{
			Driver:           getEnvOrDefault("DATABASE_DRIVER_SITE", "mysql"),
			Dbname:           getEnvOrDefault("DATABASE_DBNAME_SITE", ""),
			Username:         getEnvOrDefault("DATABASE_USERNAME_SITE", ""),
			Password:         getEnvOrDefault("DATABASE_PASSWORD_SITE", ""),
			Host:             getEnvOrDefault("DATABASE_HOST_SITE", "127.0.0.1"),
			Port:             getEnvOrDefault("DATABASE_PORT_SITE", "3306"),
			Sslmode:          parseBool("DATABASE_SSLMODE_SITE", false),
			Logmode:          parseBool("DATABASE_LOGMODE_SITE", true),
			CloudSQLInstance: getEnvOrDefault("DATABASE_CLOUD_SQL_INSTANCE_SITE", ""),
		},
		Redis: RedisConfiguration{
			Enabled:  parseBool("REDIS_ENABLED", false),
			Host:     getEnvOrDefault("REDIS_HOST", "127.0.0.1"),
			Port:     getEnvOrDefault("REDIS_PORT", "6379"),
			Password: getEnvOrDefault("REDIS_PASSWORD", ""),
			DB:       parseInt("REDIS_DB", 1),
		},
		GCS: GCSConfiguration{
			Enabled:         parseBool("GCS_ENABLED", false),
			BucketName:      getEnvOrDefault("GCS_BUCKET_NAME", ""),
			CredentialsFile: getEnvOrDefault("GCS_CREDENTIALS_FILE", ""),
		},
	}
}

// getEnvOrDefault gets environment variable or returns default value.
// Also strips surrounding quotes if present.
func getEnvOrDefault(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	// Strip surrounding quotes if present
	if len(val) >= 2 {
		if (val[0] == '"' && val[len(val)-1] == '"') ||
			(val[0] == '\'' && val[len(val)-1] == '\'') {
			return val[1 : len(val)-1]
		}
	}
	return val
}

// parseBool parses a boolean from environment variable
func parseBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	result, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue
	}
	return result
}

// parseInt parses an integer from environment variable
func parseInt(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	result, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return result
}
