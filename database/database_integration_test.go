package database

import (
	"testing"

	"github.com/turahe/pkg/config"
)

func TestCreateDatabaseConnection_MySQL_Integration(t *testing.T) {
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
