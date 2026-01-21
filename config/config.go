package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

// Config holds the global configuration instance
var Config *Configuration

// Setup initializes the configuration by loading environment variables
// and validating required settings.
func Setup(configPath string) error {
	// Try to load .env file (ignore error if it doesn't exist)
	_ = godotenv.Load()

	log.Println("Config loaded from environment variables")

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

// GetConfig returns the global configuration instance
// If no config is set, it will attempt to build from environment variables
func GetConfig() *Configuration {
	if Config == nil {
		// Try to build from environment variables as fallback
		Config = buildConfigFromEnv()
	}
	return Config
}
