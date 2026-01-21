package config

import (
	"fmt"
	"strings"
)

// validateDatabaseConfig validates that required database configuration is present
func validateDatabaseConfig(dbConfig *DatabaseConfiguration) error {
	var missingFields []string

	// Check if using Cloud SQL
	isCloudSQL := dbConfig.Driver == "cloudsql-mysql" || dbConfig.Driver == "cloudsql-postgres"
	if isCloudSQL && dbConfig.CloudSQLInstance == "" {
		missingFields = append(missingFields, "DATABASE_CLOUD_SQL_INSTANCE")
	}

	// Validate required fields for all database types
	if isEmptyOrPlaceholder(dbConfig.Dbname) {
		missingFields = append(missingFields, "DATABASE_DBNAME")
	}
	if isEmptyOrPlaceholder(dbConfig.Username) {
		missingFields = append(missingFields, "DATABASE_USERNAME")
	}
	if isEmptyOrPlaceholder(dbConfig.Password) {
		missingFields = append(missingFields, "DATABASE_PASSWORD")
	}

	if len(missingFields) > 0 {
		missingList := strings.Join(missingFields, "\n  - ")
		return fmt.Errorf("missing or invalid database configuration. Please set the following environment variables in your .env file:\n  - %s\n\nTo get started, copy env.example to .env and update with your actual database credentials:\n  Copy-Item env.example .env", missingList)
	}

	return nil
}

// isEmptyOrPlaceholder checks if a value is empty or a known placeholder
func isEmptyOrPlaceholder(value string) bool {
	return value == "" || invalidPlaceholders[value]
}

