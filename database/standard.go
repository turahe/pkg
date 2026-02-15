package database

import (
	"context"
	"fmt"

	"github.com/turahe/pkg/config"
	pkglogger "github.com/turahe/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func buildDSN(cfg *config.DatabaseConfiguration) (string, error) {
	switch cfg.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Dbname), nil
	case "postgres":
		mode := "disable"
		if cfg.Sslmode {
			mode = "require"
		}
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			cfg.Host, cfg.Username, cfg.Password, cfg.Dbname, cfg.Port, mode), nil
	case "sqlite":
		return "./" + cfg.Dbname + ".db", nil
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

func connectStandard(ctx context.Context, cfg *config.DatabaseConfiguration, opts *Options) (*gorm.DB, func() error, error) {
	dsn, err := buildDSN(cfg)
	if err != nil {
		return nil, nil, err
	}
	var db *gorm.DB
	switch cfg.Driver {
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger:      newFintechLogger(opts),
			PrepareStmt: true,
		})
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger:      newFintechLogger(opts),
			PrepareStmt: true,
		})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger:      newFintechLogger(opts),
			PrepareStmt: true,
		})
	case "sqlserver":
		db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{
			Logger:      newFintechLogger(opts),
			PrepareStmt: true,
		})
	default:
		return nil, nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("open: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("get sql.DB: %w", err)
	}
	configurePool(sqlDB, cfg, opts)
	pingCtx, cancel := context.WithTimeout(ctx, opts.PingTimeout)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("ping: %w", err)
	}
	if cfg.Driver == "sqlite" {
		pkglogger.Info("[DB] database connected", pkglogger.Fields{"driver": cfg.Driver, "dbname": cfg.Dbname})
	} else {
		pkglogger.Info("[DB] database connected", pkglogger.Fields{
			"driver": cfg.Driver, "host": cfg.Host, "port": cfg.Port, "dbname": cfg.Dbname,
		})
	}
	return db, func() error { return sqlDB.Close() }, nil
}
