package database

import (
	"database/sql"
	"time"

	"github.com/turahe/pkg/config"
)

func configurePool(sqlDB *sql.DB, cfg *config.DatabaseConfiguration, opts *Options) {
	maxOpen := opts.MaxOpenConns
	if cfg.MaxOpenConns > 0 {
		maxOpen = cfg.MaxOpenConns
	}
	maxIdle := opts.MaxIdleConns
	if cfg.MaxIdleConns > 0 {
		maxIdle = cfg.MaxIdleConns
	}
	connMaxLife := opts.ConnMaxLife
	if cfg.ConnMaxLifetimeMinutes > 0 {
		connMaxLife = time.Duration(cfg.ConnMaxLifetimeMinutes) * time.Minute
	}
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(connMaxLife)
	sqlDB.SetConnMaxIdleTime(opts.ConnMaxIdle)
}
