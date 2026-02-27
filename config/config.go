package config

import (
	"fmt"

	"github.com/joho/godotenv"

	"github.com/turahe/pkg/logger"
)

// Config holds the global configuration instance. It is set by Setup or lazily
// built from environment variables by GetConfig. Tests may set it directly.
var Config *Configuration

// Setup loads configuration from the environment and validates required database settings.
//
// If configPath is non-empty, godotenv loads that file (e.g. ".env"); otherwise
// godotenv.Load() is used (current directory .env). After loading, buildConfigFromEnv
// reads all supported variables and validateDatabaseConfig is run for the primary
// database and, if DatabaseSite.Dbname is set, for the site database.
//
// Setup returns an error if database validation fails (missing Dbname, Username,
// Password, or CloudSQLInstance when driver is cloudsql-mysql/cloudsql-postgres).
func Setup(configPath string) error {
	if configPath != "" {
		_ = godotenv.Load(configPath)
	} else {
		_ = godotenv.Load()
	}

	logger.Infof("Config loaded from environment variables")

	cfg := buildConfigFromEnv()

	// Validate required database configuration
	if err := validateDatabaseConfig(&cfg.Database); err != nil {
		return fmt.Errorf("database configuration validation failed: %w", err)
	}

	// Validate site database configuration only if it's configured (optional)
	if cfg.DatabaseSite.Dbname != "" {
		if err := validateDatabaseConfig(&cfg.DatabaseSite); err != nil {
			return fmt.Errorf("site database configuration validation failed: %w", err)
		}
	}

	Config = cfg
	return nil
}

// GetConfig returns the global configuration. If Config is nil, it builds configuration
// from the current environment via buildConfigFromEnv and assigns it to Config.
// No validation is performed on this path; call Setup at startup for validation.
func GetConfig() *Configuration {
	if Config == nil {
		Config = buildConfigFromEnv()
	}
	return Config
}
