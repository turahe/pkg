package database

import (
	"context"
	"testing"
	"time"

	"github.com/turahe/pkg/config"
)

func TestDatabase_Health_NilDB(t *testing.T) {
	d := &Database{db: nil, opts: &Options{PingTimeout: 5 * time.Second}}
	err := d.Health(context.Background())
	if err == nil {
		t.Error("Health() should fail when db is nil")
	}
	if err != nil && err.Error() != "database not initialized" {
		t.Errorf("Health() error = %v, want 'database not initialized'", err)
	}
}

func TestDatabase_Close_NoCleanups(t *testing.T) {
	cfg := &config.DatabaseConfiguration{
		Driver:  "sqlite",
		Dbname:  "test_close",
		Logmode: false,
	}
	db, err := New(cfg, Options{})
	if err != nil {
		t.Skipf("New: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Errorf("Close() = %v", err)
	}
	if err := db.Close(); err != nil {
		t.Errorf("Close() second call = %v", err)
	}
}

func TestDatabase_Close_Idempotent(t *testing.T) {
	cfg := &config.DatabaseConfiguration{
		Driver:  "sqlite",
		Dbname:  "test_close_idempotent",
		Logmode: false,
	}
	db, err := New(cfg, Options{})
	if err != nil {
		t.Skipf("New: %v", err)
	}
	_ = db.Close()
	_ = db.Close()
}

func TestDatabase_DB(t *testing.T) {
	cfg := &config.DatabaseConfiguration{
		Driver:  "sqlite",
		Dbname:  "test_db_method",
		Logmode: false,
	}
	db, err := New(cfg, Options{})
	if err != nil {
		t.Skipf("New: %v", err)
	}
	defer db.Close()
	gormDB := db.DB()
	if gormDB == nil {
		t.Fatal("DB() returned nil")
	}
}
