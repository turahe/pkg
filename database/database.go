package database

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/turahe/pkg/config"

	"cloud.google.com/go/cloudsqlconn"
	cloudsqlmysql "cloud.google.com/go/cloudsqlconn/mysql/mysql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB                   *gorm.DB             // Main database connection
	DBSite               *gorm.DB             // Site database connection
	cloudSQLDialer       *cloudsqlconn.Dialer // Store dialer for Postgres Cloud SQL connections
	cloudSQLMySQLCleanup func() error         // Store cleanup function for MySQL Cloud SQL connections
	dialerMutex          sync.Mutex           // Mutex to protect dialer initialization
)

type Database struct {
	*gorm.DB
}

func Setup() error {
	configuration := config.GetConfig()

	// Setup main database connection
	db, err := CreateDatabaseConnection(&configuration.Database)
	if err != nil {
		return err
	}
	DB = db

	// Setup site database connection (optional - only if configured)
	if configuration.DatabaseSite.Dbname != "" {
		dbSite, err := CreateDatabaseConnection(&configuration.DatabaseSite)
		if err != nil {
			return fmt.Errorf("failed to setup site database connection: %w", err)
		}
		DBSite = dbSite
	}

	return nil
}

func CreateDatabaseConnection(configuration *config.DatabaseConfiguration) (*gorm.DB, error) {
	logmode := configuration.Logmode
	loglevel := logger.Silent
	if logmode {
		loglevel = logger.Info
	}
	newDBLogger := logger.New(
		log.New(getWriter(), "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  loglevel,    // Log level (Silent, Error, Warn, Info)
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,       // Disable color
		},
	)

	var db *gorm.DB
	var err error

	// Handle Cloud SQL drivers separately (they don't need buildDSN)
	switch configuration.Driver {
	case "cloudsql-mysql":
		db, err = createCloudSQLMySQLConnection(configuration, newDBLogger)
	case "cloudsql-postgres":
		db, err = createCloudSQLPostgresConnection(configuration, newDBLogger)
	default:
		// For standard drivers, build DSN first
		dsn, dsnErr := buildDSN(configuration)
		if dsnErr != nil {
			return nil, fmt.Errorf("failed to build DSN: %w", dsnErr)
		}

		switch configuration.Driver {
		case "mysql":
			db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newDBLogger})
		case "postgres":
			db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newDBLogger})
		case "sqlite":
			db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: newDBLogger})
		case "sqlserver":
			db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{Logger: newDBLogger})
		default:
			return nil, fmt.Errorf("unsupported database driver: %s", configuration.Driver)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool settings for all database types
	if sqlDB, err := db.DB(); err == nil {
		configureConnectionPool(sqlDB)
	}

	return db, nil

}

func buildDSN(configuration *config.DatabaseConfiguration) (string, error) {
	switch configuration.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", configuration.Username, configuration.Password, configuration.Host, configuration.Port, configuration.Dbname), nil
	case "postgres":
		mode := "disable"
		if configuration.Sslmode {
			mode = "require"
		}
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", configuration.Host, configuration.Username, configuration.Password, configuration.Dbname, configuration.Port, mode), nil
	case "sqlite":
		return "./" + configuration.Dbname + ".db", nil
	case "sqlserver":
		mode := "disable"
		if configuration.Sslmode {
			mode = "true"
		}
		return fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s&encrypt=%s", configuration.Username, configuration.Password, configuration.Host, configuration.Port, configuration.Dbname, mode), nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s", configuration.Driver)
	}
}

func getWriter() io.Writer {
	file, err := os.OpenFile("log/database.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return os.Stdout
	} else {
		return file
	}
}

func GetDB() *gorm.DB {
	if DB == nil {
		panic("Database is not initialized. Call database.Setup() first.")
	}
	return DB
}

// GetDBSite returns the site database connection
// Returns the main DB if site DB is not configured
func GetDBSite() *gorm.DB {
	if DBSite != nil {
		return DBSite
	}
	// Fallback to main DB if site DB is not configured
	return GetDB()
}

// IsAlive returns true if the main database connection is healthy (ping succeeds).
func IsAlive() bool {
	if DB == nil {
		return false
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}

// HealthCheck pings the main database (and site database if configured) and returns an error if any connection is unhealthy.
// Use context for timeout, e.g. context.WithTimeout(ctx, 3*time.Second).
func HealthCheck(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("main db: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("main db ping: %w", err)
	}
	if DBSite != nil {
		sqlSite, err := DBSite.DB()
		if err != nil {
			return fmt.Errorf("site db: %w", err)
		}
		if err := sqlSite.PingContext(ctx); err != nil {
			return fmt.Errorf("site db ping: %w", err)
		}
	}
	return nil
}

func configureConnectionPool(sqlDB *sql.DB) {
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour * 24)
}

