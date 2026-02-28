package config

import (
	"fmt"
	"strings"
)

// validateDatabaseConfig checks that required database fields are set and not placeholders.
// For Cloud SQL drivers it also requires CloudSQLInstance. Returns an error listing missing env vars.
// If forSite is true, missing field names use the _SITE suffix (e.g. DATABASE_PASSWORD_SITE).
func validateDatabaseConfig(dbConfig *DatabaseConfiguration, forSite bool) error {
	var missingFields []string
	suf := ""
	if forSite {
		suf = "_SITE"
	}

	// Check if using Cloud SQL
	isCloudSQL := dbConfig.Driver == "cloudsql-mysql" || dbConfig.Driver == "cloudsql-postgres"
	if isCloudSQL && dbConfig.CloudSQLInstance == "" {
		missingFields = append(missingFields, "DATABASE_CLOUD_SQL_INSTANCE"+suf)
	}

	// Validate required fields for all database types
	if isEmptyOrPlaceholder(dbConfig.Dbname) {
		missingFields = append(missingFields, "DATABASE_DBNAME"+suf)
	}
	if isEmptyOrPlaceholder(dbConfig.Username) {
		missingFields = append(missingFields, "DATABASE_USERNAME"+suf)
	}
	if isEmptyOrPlaceholder(dbConfig.Password) {
		missingFields = append(missingFields, "DATABASE_PASSWORD"+suf)
	}

	if len(missingFields) > 0 {
		missingList := strings.Join(missingFields, "\n  - ")
		return fmt.Errorf("missing or invalid database configuration. Please set the following environment variables in your .env file:\n  - %s\n\nTo get started, copy env.example to .env and update with your actual database credentials:\n  Copy-Item env.example .env", missingList)
	}

	return nil
}

// isEmptyOrPlaceholder returns true if value is empty or matches a known placeholder (e.g. "your_database_password").
func isEmptyOrPlaceholder(value string) bool {
	return value == "" || invalidPlaceholders[value]
}

