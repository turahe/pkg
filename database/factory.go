package database

import (
	"context"
	"fmt"

	"github.com/turahe/pkg/config"
	"gorm.io/gorm"
)

type Database struct {
	db       *gorm.DB
	opts     *Options
	cleanups []func() error
}

func (d *Database) DB() *gorm.DB {
	return d.db
}

func New(cfg *config.DatabaseConfiguration, opts Options, override ...Option) (*Database, error) {
	return NewContext(context.Background(), cfg, opts, override...)
}

func NewContext(ctx context.Context, cfg *config.DatabaseConfiguration, opts Options, override ...Option) (*Database, error) {
	o := &Options{}
	*o = opts
	for _, fn := range override {
		fn(o)
	}
	o.applyDefaults()
	var db *gorm.DB
	var cleanup func() error
	var err error
	switch cfg.Driver {
	case "cloudsql-postgres":
		db, cleanup, err = connectCloudSQLPostgres(ctx, cfg, o)
	case "cloudsql-mysql":
		db, cleanup, err = connectCloudSQLMySQL(ctx, cfg, o)
	default:
		db, cleanup, err = connectStandard(ctx, cfg, o)
	}
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}
	cleanups := []func() error{}
	if cleanup != nil {
		cleanups = append(cleanups, cleanup)
	}
	return &Database{db: db, opts: o, cleanups: cleanups}, nil
}
