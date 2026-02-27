package database

import (
	"context"
	"sync"

	"github.com/turahe/pkg/config"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// defaultDB, defaultDBSite, DB, DBSite are the legacy global singleton state; compatMu protects concurrent access.
var (
	defaultDB     *Database
	defaultDBSite *Database
	DB            *gorm.DB
	DBSite        *gorm.DB
	compatMu      sync.RWMutex
)

// Setup initializes the global database from config.GetConfig(): primary database always, site database if DatabaseSite.Dbname is set. Panics if New fails.
func Setup() error {
	cfg := config.GetConfig()
	opts := Options{}
	if cfg.Database.Logmode {
		opts.LogLevel = logger.Info
	}
	db, err := New(&cfg.Database, opts)
	if err != nil {
		return err
	}
	compatMu.Lock()
	defaultDB = db
	DB = db.DB()
	compatMu.Unlock()
	if cfg.DatabaseSite.Dbname != "" {
		optsSite := Options{}
		if cfg.DatabaseSite.Logmode {
			optsSite.LogLevel = logger.Info
		}
		dbSite, err := New(&cfg.DatabaseSite, optsSite)
		if err != nil {
			return err
		}
		compatMu.Lock()
		defaultDBSite = dbSite
		DBSite = dbSite.DB()
		compatMu.Unlock()
	}
	return nil
}

// GetDB returns the global primary *gorm.DB. Panics if database not initialized (call Setup first).
func GetDB() *gorm.DB {
	compatMu.RLock()
	db := DB
	compatMu.RUnlock()
	if db == nil {
		panic("database not initialized: call database.Setup() first")
	}
	return db
}

// GetDBSite returns the global site *gorm.DB if configured; otherwise returns GetDB().
func GetDBSite() *gorm.DB {
	compatMu.RLock()
	db := DBSite
	compatMu.RUnlock()
	if db != nil {
		return db
	}
	return GetDB()
}

// HealthCheck pings both primary and site databases (if present). Returns ErrNotInitialized if Setup was not called.
func HealthCheck(ctx context.Context) error {
	compatMu.RLock()
	db := defaultDB
	dbSite := defaultDBSite
	compatMu.RUnlock()
	if db == nil {
		return ErrNotInitialized
	}
	if err := db.Health(ctx); err != nil {
		return err
	}
	if dbSite != nil {
		return dbSite.Health(ctx)
	}
	return nil
}

// Cleanup closes both defaultDB and defaultDBSite and clears globals. Returns the first error from Close if any.
func Cleanup() error {
	compatMu.Lock()
	defer compatMu.Unlock()
	var errs []error
	if defaultDB != nil {
		if err := defaultDB.Close(); err != nil {
			errs = append(errs, err)
		}
		defaultDB = nil
		DB = nil
	}
	if defaultDBSite != nil {
		if err := defaultDBSite.Close(); err != nil {
			errs = append(errs, err)
		}
		defaultDBSite = nil
		DBSite = nil
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// CreateDatabaseConnection creates a Database from cfg, sets it as the global defaultDB, and returns its *gorm.DB. For legacy callers that need a raw *gorm.DB.
func CreateDatabaseConnection(cfg *config.DatabaseConfiguration) (*gorm.DB, error) {
	opts := Options{}
	if cfg.Logmode {
		opts.LogLevel = logger.Info
	}
	db, err := New(cfg, opts)
	if err != nil {
		return nil, err
	}
	compatMu.Lock()
	defaultDB = db
	DB = db.DB()
	compatMu.Unlock()
	return db.DB(), nil
}

// IsAlive returns true if the primary database responds to Health(context.Background()); false if not initialized or ping fails.
func IsAlive() bool {
	compatMu.RLock()
	db := defaultDB
	compatMu.RUnlock()
	if db == nil {
		return false
	}
	return db.Health(context.Background()) == nil
}

// ErrNotInitialized is returned by HealthCheck when Setup has not been called or defaultDB is nil.
var ErrNotInitialized = &errNotInitialized{}

// errNotInitialized implements error for ErrNotInitialized.
type errNotInitialized struct{}

// Error returns "database not initialized".
func (e *errNotInitialized) Error() string {
	return "database not initialized"
}
