# Go Package Collection

A collection of reusable Go packages for common functionality including database connections, Redis, JWT, crypto, GCS, logging, and utilities.

## Packages

### `crypto`
Password hashing and comparison using bcrypt.

**Functions:**
- `HashAndSalt(plainPassword []byte) string` - Hash a password
- `ComparePassword(hashedPassword string, plainPassword []byte) bool` - Compare password with hash

### `database`
Database connection management with support for multiple drivers (MySQL, PostgreSQL, SQLite, SQL Server) and Cloud SQL.

**Functions:**
- `Setup() error` - Initialize database connections
- `CreateDatabaseConnection(configuration *config.DatabaseConfiguration) (*gorm.DB, error)` - Create a database connection
- `GetDB() *gorm.DB` - Get the main database connection
- `GetDBSite() *gorm.DB` - Get the site database connection
- `Cleanup() error` - Cleanup Cloud SQL connections

### `gcs`
Google Cloud Storage client wrapper.

**Functions:**
- `Setup() error` - Initialize GCS client
- `GetClient() *storage.Client` - Get GCS client
- `GetBucket() *storage.BucketHandle` - Get bucket handle
- `ReadObject(objectName string) ([]byte, error)` - Read object from bucket
- `WriteObject(objectName string, data []byte, contentType string) error` - Write object to bucket
- `DeleteObject(objectName string) error` - Delete object from bucket
- `ObjectExists(objectName string) (bool, error)` - Check if object exists
- `ListObjects(prefix string) ([]string, error)` - List objects with prefix

### `jwt`
JWT token generation and validation.

**Functions:**
- `Init()` - Initialize JWT with secret from config
- `GenerateToken(id uuid.UUID, email, name string) (string, error)` - Generate access token
- `GenerateRefreshToken(id uuid.UUID, email, name string) (string, error)` - Generate refresh token
- `ValidateToken(tokenString string) (*Claims, error)` - Validate and parse token

### `logger`
Structured logging with Google Cloud Logging format support.

**Functions:**
- `Debugf(format string, args ...interface{})` - Debug log
- `Infof(format string, args ...interface{})` - Info log
- `Warnf(format string, args ...interface{})` - Warning log
- `Errorf(format string, args ...interface{})` - Error log
- `Fatalf(format string, args ...interface{})` - Fatal log (exits)
- `SetLogLevel(level slog.Level)` - Set log level

### `redis`
Redis client wrapper with common operations.

**Functions:**
- `Setup() error` - Initialize Redis client
- `GetRedis() *redis.Client` - Get Redis client
- `IsAlive() bool` - Check if Redis is alive
- String operations: `Get`, `Set`, `Delete`, `MGet`, `MSet`
- Hash operations: `HGet`, `HGetAll`, `HSet`, `HSetMap`
- List operations: `LPush`, `RPop`, `LRange`
- Set operations: `SAdd`, `SMembers`, `SRem`
- Lock operations: `AcquireLock`, `ExtendLock`, `ReleaseLock`
- Pipeline operations: `Pipeline`, `PipelineSet`
- Pub/Sub: `PublishMessage`, `SubscribeToChannel`
- `ScanKeys(pattern string, count int64) ([]string, error)` - Scan keys

### `util`
Utility functions for common operations.

**Functions:**
- `IsEmpty(value interface{}) bool` - Check if value is empty
- `InAnySlice[T comparable](haystack []T, needle T) bool` - Check if value exists in slice
- `RemoveDuplicates[T comparable](haystack []T) []T` - Remove duplicates from slice
- `GetCurrentTimeRange() *types.TimeRange` - Get current time range

### `config`
Configuration management with environment variable support.

**Functions:**
- `Setup(configPath string) error` - Initialize configuration
- `GetConfig() *Configuration` - Get global configuration
- `SetConfig(cfg *Configuration)` - Set global configuration

### `types`
Common type definitions.

**Types:**
- `TimeRange` - Time range with start and end times

## Installation

```bash
go get github.com/turahe/pkg
```

## Usage

### Configuration Setup

```go
import "github.com/turahe/pkg/config"

// Option 1: Setup from environment variables
err := config.Setup("")

// Option 2: Set configuration manually
cfg := &config.Configuration{
    Database: config.DatabaseConfiguration{
        Driver:   "mysql",
        Host:     "localhost",
        Port:     "3306",
        Username: "user",
        Password: "pass",
        Dbname:   "dbname",
    },
}
config.SetConfig(cfg)
```

### Database

```go
import "github.com/turahe/pkg/database"

// Setup database connection
err := database.Setup()
if err != nil {
    log.Fatal(err)
}

// Get database connection
db := database.GetDB()
```

### Redis

```go
import "github.com/turahe/pkg/redis"

// Setup Redis connection
err := redis.Setup()
if err != nil {
    log.Fatal(err)
}

// Use Redis
val, err := redis.Get("key")
```

### JWT

```go
import "github.com/turahe/pkg/jwt"

// Initialize JWT
jwt.Init()

// Generate token
token, err := jwt.GenerateToken(userID, email, name)

// Validate token
claims, err := jwt.ValidateToken(tokenString)
```

### Logger

```go
import "github.com/turahe/pkg/logger"

logger.Infof("Application started")
logger.Errorf("Error occurred: %v", err)
```

## Environment Variables

The package supports configuration via environment variables:

- `DATABASE_DRIVER` - Database driver (mysql, postgres, sqlite, sqlserver, cloudsql-mysql, cloudsql-postgres)
- `DATABASE_HOST` - Database host
- `DATABASE_PORT` - Database port
- `DATABASE_USERNAME` - Database username
- `DATABASE_PASSWORD` - Database password
- `DATABASE_DBNAME` - Database name
- `REDIS_ENABLED` - Enable Redis (true/false)
- `REDIS_HOST` - Redis host
- `REDIS_PORT` - Redis port
- `GCS_ENABLED` - Enable GCS (true/false)
- `GCS_BUCKET_NAME` - GCS bucket name
- `SERVER_SECRET` - JWT secret key
- `SERVER_ACCESS_TOKEN_EXPIRY` - Access token expiry in hours
- `SERVER_REFRESH_TOKEN_EXPIRY` - Refresh token expiry in days

## Testing

Run all tests:

```bash
go test ./...
```

Run with verbose output:

```bash
go test ./... -v
```

Run with race detector:

```bash
go test ./... -race
```

Tests cover: `crypto`, `util`, `config`, `jwt`, `logger`, `redis`, `database`, `gcs`, and `types`. Packages that require external services (Redis, GCS, MySQL/Postgres) use disabled or in-memory/sqlite config where possible so tests can run without those services.

## License

MIT
