package database

import (
	"testing"

	"github.com/turahe/pkg/config"
)

func TestCreateDatabaseConnection_InvalidDriver(t *testing.T) {
	cfg := &config.DatabaseConfiguration{
		Driver:   "invalid-driver",
		Host:     "localhost",
		Port:     "3306",
		Username: "u",
		Password: "p",
		Dbname:   "db",
	}
	_, err := CreateDatabaseConnection(cfg)
	if err == nil {
		t.Error("CreateDatabaseConnection must fail for invalid driver")
	}
}

func TestCreateDatabaseConnection_SQLite(t *testing.T) {
	cfg := &config.DatabaseConfiguration{
		Driver:  "sqlite",
		Dbname:  "test_db_sqlite",
		Logmode: false,
	}
	db, err := CreateDatabaseConnection(cfg)
	if err != nil {
		t.Skipf("CreateDatabaseConnection sqlite: %v (sqlite may not be available)", err)
	}
	if db == nil {
		t.Fatal("db must not be nil")
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}
	_ = sqlDB.Close()
}
