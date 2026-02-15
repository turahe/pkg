package database

import (
	"context"
	"sync"

	"github.com/turahe/pkg/config"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	defaultDB     *Database
	defaultDBSite *Database
	DB            *gorm.DB
	DBSite        *gorm.DB
	compatMu      sync.RWMutex
)

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

func GetDB() *gorm.DB {
	compatMu.RLock()
	db := DB
	compatMu.RUnlock()
	if db == nil {
		panic("database not initialized: call database.Setup() first")
	}
	return db
}

func GetDBSite() *gorm.DB {
	compatMu.RLock()
	db := DBSite
	compatMu.RUnlock()
	if db != nil {
		return db
	}
	return GetDB()
}

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

func IsAlive() bool {
	compatMu.RLock()
	db := defaultDB
	compatMu.RUnlock()
	if db == nil {
		return false
	}
	return db.Health(context.Background()) == nil
}

var ErrNotInitialized = &errNotInitialized{}

type errNotInitialized struct{}

func (e *errNotInitialized) Error() string {
	return "database not initialized"
}