func createCloudSQLMySQLConnection(configuration *config.DatabaseConfiguration, dbLogger logger.Interface) (*gorm.DB, error) {
	if configuration.CloudSQLInstance == "" {
		return nil, fmt.Errorf("CloudSQLInstance is required for cloudsql-mysql driver")
	}

	ctx := context.Background()

	dialerMutex.Lock()
	defer dialerMutex.Unlock()

	// Register the Cloud SQL MySQL driver only once (reuse if already registered)
	if cloudSQLMySQLCleanup == nil {
		cleanup, err := cloudsqlmysql.RegisterDriver("cloudsql-mysql")
		if err != nil {
			return nil, fmt.Errorf("failed to register Cloud SQL MySQL driver: %w", err)
		}
		// Store cleanup function globally - do NOT defer cleanup() here
		// The cleanup should only be called on application shutdown
		cloudSQLMySQLCleanup = cleanup
	}

	// Build DSN for Cloud SQL MySQL
	dsn := fmt.Sprintf("%s:%s@cloudsql-mysql(%s)/%s?parseTime=true",
		configuration.Username,
		configuration.Password,
		configuration.CloudSQLInstance,
		configuration.Dbname,
	)

	// Create GORM DB instance using the mysql driver with custom driver name
	db, err := gorm.Open(mysql.New(mysql.Config{
		DriverName: "cloudsql-mysql",
		DSN:        dsn,
	}), &gorm.Config{Logger: dbLogger})
	if err != nil {
		return nil, fmt.Errorf("failed to create GORM connection: %w", err)
	}

	// Get the underlying sql.DB to test the connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	configureConnectionPool(sqlDB)

	// Test the connection
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Cloud SQL MySQL database: %w", err)
	}

	return db, nil
}

func createCloudSQLPostgresConnection(configuration *config.DatabaseConfiguration, dbLogger logger.Interface) (*gorm.DB, error) {
	if configuration.CloudSQLInstance == "" {
		return nil, fmt.Errorf("CloudSQLInstance is required for cloudsql-postgres driver")
	}

	ctx := context.Background()

	dialerMutex.Lock()
	defer dialerMutex.Unlock()

	// Create a new dialer for Cloud SQL (reuse if already exists)
	if cloudSQLDialer == nil {
		dialer, err := cloudsqlconn.NewDialer(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create Cloud SQL dialer: %w", err)
		}
		cloudSQLDialer = dialer
	}
	dialer := cloudSQLDialer

	// Build DSN for Cloud SQL Postgres
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s port=%s sslmode=disable",
		configuration.Username,
		configuration.Password,
		configuration.Dbname,
		configuration.Port,
	)

	// Parse pgx config
	pgxConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	// Set the Cloud SQL dialer function
	pgxConfig.DialFunc = func(ctx context.Context, network string, instance string) (net.Conn, error) {
		return dialer.Dial(ctx, configuration.CloudSQLInstance)
	}

	// Register the pgx driver with stdlib to get a *sql.DB
	driverName := "cloudsql-postgres"
	sql.Register(driverName, stdlib.GetDefaultDriver())

	// Create a connection string for stdlib
	connStr := stdlib.RegisterConnConfig(pgxConfig)

	// Open database connection using the registered driver
	sqlDB, err := sql.Open(driverName, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open Cloud SQL Postgres connection: %w", err)
	}

	// Configure connection pool
	configureConnectionPool(sqlDB)

	// Test the connection first
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Cloud SQL Postgres database: %w", err)
	}

	// Create GORM DB instance using the postgres driver with our sql.DB
	// Use Dialector with the underlying sql.DB
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{Logger: dbLogger})
	if err != nil {
		return nil, fmt.Errorf("failed to create GORM connection: %w", err)
	}

	return db, nil
}

// Cleanup closes Cloud SQL connections and dialers
// This should be called at application shutdown for graceful cleanup
func Cleanup() error {
	dialerMutex.Lock()
	defer dialerMutex.Unlock()

	var errs []error

	// Close Postgres dialer
	if cloudSQLDialer != nil {
		if err := cloudSQLDialer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Cloud SQL Postgres dialer: %w", err))
		}
		cloudSQLDialer = nil
	}

	// Call MySQL cleanup function
	if cloudSQLMySQLCleanup != nil {
		if err := cloudSQLMySQLCleanup(); err != nil {
			errs = append(errs, fmt.Errorf("failed to cleanup Cloud SQL MySQL driver: %w", err))
		}
		cloudSQLMySQLCleanup = nil
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing Cloud SQL connections: %v", errs)
	}

	return nil
}
