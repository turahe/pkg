package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"sync"

	"github.com/turahe/pkg/config"
	pkglogger "github.com/turahe/pkg/logger"

	"cloud.google.com/go/cloudsqlconn"
	cloudsqlmysql "cloud.google.com/go/cloudsqlconn/mysql/mysql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	postgresDriverOnce sync.Once
	mysqlDriverOnce    sync.Once
	mysqlDriverCleanup func() error
)

func buildDialerOptions(opts *Options) []cloudsqlconn.Option {
	var dialOpts []cloudsqlconn.DialOption
	if opts.UsePrivateIP {
		dialOpts = append(dialOpts, cloudsqlconn.WithPrivateIP())
	}
	var result []cloudsqlconn.Option
	if opts.UseIAM {
		result = append(result, cloudsqlconn.WithIAMAuthN())
	}
	if len(dialOpts) > 0 {
		result = append(result, cloudsqlconn.WithDefaultDialOptions(dialOpts...))
	}
	return result
}

func connectCloudSQLPostgres(ctx context.Context, cfg *config.DatabaseConfiguration, opts *Options) (*gorm.DB, func() error, error) {
	if cfg.CloudSQLInstance == "" {
		return nil, nil, fmt.Errorf("cloud_sql_instance required for cloudsql-postgres")
	}
	dialer, err := cloudsqlconn.NewDialer(ctx, buildDialerOptions(opts)...)
	if err != nil {
		return nil, nil, fmt.Errorf("create dialer: %w", err)
	}
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Username, cfg.Password, cfg.Dbname, cfg.Port)
	if opts.UseIAM {
		dsn = fmt.Sprintf("user=%s dbname=%s port=%s sslmode=disable", cfg.Username, cfg.Dbname, cfg.Port)
	}
	pgxConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		dialer.Close()
		return nil, nil, fmt.Errorf("parse pgx config: %w", err)
	}
	instance := cfg.CloudSQLInstance
	pgxConfig.DialFunc = func(ctx context.Context, _, _ string) (net.Conn, error) {
		return dialer.Dial(ctx, instance)
	}
	postgresDriverOnce.Do(func() {
		sql.Register("cloudsql-postgres", stdlib.GetDefaultDriver())
	})
	connStr := stdlib.RegisterConnConfig(pgxConfig)
	sqlDB, err := sql.Open("cloudsql-postgres", connStr)
	if err != nil {
		dialer.Close()
		return nil, nil, fmt.Errorf("open: %w", err)
	}
	configurePool(sqlDB, cfg, opts)
	pingCtx, cancel := context.WithTimeout(ctx, opts.PingTimeout)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		sqlDB.Close()
		dialer.Close()
		return nil, nil, fmt.Errorf("ping: %w", err)
	}
	connType := "public_ip"
	if opts.UsePrivateIP {
		connType = "private_ip"
	}
	pkglogger.Info("[DB] Cloud SQL Postgres connected", pkglogger.Fields{
		"driver":          "cloudsql-postgres",
		"instance":        instance,
		"connection_type": connType,
		"iam_auth":        opts.UseIAM,
	})
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger:      newFintechLogger(opts),
		PrepareStmt: true,
	})
	if err != nil {
		sqlDB.Close()
		dialer.Close()
		return nil, nil, err
	}
	return db, dialer.Close, nil
}

func connectCloudSQLMySQL(ctx context.Context, cfg *config.DatabaseConfiguration, opts *Options) (*gorm.DB, func() error, error) {
	if cfg.CloudSQLInstance == "" {
		return nil, nil, fmt.Errorf("cloud_sql_instance required for cloudsql-mysql")
	}
	var cleanup func() error
	mysqlDriverOnce.Do(func() {
		cleanup, _ = cloudsqlmysql.RegisterDriver("cloudsql-mysql", buildDialerOptions(opts)...)
		mysqlDriverCleanup = cleanup
	})
	dsn := fmt.Sprintf("%s:%s@cloudsql-mysql(%s)/%s?parseTime=true",
		cfg.Username, cfg.Password, cfg.CloudSQLInstance, cfg.Dbname)
	if opts.UseIAM {
		dsn = fmt.Sprintf("%s@cloudsql-mysql(%s)/%s?parseTime=true", cfg.Username, cfg.CloudSQLInstance, cfg.Dbname)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		DriverName: "cloudsql-mysql",
		DSN:        dsn,
	}), &gorm.Config{Logger: newFintechLogger(opts), PrepareStmt: true})
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
	connType := "public_ip"
	if opts.UsePrivateIP {
		connType = "private_ip"
	}
	pkglogger.Info("[DB] Cloud SQL MySQL connected", pkglogger.Fields{
		"driver":          "cloudsql-mysql",
		"instance":        cfg.CloudSQLInstance,
		"connection_type": connType,
		"iam_auth":        opts.UseIAM,
	})
	sqlDBToClose := sqlDB
	closeFn := func() error {
		sqlDBToClose.Close()
		mysqlCleanupOnce.Do(func() {
			if mysqlDriverCleanup != nil {
				mysqlDriverCleanup()
			}
		})
		return nil
	}
	return db, closeFn, nil
}

var mysqlCleanupOnce sync.Once
