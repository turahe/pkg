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
	compatMu.Lock()
	origDB := defaultDB
	origDBPtr := DB
	compatMu.Unlock()
	defer func() {
		Cleanup()
		compatMu.Lock()
		defaultDB = origDB
		DB = origDBPtr
		compatMu.Unlock()
	}()
	cfg := &config.DatabaseConfiguration{
		Driver:  "sqlite",
		Dbname:  "test_isalive",
		Logmode: false,
	}
	db, err := New(cfg, Options{})
	if err != nil {
		t.Skipf("New: %v", err)
	}
	compatMu.Lock()
	defaultDB = db
	DB = db.DB()
	compatMu.Unlock()
	if !IsAlive() {
		t.Error("IsAlive() should be true when connected")
	}
}

func TestErrNotInitialized(t *testing.T) {
	if ErrNotInitialized.Error() != "database not initialized" {
		t.Errorf("ErrNotInitialized.Error() = %q", ErrNotInitialized.Error())
	}
}
