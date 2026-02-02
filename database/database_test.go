package database

import (
	"context"
	"testing"
	"time"

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

func TestIsAlive(t *testing.T) {
	// When DB is nil, IsAlive returns false
	origDB := DB
	DB = nil
	defer func() { DB = origDB }()
	if IsAlive() {
		t.Error("IsAlive() should be false when DB is nil")
	}

	// When DB is set and reachable, IsAlive returns true
	cfg := &config.DatabaseConfiguration{
		Driver:  "sqlite",
		Dbname:  "test_isalive",
		Logmode: false,
	}
	db, err := CreateDatabaseConnection(cfg)
	if err != nil {
		t.Skipf("CreateDatabaseConnection: %v", err)
	}
	DB = db
	defer func() { DB = origDB }()
	if !IsAlive() {
		t.Error("IsAlive() should be true when DB is connected")
	}
}

func TestHealthCheck(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// When DB is nil, HealthCheck returns error
	origDB := DB
	origDBSite := DBSite
	DB = nil
	DBSite = nil
	defer func() { DB = origDB; DBSite = origDBSite }()
	if err := HealthCheck(ctx); err == nil {
		t.Error("HealthCheck() should return error when DB is nil")
	}

	// When DB is set and reachable, HealthCheck returns nil
	cfg := &config.DatabaseConfiguration{
		Driver:  "sqlite",
		Dbname:  "test_healthcheck",
		Logmode: false,
	}
	db, err := CreateDatabaseConnection(cfg)
	if err != nil {
		t.Skipf("CreateDatabaseConnection: %v", err)
	}
	DB = db
	defer func() { DB = origDB }()
	if err := HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck() should succeed when DB is connected: %v", err)
	}
}
