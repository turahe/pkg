package database

import (
	"os"
	"testing"

	"github.com/turahe/pkg/config"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func sqlServerIntegrationConfig() *config.DatabaseConfiguration {
	return &config.DatabaseConfiguration{
		Driver:   "sqlserver",
		Host:     envOr("SQLSERVER_HOST", "127.0.0.1"),
		Port:     envOr("SQLSERVER_PORT", "1433"),
		Username: envOr("SQLSERVER_USERNAME", "sa"),
		Password: envOr("SQLSERVER_PASSWORD", "Test_Password123"),
		Dbname:   envOr("SQLSERVER_DBNAME", "master"),
		Logmode:  false,
	}
}

func TestNew_MySQL_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := &config.DatabaseConfiguration{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     "3306",
		Username: "root",
		Password: "root",
		Dbname:   "testdb",
		Logmode:  false,
	}
	db, err := New(cfg, Options{})
	if err != nil {
		t.Skipf("MySQL not available (start with docker compose): %v", err)
	}
	defer db.Close()
	if db == nil {
		t.Fatal("db must not be nil")
	}
	sqlDB, err := db.DB().DB()
	if err != nil {
		t.Fatalf("db.DB().DB(): %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestNew_Postgres_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := &config.DatabaseConfiguration{
		Driver:   "postgres",
		Host:     "127.0.0.1",
		Port:     "5432",
		Username: "test",
		Password: "test",
		Dbname:   "testdb",
		Logmode:  false,
	}
	db, err := New(cfg, Options{})
	if err != nil {
		t.Skipf("Postgres not available (start with docker compose): %v", err)
	}
	defer db.Close()
	if db == nil {
		t.Fatal("db must not be nil")
	}
	sqlDB, err := db.DB().DB()
	if err != nil {
		t.Fatalf("db.DB().DB(): %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestCreateDatabaseConnection_MySQL_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := &config.DatabaseConfiguration{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     "3306",
		Username: "root",
		Password: "root",
		Dbname:   "testdb",
		Logmode:  false,
	}
	db, err := CreateDatabaseConnection(cfg)
	if err != nil {
		t.Skipf("MySQL not available (start with docker compose): %v", err)
	}
	if db == nil {
		t.Fatal("db must not be nil")
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	_ = sqlDB.Close()
}

func TestCreateDatabaseConnection_Postgres_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := &config.DatabaseConfiguration{
		Driver:   "postgres",
		Host:     "127.0.0.1",
		Port:     "5432",
		Username: "test",
		Password: "test",
		Dbname:   "testdb",
		Logmode:  false,
	}
	db, err := CreateDatabaseConnection(cfg)
	if err != nil {
		t.Skipf("Postgres not available (start with docker compose): %v", err)
	}
	if db == nil {
		t.Fatal("db must not be nil")
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	_ = sqlDB.Close()
}

func TestNew_SQLServer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := sqlServerIntegrationConfig()
	db, err := New(cfg, Options{})
	if err != nil {
		t.Skipf("SQL Server not available (start with docker compose): %v", err)
	}
	defer db.Close()
	if db == nil {
		t.Fatal("db must not be nil")
	}
	sqlDB, err := db.DB().DB()
	if err != nil {
		t.Fatalf("db.DB().DB(): %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	if err := db.Health(t.Context()); err != nil {
		t.Fatalf("Health: %v", err)
	}
}

func TestCreateDatabaseConnection_SQLServer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := sqlServerIntegrationConfig()
	db, err := CreateDatabaseConnection(cfg)
	if err != nil {
		t.Skipf("SQL Server not available (start with docker compose): %v", err)
	}
	if db == nil {
		t.Fatal("db must not be nil")
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	_ = sqlDB.Close()
}
