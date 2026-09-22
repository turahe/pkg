package database

import (
	"context"
	"fmt"

	"github.com/turahe/pkg/config"
	pkglogger "github.com/turahe/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func buildDSN(cfg *config.DatabaseConfiguration) (string, error) {
	switch cfg.Driver {
	case "mysql":
		tz, err := resolveConnectionTimezone(cfg)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=%s",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Dbname, mysqlLocQueryValue(tz)), nil
	case "postgres":
		tz, err := resolveConnectionTimezone(cfg)
		if err != nil {
			return "", err
		}
		mode := "disable"
		if cfg.Sslmode {
			mode = "require"
		}
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s timezone=%s",
			cfg.Host, cfg.Username, cfg.Password, cfg.Dbname, cfg.Port, mode, tz), nil
	case "sqlserver":
		mode := "disable"
		if cfg.Sslmode {
			mode = "true"
		}
		return fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s&encrypt=%s",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Dbname, mode), nil
	default:
		return "", fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}
}

var (
	openGormMySQL = func(dsn string, opts *Options) (*gorm.DB, error) {
		return gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger:      newFintechLogger(opts),
			PrepareStmt: true,
		})
	}
	openGormPostgres = func(dsn string, opts *Options) (*gorm.DB, error) {
		return gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger:      newFintechLogger(opts),
			PrepareStmt: true,
		})
	}
	openGormSQLServer = func(dsn string, opts *Options) (*gorm.DB, error) {
		return gorm.Open(sqlserver.Open(dsn), &gorm.Config{
			Logger:      newFintechLogger(opts),
			PrepareStmt: true,
		})
	}
)

func connectStandard(ctx context.Context, cfg *config.DatabaseConfiguration, opts *Options) (*gorm.DB, func() error, error) {
	dsn, err := buildDSN(cfg)
	if err != nil {
		return nil, nil, err
	}
	var db *gorm.DB
	switch cfg.Driver {
	case "mysql":
		db, err = openGormMySQL(dsn, opts)
	case "postgres":
		db, err = openGormPostgres(dsn, opts)
	case "sqlserver":
		db, err = openGormSQLServer(dsn, opts)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("open: %w", err)
	}
	sqlDB, err := gormSQLDB(db)
	if err != nil {
		return nil, nil, fmt.Errorf("get sql.DB: %w", err)
	}
	configurePool(sqlDB, cfg, opts)
	pingCtx, cancel := context.WithTimeout(ctx, opts.PingTimeout)
	defer cancel()
	if err := pingSQLDB(pingCtx, sqlDB); err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("ping: %w", err)
	}
	logDatabaseConnected(cfg)
	return db, func() error { return sqlDB.Close() }, nil
}

func logDatabaseConnected(cfg *config.DatabaseConfiguration) {
	pkglogger.Info("[DB] database connected", pkglogger.Fields{
		"driver": cfg.Driver, "host": cfg.Host, "port": cfg.Port, "dbname": cfg.Dbname,
	})
}
