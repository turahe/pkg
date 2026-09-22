package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
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
	db, _ := newMockDatabase(t)
	if err := db.Close(); err != nil {
		t.Errorf("Close() = %v", err)
	}
	if err := db.Close(); err != nil {
		t.Errorf("Close() second call = %v", err)
	}
}

func TestDatabase_Close_Idempotent(t *testing.T) {
	db, _ := newMockDatabase(t)
	_ = db.Close()
	_ = db.Close()
}

func TestDatabase_DB(t *testing.T) {
	db, _ := newMockDatabase(t)
	defer db.Close()
	if db.DB() == nil {
		t.Fatal("DB() returned nil")
	}
}

func TestDatabase_Close_CleanupError(t *testing.T) {
	d := &Database{
		db:   &gorm.DB{},
		opts: &Options{},
		cleanups: []func() error{
			func() error { return errors.New("cleanup1") },
		},
	}
	if err := d.Close(); err == nil {
		t.Fatal("expected close error")
	}
}
