package database

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/turahe/pkg/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openMockGORM returns a *gorm.DB backed by sqlmock (no real database, no CGO).
// Pings are not monitored so Health/IsAlive work without extra expectations.
func openMockGORM(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() {
		mock.ExpectClose()
		_ = sqlDB.Close()
	})
	gdb, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return gdb, mock
}

func newMockDatabase(t *testing.T) (*Database, sqlmock.Sqlmock) {
	t.Helper()
	gdb, mock := openMockGORM(t)
	opts := &Options{PingTimeout: time.Second, LogLevel: logger.Silent}
	opts.applyDefaults()
	return &Database{
		db:       gdb,
		opts:     opts,
		cleanups: nil, // sqlDB closed via openMockGORM cleanup
	}, mock
}

// stubConnectStandardWithMock makes New/Setup succeed offline via sqlmock for supported drivers.
func stubConnectStandardWithMock(t *testing.T) {
	t.Helper()
	orig := connectStandardFn
	connectStandardFn = func(ctx context.Context, cfg *config.DatabaseConfiguration, opts *Options) (*gorm.DB, func() error, error) {
		switch cfg.Driver {
		case "mysql", "postgres", "sqlserver":
			gdb, mock := openMockGORM(t)
			_ = mock
			return gdb, func() error { return nil }, nil
		default:
			return connectStandard(ctx, cfg, opts)
		}
	}
	t.Cleanup(func() { connectStandardFn = orig })
}
