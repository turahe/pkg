# Go Package Collection

[![Go Reference](https://pkg.go.dev/badge/github.com/turahe/pkg.svg)](https://pkg.go.dev/github.com/turahe/pkg)
[![Test](https://github.com/turahe/pkg/actions/workflows/test.yml/badge.svg)](https://github.com/turahe/pkg/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/turahe/pkg)](https://goreportcard.com/report/github.com/turahe/pkg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A collection of reusable Go packages for common functionality including database connections, Redis (with cluster support), JWT, crypto, GCS, logging, middleware, and utilities.

## Packages

### `crypto`
Password hashing and comparison using bcrypt.

**Functions:**
- `HashAndSalt(plainPassword []byte) string` - Hash a password
- `ComparePassword(hashedPassword string, plainPassword []byte) bool` - Compare password with hash

### `database`
Enterprise-grade database layer with dependency injection, Cloud SQL (Postgres/MySQL), IAM auth, Private IP, and production-tuned connection pools.

**Drivers:** mysql, postgres, sqlite, sqlserver, cloudsql-mysql, cloudsql-postgres

**New API (recommended):**
- `New(cfg *config.DatabaseConfiguration, opts Options, override ...Option) (*Database, error)` - Create database with options
- `NewContext(ctx, cfg, opts, override...) (*Database, error)` - Create with context
- `(d *Database) DB() *gorm.DB` - Get GORM instance
- `(d *Database) Health(ctx context.Context) error` - Health check with timeout
- `(d *Database) Close() error` - Graceful shutdown

**Options:** `UseIAM`, `UsePrivateIP`, `LogLevel`, `MaxOpenConns` (default 30), `MaxIdleConns` (default 10), `ConnMaxLife` (30m), `ConnMaxIdle` (10m), `SlowThreshold` (500ms)

**Backward compatibility:**
- `Setup() error` - Initialize from config (uses env)
- `GetDB() *gorm.DB` / `GetDBSite() *gorm.DB` - Get connections
- `CreateDatabaseConnection(cfg) (*gorm.DB, error)` - Create single connection
- `HealthCheck(ctx) error` / `IsAlive() bool` - Health checks
- `Cleanup() error` - Shutdown Cloud SQL

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
Structured logging with Google Cloud Logging format support. Log entries are JSON with `severity`, `time`, `message`, optional `trace_id`/`correlation_id`, `sourceLocation`, and `fields`.

**Functions:**
- `Debugf(format string, args ...interface{})` - Debug log
- `Infof(format string, args ...interface{})` - Info log
- `Warnf(format string, args ...interface{})` - Warning log
- `Errorf(format string, args ...interface{})` - Error log
- `Fatalf(format string, args ...interface{})` - Fatal log (exits)
- `Debug(msg string, fields Fields)` / `Info` / `Warn` / `Error` - Structured log with fields
- `SetLogLevel(level slog.Level)` - Set log level
- `GetLogger() *slog.Logger` - Get underlying slog logger

**Trace ID / Correlation ID (for request-scoped logging):**
- `WithTraceID(ctx, traceID)` / `WithCorrelationID(ctx, correlationID)` - Store IDs in context
- `GetTraceID(ctx)` / `GetCorrelationID(ctx)` - Read IDs from context
- `WithContext(ctx) *Ctx` - Context-bound logger; `Infof`/`Debugf`/etc. automatically include trace_id/correlation_id in JSON
- `DebugfContext(ctx, format, args...)` / `InfofContext` / `WarnfContext` / `ErrorfContext` - Log with context (IDs in JSON)

### `redis`
Redis client wrapper with common operations. Supports both standard Redis and Redis Cluster (Google Cloud Memorystore).

**Functions:**
- `Setup() error` - Initialize Redis client (standard or cluster)
- `GetRedis() *redis.Client` - Get standard Redis client
- `GetRedisCluster() *redis.ClusterClient` - Get Redis cluster client
- `GetUniversalClient() redis.Cmdable` - Get universal client (works with both modes)
- `IsAlive() bool` - Check if Redis is alive
- String operations: `Get`, `Set`, `Delete`, `MGet`, `MSet`
- Hash operations: `HGet`, `HGetAll`, `HSet`, `HSetMap`
- List operations: `LPush`, `RPop`, `LRange`
- Set operations: `SAdd`, `SMembers`, `SRem`
- Lock operations: `AcquireLock`, `ExtendLock`, `ReleaseLock`
- Pipeline operations: `Pipeline`, `PipelineSet`
- Pub/Sub: `PublishMessage`, `SubscribeToChannel`
- `ScanKeys(pattern string, count int64) ([]string, error)` - Scan keys (cluster-aware)

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

### `middlewares`
Gin framework middleware collection for common HTTP operations.

**Middleware Functions:**
- `RequestID() gin.HandlerFunc` - Ensures a request/correlation ID on every request. Reads `X-Request-ID` or `X-Trace-ID` from headers; generates a new UUID if missing. Injects ID into context so LoggerMiddleware and `logger.WithContext` include it in structured logs.
- `TraceMiddleware() gin.HandlerFunc` - Injects trace_id and correlation_id from headers (`X-Trace-Id`, `X-Correlation-Id`, or `X-Request-Id`) into request context; generates UUID if missing. Use before LoggerMiddleware so logs include IDs.
- `LoggerMiddleware() gin.HandlerFunc` - HTTP request logging (method, path, status, latency, IP); includes trace_id/correlation_id in JSON when RequestID or TraceMiddleware is used
- `AuthMiddleware() gin.HandlerFunc` - JWT authentication middleware
- `CORS() gin.HandlerFunc` - CORS middleware
- `RateLimiter() gin.HandlerFunc` - Rate limiting middleware (requires Redis)
- `NoMethodHandler() gin.HandlerFunc` - Handle unsupported HTTP methods
- `NoRouteHandler() gin.HandlerFunc` - Handle 404 routes
- `RecoveryHandler(ctx *gin.Context)` - Panic recovery middleware

### `response`
Response code management and standardized API responses.

**Service Codes:**
- `ServiceCodeCommon` - Common/General services
- `ServiceCodeAuth` - Authentication service
- `ServiceCodeTransaction` - Transaction service
- `ServiceCodeWallet` - Wallet service
- And more...

**Functions:**
- `BuildResponseCode(httpStatus int, serviceCode, caseCode string) int` - Build response code
- `ParseResponseCode(code int) (httpStatus int, serviceCode, caseCode string)` - Parse response code

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

**New API (dependency injection):**

```go
import (
    "github.com/turahe/pkg/config"
    "github.com/turahe/pkg/database"
    "gorm.io/gorm/logger"
)

cfg := config.GetConfig()
db, err := database.New(&cfg.Database, database.Options{
    UseIAM:       true,
    UsePrivateIP: true,
    LogLevel:     logger.Warn,
})
if err != nil {
    log.Fatal(err)
}
defer db.Close()

gormDB := db.DB()
if err := db.Health(ctx); err != nil {
    log.Fatal(err)
}
```

**Legacy API (global):**

```go
err := database.Setup()
if err != nil {
    log.Fatal(err)
}
db := database.GetDB()
```

### Redis

**Standard Redis:**

```go
import "github.com/turahe/pkg/redis"

// Setup Redis connection
err := redis.Setup()
if err != nil {
    log.Fatal(err)
}

// Use Redis
val, err := redis.Get("key")
err = redis.Set("key", "value", 10*time.Second)
```

**Redis Cluster (Google Cloud Memorystore):**

```go
import "github.com/turahe/pkg/redis"

// Setup Redis cluster connection
err := redis.Setup()
if err != nil {
    log.Fatal(err)
}

// Use universal client (works with both standard and cluster)
client := redis.GetUniversalClient()
val, err := client.Get(ctx, "key").Result()

// Or use cluster-specific client
clusterClient := redis.GetRedisCluster()
```

### Middlewares

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/turahe/pkg/middlewares"
)

router := gin.Default()

// Trace first so request context has trace_id/correlation_id for logging
router.Use(middlewares.TraceMiddleware())
router.Use(middlewares.LoggerMiddleware())
router.Use(middlewares.CORS())
router.Use(middlewares.AuthMiddleware()) // Requires JWT setup
router.Use(middlewares.RateLimiter())    // Requires Redis setup

// Setup error handlers
router.NoMethod(middlewares.NoMethodHandler())
router.NoRoute(middlewares.NoRouteHandler())
router.Use(middlewares.RecoveryHandler)
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

// Package-level (no trace/correlation)
logger.Infof("Application started")
logger.Errorf("Error occurred: %v", err)

// With structured fields
logger.Info("user login", logger.Fields{"user_id": id, "ip": ip})

// Request-scoped: use context-bound logger so trace_id/correlation_id appear in JSON
// (use after TraceMiddleware so c.Request.Context() has IDs)
func myHandler(c *gin.Context) {
    log := logger.WithContext(c.Request.Context())
    log.Infof("user %s logged in", userID)
    log.Errorf("operation failed: %v", err)
}
```

## Environment Variables

The package supports configuration via environment variables:

- `DATABASE_DRIVER` - Database driver (mysql, postgres, sqlite, sqlserver, cloudsql-mysql, cloudsql-postgres)
- `DATABASE_HOST` - Database host
- `DATABASE_PORT` - Database port
- `DATABASE_USERNAME` - Database username
- `DATABASE_PASSWORD` - Database password
- `DATABASE_DBNAME` - Database name
- `DATABASE_SSLMODE` - Enable SSL (true/false)
- `DATABASE_LOGMODE` - Enable query logging (true/false)
- `DATABASE_CLOUD_SQL_INSTANCE` - Cloud SQL instance (format: project:region:instance) for cloudsql-mysql/cloudsql-postgres
- `DATABASE_MAX_IDLE_CONNS` - Max idle connections in pool (default: 5; production: 10)
- `DATABASE_MAX_OPEN_CONNS` - Max open connections (default: 10; production: 30)
- `DATABASE_CONN_MAX_LIFETIME` - Max connection lifetime in minutes (default: 1440; production: 30)
- `REDIS_ENABLED` - Enable Redis (true/false)
- `REDIS_HOST` - Redis host
- `REDIS_PORT` - Redis port
- `REDIS_PASSWORD` - Redis password
- `REDIS_DB` - Redis database number
- `REDIS_CLUSTER_MODE` - Enable Redis cluster mode (true/false)
- `REDIS_CLUSTER_NODES` - Comma-separated cluster node addresses
- `RATE_LIMITER_ENABLED` - Enable rate limiter (true/false)
- `RATE_LIMITER_REQUESTS` - Number of requests allowed per window
- `RATE_LIMITER_WINDOW` - Time window in seconds
- `RATE_LIMITER_KEY_BY` - Key strategy: "ip" or "user"
- `RATE_LIMITER_SKIP_PATHS` - Comma-separated paths to skip rate limiting
- `GCS_ENABLED` - Enable GCS (true/false)
- `GCS_BUCKET_NAME` - GCS bucket name
- `SERVER_SECRET` - JWT secret key
- `SERVER_ACCESS_TOKEN_EXPIRY` - Access token expiry in hours
- `SERVER_REFRESH_TOKEN_EXPIRY` - Refresh token expiry in days
- `SERVER_TIMEZONE` - Server timezone (IANA format, e.g., "Asia/Jakarta")

### Example .env

Copy `env.example` to `.env` and adjust for your environment. `config.Setup("")` loads `.env` via [godotenv](https://github.com/joho/godotenv).

```bash
cp env.example .env
```

```env
# Server
SERVER_PORT=8080
SERVER_SECRET=your-jwt-secret-key-change-in-production
SERVER_MODE=debug
SERVER_ACCESS_TOKEN_EXPIRY=1
SERVER_REFRESH_TOKEN_EXPIRY=7

# CORS
CORS_GLOBAL=true
CORS_IPS=

# Database (required: DBNAME, USERNAME, PASSWORD)
# Drivers: mysql, postgres, sqlite, sqlserver, cloudsql-mysql, cloudsql-postgres
DATABASE_DRIVER=mysql
DATABASE_HOST=127.0.0.1
DATABASE_PORT=3306
DATABASE_USERNAME=appuser
DATABASE_PASSWORD=apppassword
DATABASE_DBNAME=appdb
DATABASE_SSLMODE=false
DATABASE_LOGMODE=true
# Connection pool (defaults: 5 idle, 10 open, 1440 min lifetime)
# Production-tuned: 10 idle, 30 open, 30 min lifetime
DATABASE_MAX_IDLE_CONNS=10
DATABASE_MAX_OPEN_CONNS=30
DATABASE_CONN_MAX_LIFETIME=30
# Cloud SQL: project:region:instance (use database.Options for IAM/Private IP)
DATABASE_CLOUD_SQL_INSTANCE=

# Database Site (optional; leave DBNAME empty to disable)
DATABASE_DRIVER_SITE=mysql
DATABASE_HOST_SITE=127.0.0.1
DATABASE_PORT_SITE=3306
DATABASE_USERNAME_SITE=
DATABASE_PASSWORD_SITE=
DATABASE_DBNAME_SITE=
DATABASE_SSLMODE_SITE=false
DATABASE_LOGMODE_SITE=true
DATABASE_CLOUD_SQL_INSTANCE_SITE=

# Redis
REDIS_ENABLED=false
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=1
# Redis Cluster (for Google Cloud Memorystore Redis Cluster)
REDIS_CLUSTER_MODE=false
REDIS_CLUSTER_NODES=10.0.0.1:6379,10.0.0.2:6379,10.0.0.3:6379

# Rate Limiter (requires Redis)
RATE_LIMITER_ENABLED=true
RATE_LIMITER_REQUESTS=100
RATE_LIMITER_WINDOW=60
RATE_LIMITER_KEY_BY=ip
RATE_LIMITER_SKIP_PATHS=/health,/metrics

# GCS
GCS_ENABLED=false
GCS_BUCKET_NAME=
GCS_CREDENTIALS_FILE=
```

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

### Integration tests (Redis, MySQL, Postgres)

To run integration tests against real Redis, MySQL, and Postgres, start the services with Docker Compose:

```bash
docker compose up -d
```

Then run tests with:

```bash
REDIS_ENABLED=true REDIS_HOST=127.0.0.1 REDIS_PORT=6379 \
DATABASE_DRIVER=mysql DATABASE_HOST=127.0.0.1 DATABASE_PORT=3306 \
DATABASE_USERNAME=root DATABASE_PASSWORD=root DATABASE_DBNAME=testdb \
go test ./...
```

Integration tests skip automatically when services are not available. CI uses the same `docker-compose.yml` services (Redis, MySQL, Postgres) via GitHub Actions.

Tests cover: `crypto`, `util`, `config`, `jwt`, `logger`, `redis`, `database`, `gcs`, `types`, `middlewares`, `response`, and `handler`. Packages that require external services (Redis, GCS, MySQL/Postgres) use disabled or in-memory/sqlite config where possible so tests can run without those services.

## License

MIT
