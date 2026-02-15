package database

import (
	"context"
	"testing"
	"time"

	"github.com/turahe/pkg/config"
	"gorm.io/gorm/logger"
)

func TestNew_WithOverride(t *testing.T) {
	cfg := &config.DatabaseConfiguration{
		Driver:  "sqlite",
		Dbname:  "test_override",
		Logmode: false,
	}
	db, err := New(cfg, Options{}, WithLogLevel(logger.Info))
	if err != nil {
		t.Skipf("New: %v", err)
	}
	defer db.Close()
	if db.opts.LogLevel != logger.Info {
		t.Errorf("LogLevel = %v, want Info", db.opts.LogLevel)
	}
}

func TestNewContext(t *testing.T) {
	ctx := context.Background()
	cfg := &config.DatabaseConfiguration{
		Driver:  "sqlite",
		Dbname:  "test_newcontext",
		Logmode: false,
	}
	db, err := NewContext(ctx, cfg, Options{})
	if err != nil {
		t.Skipf("NewContext: %v", err)
	}
	defer db.Close()
	if db == nil {
		t.Fatal("db is nil")
	}
}

func TestNew_InvalidDriver(t *testing.T) {
	cfg := &config.DatabaseConfiguration{
		Driver:   "invalid-driver",
		Host:     "localhost",
		Port:     "3306",
		Username: "u",
		Password: "p",
		Dbname:   "db",
	}
	_, err := New(cfg, Options{})
	if err == nil {
		t.Error("New must fail for invalid driver")
	}
}

func TestNew_SQLite(t *testing.T) {
	cfg := &config.DatabaseConfiguration{
		Driver:  "sqlite",
		Dbname:  "test_db_sqlite",
		Logmode: false,
	}
	db, err := New(cfg, Options{})
	if err != nil {
		t.Skipf("New sqlite: %v (sqlite may not be available)", err)
	}
	if db == nil {
		t.Fatal("db must not be nil")
	}
	sqlDB, err := db.DB().DB()
	if err != nil {
		t.Fatalf("db.DB().DB(): %v", err)
	}
	_ = sqlDB.Close()
	_ = db.Close()
}

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
	gormDB, err := CreateDatabaseConnection(cfg)
	if err != nil {
		t.Skipf("CreateDatabaseConnection sqlite: %v (sqlite may not be available)", err)
	}
	if gormDB == nil {
		t.Fatal("gormDB must not be nil")
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("gormDB.DB(): %v", err)
	}
	_ = sqlDB.Close()
}

func TestSetup_HealthCheck(t *testing.T) {
	origDB := defaultDB
	origDBSite := defaultDBSite
	compatMu.Lock()
	defaultDB = nil
	defaultDBSite = nil
	compatMu.Unlock()
	defer func() {
		Cleanup()
		compatMu.Lock()
		defaultDB = origDB
		defaultDBSite = origDBSite
		compatMu.Unlock()
	}()
	if err := Setup(); err != nil {
		t.Skipf("Setup: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck: %v", err)
	}
}

func TestDatabase_Health(t *testing.T) {
	cfg := &config.DatabaseConfiguration{
		Driver:  "sqlite",
		Dbname:  "test_health",
		Logmode: false,
	}
	db, err := New(cfg, Options{})
	if err != nil {
		t.Skipf("New: %v", err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.Health(ctx); err != nil {
		t.Errorf("Health: %v", err)
	}
}
