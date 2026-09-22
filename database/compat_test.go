package database

import (
	"context"
	"testing"
	"time"

	"github.com/turahe/pkg/config"
)

func TestHealthCheck_NotInitialized(t *testing.T) {
	compatMu.Lock()
	origDB := defaultDB
	origDBSite := defaultDBSite
	defaultDB = nil
	defaultDBSite = nil
	compatMu.Unlock()
	defer func() {
		compatMu.Lock()
		defaultDB = origDB
		defaultDBSite = origDBSite
		compatMu.Unlock()
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := HealthCheck(ctx)
	if err != ErrNotInitialized {
		t.Errorf("HealthCheck() = %v, want ErrNotInitialized", err)
	}
}

func TestIsAlive_NotInitialized(t *testing.T) {
	compatMu.Lock()
	origDB := defaultDB
	defaultDB = nil
	compatMu.Unlock()
	defer func() {
		compatMu.Lock()
		defaultDB = origDB
		compatMu.Unlock()
	}()
	if IsAlive() {
		t.Error("IsAlive() should be false when not initialized")
	}
}

func TestIsAlive_Initialized(t *testing.T) {
	db, _ := newMockDatabase(t)

	compatMu.Lock()
	origDB := defaultDB
	origDBPtr := DB
	defaultDB = db
	DB = db.DB()
	compatMu.Unlock()
	defer func() {
		compatMu.Lock()
		defaultDB = origDB
		DB = origDBPtr
		compatMu.Unlock()
	}()
	if !IsAlive() {
		t.Error("IsAlive() should be true when connected")
	}
}

func TestErrNotInitialized(t *testing.T) {
	if ErrNotInitialized.Error() != "database not initialized" {
		t.Errorf("ErrNotInitialized.Error() = %q", ErrNotInitialized.Error())
	}
}

func TestCreateDatabaseConnection_UsesStub(t *testing.T) {
	stubConnectStandardWithMock(t)
	defer Cleanup()
	cfg := &config.DatabaseConfiguration{
		Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db",
	}
	gormDB, err := CreateDatabaseConnection(cfg)
	if err != nil {
		t.Fatalf("CreateDatabaseConnection: %v", err)
	}
	if gormDB == nil {
		t.Fatal("nil")
	}
}
