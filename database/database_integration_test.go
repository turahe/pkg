package database

import (
	"testing"

	"github.com/turahe/pkg/config"
)

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
