package database

import (
	"fmt"

	"github.com/turahe/pkg/config"
	"github.com/turahe/pkg/otelx"
	"gorm.io/gorm"
)

func applyOpenTelemetryDefaults(opts Options) Options {
	if opts.EnableOpenTelemetry {
		return opts
	}
	cfg := config.GetConfig()
	if cfg != nil && otelx.GORMEnabled(cfg.OpenTelemetry) {
		opts.EnableOpenTelemetry = true
	}
	return opts
}

func registerGORMInstrumentation(db *gorm.DB, cfg *config.DatabaseConfiguration, opts *Options) error {
	if !opts.EnableOpenTelemetry {
		return nil
	}
	return otelx.RegisterGORM(db, otelx.GORMOptions{
		DBSystem: dbSystemForDriver(cfg.Driver),
	})
}

func dbSystemForDriver(driver string) string {
	switch driver {
	case "mysql", "cloudsql-mysql":
		return "mysql"
	case "postgres", "cloudsql-postgres":
		return "postgresql"
	case "sqlite":
		return "sqlite"
	case "sqlserver":
		return "mssql"
	default:
		return driver
	}
}

func wrapConnectErr(err error, cleanup func() error) error {
	if cleanup != nil {
		_ = cleanup()
	}
	return fmt.Errorf("otel gorm: %w", err)
}
