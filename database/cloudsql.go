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
	mysqlDriverMu      sync.Mutex
	mysqlDriverReady   bool
	mysqlDriverCleanup func() error

	// pingSQLDB is the pool ping used after open; overridden in tests to avoid needing a live Cloud SQL instance.
	pingSQLDB = func(ctx context.Context, db *sql.DB) error {
		return db.PingContext(ctx)
	}

	newCloudSQLDialer       = cloudsqlconn.NewDialer
	registerCloudSQLMySQL   = cloudsqlmysql.RegisterDriver
	cloudSQLMySQLDriverName = "cloudsql-mysql"
	sqlOpen                 = sql.Open
	gormSQLDB               = func(db *gorm.DB) (*sql.DB, error) { return db.DB() }
	sqlDBClose              = func(db *sql.DB) error { return db.Close() }
	cloudSQLDial            = func(ctx context.Context, dialer *cloudsqlconn.Dialer, instance string) (net.Conn, error) {
		return dialer.Dial(ctx, instance)
	}

	connectCloudSQLPostgresFn = connectCloudSQLPostgres
	connectCloudSQLMySQLFn    = connectCloudSQLMySQL
	connectStandardFn         = connectStandard

	openCloudSQLPostgresGORM = func(sqlDB *sql.DB, opts *Options) (*gorm.DB, error) {
		return gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
			Logger:      newFintechLogger(opts),
			PrepareStmt: true,
		})
	}
	openCloudSQLMySQLGORM = func(cfg *config.DatabaseConfiguration, opts *Options, loc string) (*gorm.DB, error) {
		dsn := fmt.Sprintf("%s:%s@%s(%s)/%s?parseTime=true&loc=%s",
			cfg.Username, cfg.Password, cloudSQLMySQLDriverName, cfg.CloudSQLInstance, cfg.Dbname, loc)
		if opts.UseIAM {
			dsn = fmt.Sprintf("%s@%s(%s)/%s?parseTime=true&loc=%s", cfg.Username, cloudSQLMySQLDriverName, cfg.CloudSQLInstance, cfg.Dbname, loc)
		}
		return gorm.Open(mysql.New(mysql.Config{
			DriverName: cloudSQLMySQLDriverName,
			DSN:        dsn,
		}), &gorm.Config{Logger: newFintechLogger(opts), PrepareStmt: true})
	}
)

// defaultCloudSQLDial retains the package default for coverage tests that override cloudSQLDial.
var defaultCloudSQLDial = cloudSQLDial

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
	dialer, err := newCloudSQLDialer(ctx, buildDialerOptions(opts)...)
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
	tz, err := resolveConnectionTimezone(cfg)
	if err != nil {
		dialer.Close()
		return nil, nil, err
	}
	pgxConfig.RuntimeParams["timezone"] = tz
	instance := cfg.CloudSQLInstance
	pgxConfig.DialFunc = makeCloudSQLDialFunc(dialer, instance)
	postgresDriverOnce.Do(func() {
		sql.Register("cloudsql-postgres", stdlib.GetDefaultDriver())
	})
	connStr := stdlib.RegisterConnConfig(pgxConfig)
	sqlDB, err := sqlOpen("cloudsql-postgres", connStr)
	if err != nil {
		dialer.Close()
		return nil, nil, fmt.Errorf("open: %w", err)
	}
	configurePool(sqlDB, cfg, opts)
	pingCtx, cancel := context.WithTimeout(ctx, opts.PingTimeout)
	defer cancel()
	if pingErr := pingSQLDB(pingCtx, sqlDB); pingErr != nil {
		sqlDB.Close()
		dialer.Close()
		return nil, nil, fmt.Errorf("ping: %w", pingErr)
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
	db, err := openCloudSQLPostgresGORM(sqlDB, opts)
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
	tz, err := resolveConnectionTimezone(cfg)
	if err != nil {
		return nil, nil, err
	}
	mysqlDriverMu.Lock()
	if !mysqlDriverReady {
		var registerErr error
		mysqlDriverCleanup, registerErr = registerCloudSQLMySQL(cloudSQLMySQLDriverName, buildDialerOptions(opts)...)
		if registerErr != nil {
			mysqlDriverMu.Unlock()
			return nil, nil, fmt.Errorf("register cloudsql-mysql driver: %w", registerErr)
		}
		mysqlDriverReady = true
	}
	mysqlDriverMu.Unlock()
	loc := mysqlLocQueryValue(tz)
	db, err := openCloudSQLMySQLGORM(cfg, opts, loc)
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
	if pingErr := pingSQLDB(pingCtx, sqlDB); pingErr != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("ping: %w", pingErr)
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
		if closeErr := sqlDBClose(sqlDBToClose); closeErr != nil {
			return closeErr
		}
		var cleanupErr error
		mysqlCleanupOnce.Do(func() {
			if mysqlDriverCleanup != nil {
				cleanupErr = mysqlDriverCleanup()
			}
		})
		return cleanupErr
	}
	return db, closeFn, nil
}

var mysqlCleanupOnce sync.Once

func makeCloudSQLDialFunc(dialer *cloudsqlconn.Dialer, instance string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, _, _ string) (net.Conn, error) {
		return cloudSQLDial(ctx, dialer, instance)
	}
}
