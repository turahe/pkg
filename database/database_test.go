package database

import (
	"context"
	"testing"
	"time"

	"github.com/turahe/pkg/config"
	"gorm.io/gorm/logger"
)

func TestNew_WithOverride(t *testing.T) {
	stubConnectStandardWithMock(t)
	cfg := &config.DatabaseConfiguration{
		Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db",
	}
	db, err := New(cfg, Options{}, WithLogLevel(logger.Info))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()
	if db.opts.LogLevel != logger.Info {
		t.Errorf("LogLevel = %v, want Info", db.opts.LogLevel)
	}
}

func TestNewContext(t *testing.T) {
	stubConnectStandardWithMock(t)
	cfg := &config.DatabaseConfiguration{
		Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db",
	}
	db, err := NewContext(context.Background(), cfg, Options{})
	if err != nil {
		t.Fatalf("NewContext: %v", err)
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

func TestNew_MySQLViaStub(t *testing.T) {
	stubConnectStandardWithMock(t)
	cfg := &config.DatabaseConfiguration{
		Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db",
	}
	db, err := New(cfg, Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()
	if db.DB() == nil {
		t.Fatal("db.DB() nil")
	}
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

func TestCreateDatabaseConnection_Success(t *testing.T) {
	stubConnectStandardWithMock(t)
	cfg := &config.DatabaseConfiguration{
		Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db",
	}
	gormDB, err := CreateDatabaseConnection(cfg)
	if err != nil {
		t.Fatalf("CreateDatabaseConnection: %v", err)
	}
	if gormDB == nil {
		t.Fatal("gormDB must not be nil")
	}
	_ = Cleanup()
}

func TestSetup_HealthCheck(t *testing.T) {
	stubConnectStandardWithMock(t)
	orig := config.Config
	defer func() {
		_ = Cleanup()
		config.Config = orig
	}()
	config.Config = &config.Configuration{
		Database: config.DatabaseConfiguration{
			Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db",
		},
	}
	if err := Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck: %v", err)
	}
}

func TestDatabase_Health(t *testing.T) {
	db, _ := newMockDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.Health(ctx); err != nil {
		t.Errorf("Health: %v", err)
	}
}
