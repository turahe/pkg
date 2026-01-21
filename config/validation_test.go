package config

import (
	"os"
	"strings"
	"testing"
)

func TestIsEmptyOrPlaceholder_Empty(t *testing.T) {
	if !isEmptyOrPlaceholder("") {
		t.Error("isEmptyOrPlaceholder(\"\") = false, want true")
	}
}

func TestIsEmptyOrPlaceholder_Placeholders(t *testing.T) {
	for _, p := range []string{"your_database_name", "your_database_user", "your_database_password"} {
		if !isEmptyOrPlaceholder(p) {
			t.Errorf("isEmptyOrPlaceholder(%q) = false, want true", p)
		}
	}
}

func TestIsEmptyOrPlaceholder_Valid(t *testing.T) {
	for _, v := range []string{"mydb", "user", "secret", "a"} {
		if isEmptyOrPlaceholder(v) {
			t.Errorf("isEmptyOrPlaceholder(%q) = true, want false", v)
		}
	}
}

func TestValidateDatabaseConfig_MissingDbname(t *testing.T) {
	os.Unsetenv("DATABASE_DBNAME")
	os.Setenv("DATABASE_USERNAME", "u")
	os.Setenv("DATABASE_PASSWORD", "p")
	os.Setenv("DATABASE_DRIVER", "mysql")
	defer func() {
		os.Unsetenv("DATABASE_USERNAME")
		os.Unsetenv("DATABASE_PASSWORD")
		os.Unsetenv("DATABASE_DRIVER")
	}()

	err := Setup("")
	if err == nil {
		t.Fatal("Setup should fail when DATABASE_DBNAME is empty")
	}
	if !strings.Contains(err.Error(), "DATABASE_DBNAME") {
		t.Errorf("error should mention DATABASE_DBNAME, got: %v", err)
	}
}

func TestValidateDatabaseConfig_PlaceholderDbname(t *testing.T) {
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
		t.Fatal("Setup should fail when DATABASE_DBNAME is placeholder")
	}
	if !strings.Contains(err.Error(), "DATABASE_DBNAME") {
		t.Errorf("error should mention DATABASE_DBNAME, got: %v", err)
	}
}

func TestValidateDatabaseConfig_PlaceholderUsername(t *testing.T) {
	os.Setenv("DATABASE_DBNAME", "db")
	os.Setenv("DATABASE_USERNAME", "your_database_user")
	os.Setenv("DATABASE_PASSWORD", "p")
	defer func() {
		os.Unsetenv("DATABASE_DBNAME")
		os.Unsetenv("DATABASE_USERNAME")
		os.Unsetenv("DATABASE_PASSWORD")
	}()

	err := Setup("")
	if err == nil {
		t.Fatal("Setup should fail when DATABASE_USERNAME is placeholder")
	}
	if !strings.Contains(err.Error(), "DATABASE_USERNAME") {
		t.Errorf("error should mention DATABASE_USERNAME, got: %v", err)
	}
}

func TestValidateDatabaseConfig_PlaceholderPassword(t *testing.T) {
	os.Setenv("DATABASE_DBNAME", "db")
	os.Setenv("DATABASE_USERNAME", "u")
	os.Setenv("DATABASE_PASSWORD", "your_database_password")
	defer func() {
		os.Unsetenv("DATABASE_DBNAME")
		os.Unsetenv("DATABASE_USERNAME")
		os.Unsetenv("DATABASE_PASSWORD")
	}()

	err := Setup("")
	if err == nil {
		t.Fatal("Setup should fail when DATABASE_PASSWORD is placeholder")
	}
	if !strings.Contains(err.Error(), "DATABASE_PASSWORD") {
		t.Errorf("error should mention DATABASE_PASSWORD, got: %v", err)
	}
}

func TestValidateDatabaseConfig_CloudSQL_MissingInstance(t *testing.T) {
	os.Setenv("DATABASE_DRIVER", "cloudsql-mysql")
	os.Setenv("DATABASE_DBNAME", "db")
	os.Setenv("DATABASE_USERNAME", "u")
	os.Setenv("DATABASE_PASSWORD", "p")
	os.Unsetenv("DATABASE_CLOUD_SQL_INSTANCE")
	defer func() {
		os.Unsetenv("DATABASE_DRIVER")
		os.Unsetenv("DATABASE_DBNAME")
		os.Unsetenv("DATABASE_USERNAME")
		os.Unsetenv("DATABASE_PASSWORD")
	}()

	err := Setup("")
	if err == nil {
		t.Fatal("Setup should fail when cloudsql-mysql has empty DATABASE_CLOUD_SQL_INSTANCE")
	}
	if !strings.Contains(err.Error(), "DATABASE_CLOUD_SQL_INSTANCE") {
		t.Errorf("error should mention DATABASE_CLOUD_SQL_INSTANCE, got: %v", err)
	}
}

func TestValidateDatabaseConfig_CloudSQLPostgres_MissingInstance(t *testing.T) {
	os.Setenv("DATABASE_DRIVER", "cloudsql-postgres")
	os.Setenv("DATABASE_DBNAME", "db")
	os.Setenv("DATABASE_USERNAME", "u")
	os.Setenv("DATABASE_PASSWORD", "p")
	os.Unsetenv("DATABASE_CLOUD_SQL_INSTANCE")
	defer func() {
		os.Unsetenv("DATABASE_DRIVER")
		os.Unsetenv("DATABASE_DBNAME")
		os.Unsetenv("DATABASE_USERNAME")
		os.Unsetenv("DATABASE_PASSWORD")
	}()

	err := Setup("")
	if err == nil {
		t.Fatal("Setup should fail when cloudsql-postgres has empty DATABASE_CLOUD_SQL_INSTANCE")
	}
	if !strings.Contains(err.Error(), "DATABASE_CLOUD_SQL_INSTANCE") {
		t.Errorf("error should mention DATABASE_CLOUD_SQL_INSTANCE, got: %v", err)
	}
}

func TestValidateDatabaseConfig_CloudSQL_Valid(t *testing.T) {
	os.Setenv("DATABASE_DRIVER", "cloudsql-mysql")
	os.Setenv("DATABASE_CLOUD_SQL_INSTANCE", "proj:region:inst")
	os.Setenv("DATABASE_DBNAME", "db")
	os.Setenv("DATABASE_USERNAME", "u")
	os.Setenv("DATABASE_PASSWORD", "p")
	defer func() {
		os.Unsetenv("DATABASE_DRIVER")
		os.Unsetenv("DATABASE_CLOUD_SQL_INSTANCE")
		os.Unsetenv("DATABASE_DBNAME")
		os.Unsetenv("DATABASE_USERNAME")
		os.Unsetenv("DATABASE_PASSWORD")
	}()

	err := Setup("")
	if err != nil {
		t.Errorf("Setup with valid Cloud SQL config should succeed: %v", err)
	}
}

func TestValidateDatabaseConfig_Valid(t *testing.T) {
	os.Setenv("DATABASE_DBNAME", "mydb")
	os.Setenv("DATABASE_USERNAME", "myuser")
	os.Setenv("DATABASE_PASSWORD", "mypass")
	defer func() {
		os.Unsetenv("DATABASE_DBNAME")
		os.Unsetenv("DATABASE_USERNAME")
		os.Unsetenv("DATABASE_PASSWORD")
	}()

	err := Setup("")
	if err != nil {
		t.Errorf("Setup with valid DB config should succeed: %v", err)
	}
	if Config == nil {
		t.Error("Config should be set after successful Setup")
	}
}
